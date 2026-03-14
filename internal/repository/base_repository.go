package repository

import (
	"gorm.io/gorm"
)

// Repository is the generic base that every domain repo embeds.
// T is the GORM model type (e.g. models.User, models.Order).
type Repository[T any] struct {
	db *gorm.DB
}

// creates a repository for model T.
func NewRepository[T any](db *gorm.DB) *Repository[T] {
	return &Repository[T]{db: db}
}

// DB returns the underlying *gorm.DB so sub-repos can extend queries.
func (r *Repository[T]) DB() *gorm.DB {
	return r.db
}

// inserts a new record and returns any error.
func (r *Repository[T]) Create(entity *T) error {
	return r.db.Create(entity).Error
}

// reates-or-updates (upserts) a record.
func (r *Repository[T]) Save(entity *T) error {
	return r.db.Save(entity).Error
}

// fetches a single record by primary key.
func (r *Repository[T]) FindByID(id interface{}, dest *T) error {
	return r.db.First(dest, "id = ?", id).Error
}

// returns every non-deleted record.
func (r *Repository[T]) FindAll(dest *[]T) error {
	return r.db.Find(dest).Error
}

// applies a map of column→value changes.
func (r *Repository[T]) Update(entity *T, updates map[string]interface{}) error {
	return r.db.Model(entity).Updates(updates).Error
}

// soft-deletes (sets deleted_at) the record.
func (r *Repository[T]) Delete(entity *T) error {
	return r.db.Delete(entity).Error
}

// soft-deletes by primary key.
func (r *Repository[T]) DeleteByID(id interface{}) error {
	var zero T
	return r.db.Delete(&zero, "id = ?", id).Error
}

// returns the number of rows matching the given condition.
func (r *Repository[T]) Count(condition string, args ...interface{}) (int64, error) {
	var zero T
	var count int64
	err := r.db.Model(&zero).Where(condition, args...).Count(&count).Error
	return count, err
}

// returns true when at least one row matches.
func (r *Repository[T]) Exists(condition string, args ...interface{}) (bool, error) {
	n, err := r.Count(condition, args...)
	return n > 0, err
}

// Paginate is a convenience helper that adds LIMIT/OFFSET to a query.
func Paginate(page, limit int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		if limit <= 0 || limit > 200 {
			limit = 20
		}
		return db.Offset((page - 1) * limit).Limit(limit)
	}
}
