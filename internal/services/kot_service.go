package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/prayosha/go-pos-backend/internal/models"
	"gorm.io/gorm"
)

type KOTService struct{ db *gorm.DB }

func NewKOTService(db *gorm.DB) *KOTService { return &KOTService{db: db} }

type CreateKOTRequest struct {
	OrderID  string         `json:"order_id" binding:"required"`
	OutletID string         `json:"outlet_id" binding:"required"`
	Notes    string         `json:"notes"`
	Items    []KOTItemInput `json:"items" binding:"required,min=1"`
}

type KOTItemInput struct {
	MenuItemID string `json:"menu_item_id"`
	Name       string `json:"name" binding:"required"`
	Quantity   int    `json:"quantity" binding:"required,min=1"`
	Notes      string `json:"notes"`
}

func (s *KOTService) List(filter map[string]string) ([]models.KOT, error) {
	q := s.db.Preload("Items").Model(&models.KOT{})
	if v := filter["order_id"]; v != "" {
		q = q.Where("order_id = ?", v)
	}
	if v := filter["outlet_id"]; v != "" {
		q = q.Where("outlet_id = ?", v)
	}
	if v := filter["status"]; v != "" {
		q = q.Where("status = ?", v)
	}
	var kots []models.KOT
	return kots, q.Order("created_at DESC").Find(&kots).Error
}

func (s *KOTService) Create(req *CreateKOTRequest, createdBy uuid.UUID) (*models.KOT, error) {
	orderID, err := uuid.Parse(req.OrderID)
	if err != nil {
		return nil, fmt.Errorf("invalid order_id")
	}
	outletID, err := uuid.Parse(req.OutletID)
	if err != nil {
		return nil, fmt.Errorf("invalid outlet_id")
	}

	kotNumber, err := s.generateKOTNumber(outletID)
	if err != nil {
		return nil, err
	}

	kot := &models.KOT{
		OrderID:   orderID,
		OutletID:  outletID,
		KOTNumber: kotNumber,
		Status:    models.KOTStatusPending,
		Notes:     req.Notes,
	}

	var items []models.KOTItem
	for _, it := range req.Items {
		item := models.KOTItem{
			Name:     it.Name,
			Quantity: it.Quantity,
			Notes:    it.Notes,
		}
		if it.MenuItemID != "" {
			if id, err := uuid.Parse(it.MenuItemID); err == nil {
				item.MenuItemID = id
			}
		}
		items = append(items, item)
	}
	kot.Items = items

	if err := s.db.Create(kot).Error; err != nil {
		return nil, err
	}
	s.db.Preload("Items").First(kot, "id = ?", kot.ID)
	return kot, nil
}

func (s *KOTService) UpdateStatus(id uuid.UUID, status models.KOTStatus) (*models.KOT, error) {
	var kot models.KOT
	if err := s.db.First(&kot, "id = ?", id).Error; err != nil {
		return nil, err
	}
	s.db.Model(&kot).Update("status", status)
	s.db.Preload("Items").First(&kot, "id = ?", id)
	return &kot, nil
}

func (s *KOTService) MarkPrinted(id uuid.UUID) (*models.KOT, error) {
	now := time.Now()
	var kot models.KOT
	if err := s.db.First(&kot, "id = ?", id).Error; err != nil {
		return nil, err
	}
	s.db.Model(&kot).Update("printed_at", now)
	return &kot, nil
}

func (s *KOTService) generateKOTNumber(outletID uuid.UUID) (string, error) {
	var count int64
	today := time.Now().Format("20060102")
	s.db.Model(&models.KOT{}).
		Where("outlet_id = ? AND DATE(created_at) = CURRENT_DATE", outletID).
		Count(&count)
	return fmt.Sprintf("KOT-%s-%04d", today, count+1), nil
}
