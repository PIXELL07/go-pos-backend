package repository

import (
	"github.com/google/uuid"
	"github.com/prayosha/go-pos-backend/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	*Repository[models.User]
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{Repository: NewRepository[models.User](db)}
}

// returns the user with the given e-mail address.
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ? AND deleted_at IS NULL", email).First(&user).Error
	return &user, err
}

// returns the user with the given mobile number.
func (r *UserRepository) FindByMobile(mobile string) (*models.User, error) {
	var user models.User
	err := r.db.Where("mobile = ? AND deleted_at IS NULL", mobile).First(&user).Error
	return &user, err
}

// returns the user matching either credential.
func (r *UserRepository) FindByEmailOrMobile(value string) (*models.User, error) {
	var user models.User
	err := r.db.Where("(email = ? OR mobile = ?) AND deleted_at IS NULL", value, value).
		First(&user).Error
	return &user, err
}

// returns the user linked to the given Google sub claim.
func (r *UserRepository) FindByGoogleID(googleID string) (*models.User, error) {
	var user models.User
	err := r.db.Where("google_id = ? AND deleted_at IS NULL", googleID).First(&user).Error
	return &user, err
}

// FindByRole returns all active users with the given role,
// optionally scoped to a specific outlet.
func (r *UserRepository) FindByRole(role models.UserRole, outletID string) ([]models.User, error) {
	query := r.db.Where("role = ? AND is_active = true AND deleted_at IS NULL", role)
	if outletID != "" {
		query = query.
			Joins("JOIN outlet_accesses oa ON oa.user_id = users.id").
			Where("oa.outlet_id = ? AND oa.deleted_at IS NULL", outletID)
	}
	var users []models.User
	return users, query.Find(&users).Error
}

// returns true when an active user with that e-mail already exists.
func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	return r.Exists("email = ? AND deleted_at IS NULL", email)
}

// returns true when an active user with that mobile already exists.
func (r *UserRepository) ExistsByMobile(mobile string) (bool, error) {
	return r.Exists("mobile = ? AND deleted_at IS NULL", mobile)
}

// UpdateLastLogin sets last_login_at to now.
func (r *UserRepository) UpdateLastLogin(id uuid.UUID) error {
	return r.db.Model(&models.User{}).
		Where("id = ?", id).
		Update("last_login_at", gorm.Expr("NOW()")).Error
}

// sets the hashed password for the given user.
func (r *UserRepository) UpdatePassword(id uuid.UUID, hash string) error {
	return r.db.Model(&models.User{}).
		Where("id = ?", id).
		Update("password_hash", hash).Error
}

// creates or updates outlet access for a user.
func (r *UserRepository) AssignOutletAccess(access *models.OutletAccess) error {
	return r.db.
		Where(models.OutletAccess{UserID: access.UserID, OutletID: access.OutletID}).
		Assign(models.OutletAccess{Role: access.Role}).
		FirstOrCreate(access).Error
}
