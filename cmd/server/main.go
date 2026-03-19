package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prayosha/go-pos-backend/config"
	"github.com/prayosha/go-pos-backend/internal/handlers"
	"github.com/prayosha/go-pos-backend/internal/middleware"
	"github.com/prayosha/go-pos-backend/internal/models"
	"github.com/prayosha/go-pos-backend/internal/services"
	"github.com/prayosha/go-pos-backend/pkg/database"
	"github.com/prayosha/go-pos-backend/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()
	logger.Init(cfg.App.Env)
	defer logger.Sync()
	logger.Infof("Starting %s [%s]", cfg.App.Name, cfg.App.Env)

	// Databases
	db, err := database.ConnectPostgres(cfg)
	if err != nil {
		logger.Fatal("PostgreSQL connection failed", zap.Error(err))
	}
	redisClient, err := database.ConnectRedis(cfg)
	if err != nil {
		logger.Warn("Redis unavailable — sessions stored in-DB only", zap.Error(err))
	}

	// Auto-migrate
	if err := db.AutoMigrate(
		&models.User{}, &models.RefreshToken{},
		&models.UserGroup{}, &models.UserGroupMember{},
		&models.Franchise{}, &models.Outlet{}, &models.OutletAccess{},
		&models.Zone{}, &models.Table{},
		&models.Category{}, &models.MenuItem{},
		&models.KOT{}, &models.KOTItem{},
		&models.Order{}, &models.OrderItem{}, &models.Payment{},
		&models.PendingPurchase{},
		&models.StoreStatusSnapshot{},
		&models.MenuTriggerLog{}, &models.OnlineStoreLog{}, &models.OnlineItemLog{},
		&models.ThirdPartyConfig{},
		&models.Notification{},
		&models.Upload{},
		&models.DeviceToken{},
	); err != nil {
		logger.Fatal("Auto-migration failed", zap.Error(err))
	}
	logger.Info("Database migrations complete")

	// Services
	authSvc := services.NewAuthService(db, redisClient, cfg)
	dashSvc := services.NewDashboardService(db)
	outletSvc := services.NewOutletService(db)
	orderSvc := services.NewOrderService(db)
	menuSvc := services.NewMenuService(db)
	reportsSvc := services.NewReportsService(db)
	inventorySvc := services.NewInventoryService(db)
	notifSvc := services.NewNotificationService(db)
	tpSvc := services.NewThirdPartyService(db)
	logsSvc := services.NewLogsService(db)
	franchiseSvc := services.NewFranchiseService(db)
	userSvc := services.NewUserService(db)
	groupSvc := services.NewUserGroupService(db)
	storeStatusSvc := services.NewStoreStatusService(db)
	kotSvc := services.NewKOTService(db)
	uploadDir := cfg.App.UploadDir
	baseURL := cfg.App.BaseURL
	uploadSvc := services.NewUploadService(db, uploadDir, baseURL)
	fcmSvc := services.NewFCMService(db, cfg.FCM.ServerKey, cfg.FCM.ProjectID)
	exportSvc := services.NewExportService(reportsSvc)

	// Background cron
	cron := services.NewCronService(menuSvc, authSvc, storeStatusSvc)
	cron.Start()

	// Handlers
	authH := handlers.NewAuthHandler(authSvc, cfg)
	dashH := handlers.NewDashboardHandler(dashSvc)
	outletH := handlers.NewOutletHandler(outletSvc)
	orderH := handlers.NewOrderHandler(orderSvc)
	menuH := handlers.NewMenuHandler(menuSvc)
	reportsH := handlers.NewReportsHandler(reportsSvc)
	inventoryH := handlers.NewInventoryHandler(inventorySvc)
	notifH := handlers.NewNotificationHandler(notifSvc)
	tpH := handlers.NewThirdPartyHandler(tpSvc)
	logsH := handlers.NewLogsHandler(logsSvc)
	franchiseH := handlers.NewFranchiseHandler(franchiseSvc)
	userH := handlers.NewUserHandler(userSvc)
	groupH := handlers.NewUserGroupHandler(groupSvc)
	cloudH := handlers.NewCloudAccessHandler(userSvc)
	outletTypeH := handlers.NewOutletTypeHandler(outletSvc)
	storeStatusH := handlers.NewStoreStatusHandler(storeStatusSvc)
	kotH := handlers.NewKOTHandler(kotSvc)
	uploadH := handlers.NewUploadHandler(uploadSvc)
	fcmH := handlers.NewFCMHandler(fcmSvc)
	exportH := handlers.NewExportHandler(exportSvc)

	// Gin engine
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(zapRequestLogger(logger.Get()))
	r.Use(middleware.CORS(cfg.CORS.AllowedOrigins))

	// Static uploads
	r.Static("/static/uploads", uploadDir)

	// Health probes
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": cfg.App.Name, "ts": time.Now().UTC()})
	})
	r.GET("/ready", func(c *gin.Context) {
		sqlDB, _ := db.DB()
		if err := sqlDB.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	// API v1
	v1 := r.Group("/api/v1")

	// Auth (public)
	auth := v1.Group("/auth", middleware.RateLimitAuth())
	{
		auth.POST("/login", authH.Login)
		auth.POST("/register", authH.Register)
		auth.POST("/refresh", authH.RefreshToken)
		auth.POST("/logout", authH.Logout)
		auth.POST("/google", authH.GoogleAuth)
	}

	// Protected
	p := v1.Group("", middleware.AuthRequired(cfg), middleware.RateLimitAPI())

	// Auth (protected)
	p.GET("/auth/me", authH.Me)
	p.PUT("/auth/change-password", authH.ChangePassword)

	// Uploads
	p.POST("/uploads", middleware.RateLimitUpload(), uploadH.Upload)

	// Push Notifications (FCM device token registration)
	push := p.Group("/push")
	{
		push.POST("/register", fcmH.RegisterToken)
		push.DELETE("/register", fcmH.RemoveToken)
	}

	// Dashboard
	dash := p.Group("/dashboard")
	{
		dash.GET("/stats", dashH.GetStats)
		dash.GET("/outlet-stats", dashH.GetOutletStats)
		dash.GET("/orders-chart", dashH.GetOrdersChart)
		dash.GET("/orders-chart-by-platform", reportsH.GetChartByPlatform)
		dash.GET("/summary", dashH.GetSummary)
	}

	// Outlets
	outlets := p.Group("/outlets")
	{
		outlets.GET("", outletH.List)
		outlets.GET("/types", outletH.GetTypes)
		outlets.POST("", middleware.RequireRole(models.RoleAdmin, models.RoleOwner), outletH.Create)
		outlets.GET("/:id", outletH.Get)
		outlets.PUT("/:id", middleware.RequireRole(models.RoleAdmin, models.RoleOwner), outletH.Update)
		outlets.DELETE("/:id", middleware.RequireRole(models.RoleAdmin), outletH.Delete)
		outlets.PATCH("/:id/lock", middleware.RequireRole(models.RoleAdmin, models.RoleOwner), outletH.ToggleLock)
		outlets.GET("/:id/zones", outletH.GetZones)
		outlets.POST("/:id/zones", middleware.RequireRole(models.RoleAdmin, models.RoleOwner), outletH.CreateZone)
		outlets.PUT("/zones/:zone_id", middleware.RequireRole(models.RoleAdmin, models.RoleOwner), outletH.UpdateZone)
		outlets.DELETE("/zones/:zone_id", middleware.RequireRole(models.RoleAdmin), outletH.DeleteZone)
		outlets.POST("/zones/:zone_id/tables", middleware.RequireRole(models.RoleAdmin, models.RoleOwner), outletH.CreateTable)
		outlets.PUT("/tables/:table_id", middleware.RequireRole(models.RoleAdmin, models.RoleOwner), outletH.UpdateTable)
		outlets.DELETE("/tables/:table_id", middleware.RequireRole(models.RoleAdmin), outletH.DeleteTable)
	}

	// Outlet types (franchise classification)
	outletTypes := p.Group("/outlet-types", middleware.RequireRole(models.RoleAdmin, models.RoleOwner))
	{
		outletTypes.GET("", outletTypeH.List)
		outletTypes.PUT("/:id", outletTypeH.Update)
	}

	// Orders
	orders := p.Group("/orders")
	{
		orders.GET("", orderH.List)
		orders.POST("", orderH.Create)
		orders.GET("/running", orderH.GetRunningOrders)
		orders.GET("/online", orderH.GetOnlineOrders)
		orders.GET("/platforms", orderH.GetPlatforms)
		orders.GET("/:id", orderH.Get)
		orders.PATCH("/:id/status", orderH.UpdateStatus)
		orders.PATCH("/:id/cancel", orderH.Cancel)
		orders.POST("/:id/print", orderH.MarkPrinted)
	}

	// KOT
	kots := p.Group("/kots")
	{
		kots.GET("", kotH.List)
		kots.POST("", kotH.Create)
		kots.PATCH("/:id/status", kotH.UpdateStatus)
		kots.POST("/:id/print", kotH.MarkPrinted)
	}

	// Menu
	menu := p.Group("/menu")
	{
		menu.GET("/categories", menuH.GetCategories)
		menu.POST("/categories", middleware.RequireRole(models.RoleAdmin, models.RoleOwner), menuH.CreateCategory)
		menu.PUT("/categories/:id", middleware.RequireRole(models.RoleAdmin, models.RoleOwner), menuH.UpdateCategory)
		menu.DELETE("/categories/:id", middleware.RequireRole(models.RoleAdmin), menuH.DeleteCategory)
		menu.GET("/items", menuH.GetItems)
		menu.POST("/items", middleware.RequireRole(models.RoleAdmin, models.RoleOwner), menuH.CreateItem)
		menu.PUT("/items/:id", middleware.RequireRole(models.RoleAdmin, models.RoleOwner), menuH.UpdateItem)
		menu.DELETE("/items/:id", middleware.RequireRole(models.RoleAdmin), menuH.DeleteItem)
		menu.GET("/out-of-stock", menuH.GetOutOfStockItems)
		menu.PATCH("/items/:id/availability", menuH.ToggleAvailability)
		menu.PATCH("/items/:id/online", middleware.RequireRole(models.RoleAdmin, models.RoleOwner), menuH.ToggleOnlineStatus)
	}

	// Reports
	reports := p.Group("/reports")
	{
		reports.GET("/list", reportsH.GetReportsList)
		reports.GET("/sales", reportsH.GetSalesReport)
		reports.GET("/item-wise", reportsH.GetItemWiseReport)
		reports.GET("/category-wise", reportsH.GetCategoryWiseReport)
		reports.GET("/invoices", reportsH.GetInvoiceReport)
		reports.GET("/cancelled-orders", reportsH.GetCancelledOrderReport)
		reports.GET("/discounts", reportsH.GetDiscountReport)
		reports.GET("/hourly", reportsH.GetHourlyReport)
		reports.GET("/pax-biller", reportsH.GetPaxBillerReport)
		reports.GET("/day-wise", reportsH.GetDayWiseReport)
		reports.GET("/orders-master", reportsH.GetOrderMasterReport)
		reports.GET("/online-orders", reportsH.GetOnlineOrderReport)
		reports.GET("/chart-by-platform", reportsH.GetChartByPlatform)
	}

	// Purchases / Inventory
	purchases := p.Group("/purchases")
	{
		purchases.GET("/pending", inventoryH.GetPendingPurchases)
		purchases.POST("", inventoryH.CreatePurchase)
		purchases.PATCH("/:id/status", inventoryH.UpdatePurchaseStatus)
	}

	// Notifications
	notifs := p.Group("/notifications")
	{
		notifs.GET("", notifH.List)
		notifs.PATCH("/:id/read", notifH.MarkRead)
		notifs.PATCH("/read-all", notifH.MarkAllRead)
	}

	// Third-party configs
	tp := p.Group("/thirdparty", middleware.RequireRole(models.RoleAdmin, models.RoleOwner))
	{
		tp.GET("", tpH.List)
		tp.PUT("/:id", tpH.Update)
	}

	// Logs
	logs := p.Group("/logs")
	{
		logs.GET("/menu-triggers", logsH.GetMenuTriggerLogs)
		logs.GET("/online-store", logsH.GetOnlineStoreLogs)
		logs.GET("/online-items", logsH.GetOnlineItemLogs)
	}

	// Franchises
	franchise := p.Group("/franchises", middleware.RequireRole(models.RoleAdmin, models.RoleOwner))
	{
		franchise.GET("", franchiseH.List)
		franchise.POST("", franchiseH.Create)
		franchise.POST("/:id/outlets", franchiseH.AssignOutlet)
	}

	// Users
	users := p.Group("/users")
	{
		users.GET("/billers", middleware.RequireRole(models.RoleAdmin, models.RoleOwner), userH.GetBillers)
		users.GET("/admins", middleware.RequireRole(models.RoleAdmin), userH.GetAdmins)
		users.POST("/invite", middleware.RequireRole(models.RoleAdmin, models.RoleOwner), userH.InviteUser)
		users.PUT("/:id", middleware.RequireRole(models.RoleAdmin), userH.Update)
		users.DELETE("/:id", middleware.RequireRole(models.RoleAdmin), userH.Delete)
	}

	// User Groups (Admin Groups + Biller Groups)
	groups := p.Group("/groups", middleware.RequireRole(models.RoleAdmin, models.RoleOwner))
	{
		groups.GET("", groupH.List)
		groups.POST("", groupH.Create)
		groups.GET("/:id", groupH.Get)
		groups.PUT("/:id", groupH.Update)
		groups.DELETE("/:id", middleware.RequireRole(models.RoleAdmin), groupH.Delete)
		groups.POST("/:id/members", groupH.AddMember)
		groups.DELETE("/:id/members/:user_id", groupH.RemoveMember)
		groups.POST("/:id/bulk-status", groupH.BulkSetStatus)
	}

	// Cloud Access
	cloud := p.Group("/cloud-access", middleware.RequireRole(models.RoleAdmin, models.RoleOwner))
	{
		cloud.GET("", cloudH.List)
		cloud.PATCH("/bulk-status", cloudH.BulkSetStatus)
	}

	// Store Status Tracking
	storeStatus := p.Group("/store-status")
	{
		storeStatus.GET("", storeStatusH.List)
		storeStatus.POST("/refresh", storeStatusH.Refresh)
		storeStatus.GET("/history", storeStatusH.History)
	}

	// Export (CSV downloads)
	exports := p.Group("/export")
	{
		exports.GET("/sales", exportH.SalesReport)
		exports.GET("/item-wise", exportH.ItemWise)
		exports.GET("/category-wise", exportH.CategoryWise)
		exports.GET("/invoices", exportH.Invoices)
		exports.GET("/orders-master", exportH.OrdersMaster)
		exports.GET("/cancelled-orders", exportH.CancelledOrders)
		exports.GET("/discounts", exportH.Discounts)
		exports.GET("/hourly", exportH.Hourly)
		exports.GET("/day-wise", exportH.DayWise)
		exports.GET("/pending-purchases", exportH.PendingPurchases)
		exports.GET("/store-status", exportH.StoreStatus)
	}

	// Start
	addr := fmt.Sprintf(":%s", cfg.App.Port)
	logger.Infof("Listening on http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		logger.Fatal("Server failed to start", zap.Error(err))
	}
}

func zapRequestLogger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("ip", c.ClientIP()),
		}
		switch {
		case c.Writer.Status() >= 500:
			log.Error("server error", fields...)
		case c.Writer.Status() >= 400:
			log.Warn("client error", fields...)
		default:
			log.Info("request", fields...)
		}
	}
}
