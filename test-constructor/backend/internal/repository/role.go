package repository

import (
	"test-constructor/internal/domain"

	"gorm.io/gorm"
)

type RoleRepository interface {
	Create(role *domain.Role) error
	FindByCode(code string) (*domain.Role, error)
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(role *domain.Role) error {
	return r.db.Create(role).Error
}

func (r *roleRepository) FindByCode(code string) (*domain.Role, error) {
	var role domain.Role
	err := r.db.Where("code = ?", code).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}
