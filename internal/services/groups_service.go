package services

import (
	"github.com/google/uuid"
	"github.com/prayosha/go-pos-backend/internal/models"
	"gorm.io/gorm"
)

type UserGroupService struct{ db *gorm.DB }

func NewUserGroupService(db *gorm.DB) *UserGroupService { return &UserGroupService{db: db} }

type GroupFilter struct {
	Type     string
	OutletID string
	Name     string
	Page     int
	Limit    int
}

func (s *UserGroupService) List(f GroupFilter) ([]models.UserGroup, int64, error) {
	q := s.db.Model(&models.UserGroup{}).Preload("Members.User").Preload("Outlet")
	if f.Type != "" {
		q = q.Where("type = ?", f.Type)
	}
	if f.OutletID != "" {
		q = q.Where("outlet_id = ?", f.OutletID)
	}
	if f.Name != "" {
		q = q.Where("name ILIKE ?", "%"+f.Name+"%")
	}
	var total int64
	q.Count(&total)
	var groups []models.UserGroup
	offset := (f.Page - 1) * f.Limit
	err := q.Order("name ASC").Offset(offset).Limit(f.Limit).Find(&groups).Error
	return groups, total, err
}

func (s *UserGroupService) GetByID(id uuid.UUID) (*models.UserGroup, error) {
	var g models.UserGroup
	err := s.db.Preload("Members.User").Preload("Outlet").First(&g, "id = ?", id).Error
	return &g, err
}

func (s *UserGroupService) Create(g *models.UserGroup) error {
	return s.db.Create(g).Error
}

func (s *UserGroupService) Update(id uuid.UUID, updates map[string]interface{}) (*models.UserGroup, error) {
	var g models.UserGroup
	if err := s.db.First(&g, "id = ?", id).Error; err != nil {
		return nil, err
	}
	s.db.Model(&g).Updates(updates)
	return &g, nil
}

func (s *UserGroupService) Delete(id uuid.UUID) error {
	return s.db.Delete(&models.UserGroup{}, "id = ?", id).Error
}

func (s *UserGroupService) AddMember(groupID, userID uuid.UUID, billerType models.UserBillerType) (*models.UserGroupMember, error) {
	m := &models.UserGroupMember{
		GroupID:    groupID,
		UserID:     userID,
		BillerType: billerType,
	}
	// upsert – if already exists, just update biller_type
	err := s.db.Where(models.UserGroupMember{GroupID: groupID, UserID: userID}).
		Assign(models.UserGroupMember{BillerType: billerType}).
		FirstOrCreate(m).Error
	if err != nil {
		return nil, err
	}
	s.db.Preload("User").First(m, "id = ?", m.ID)
	return m, nil
}

func (s *UserGroupService) RemoveMember(groupID, userID uuid.UUID) error {
	return s.db.Where("group_id = ? AND user_id = ?", groupID, userID).
		Delete(&models.UserGroupMember{}).Error
}

func (s *UserGroupService) BulkSetMemberStatus(groupID uuid.UUID, isActive bool) error {
	// Fetch user IDs in group
	var memberUserIDs []uuid.UUID
	s.db.Model(&models.UserGroupMember{}).
		Where("group_id = ? AND deleted_at IS NULL", groupID).
		Pluck("user_id", &memberUserIDs)
	if len(memberUserIDs) == 0 {
		return nil
	}
	return s.db.Model(&models.User{}).
		Where("id IN ?", memberUserIDs).
		Update("is_active", isActive).Error
}

// extended UserService methods

type CloudUserFilter struct {
	Name     string
	Email    string
	UserType string // franchise_owner | restaurant_user (maps to role)
	Status   string // active | inactive | all
	Page     int
	Limit    int
}

// returns paginated users for the cloud-access page.
func (s *UserService) ListCloudUsers(f CloudUserFilter) ([]models.User, int64, error) {
	q := s.db.Model(&models.User{}).Preload("OutletAccesses.Outlet")

	if f.Name != "" {
		q = q.Where("name ILIKE ?", "%"+f.Name+"%")
	}
	if f.Email != "" {
		q = q.Where("email ILIKE ?", "%"+f.Email+"%")
	}

	// map UI type to roles
	switch f.UserType {
	case "franchise_owner":
		q = q.Where("role = ?", models.RoleOwner)
	case "restaurant_user":
		q = q.Where("role IN ?", []models.UserRole{models.RoleAdmin, models.RoleBiller})
	}
	switch f.Status {
	case "active":
		q = q.Where("is_active = true")
	case "inactive":
		q = q.Where("is_active = false")
	}

	var total int64
	q.Count(&total)

	offset := (f.Page - 1) * f.Limit
	var users []models.User
	err := q.Order("name ASC").Offset(offset).Limit(f.Limit).Find(&users).Error
	return users, total, err
}

// enables/disables a batch of users at once.
func (s *UserService) BulkSetActiveStatus(ids []uuid.UUID, isActive bool) error {
	return s.db.Model(&models.User{}).
		Where("id IN ?", ids).
		Update("is_active", isActive).Error
}
