package migrations

import (
	"errors"
	"log"
	"test-constructor/config"
	"test-constructor/internal/domain"

	"gorm.io/gorm"
)

func SeedAdmin(db *gorm.DB) error {
	cfg := config.Load()

	var role domain.Role
	if err := db.Where("code = ?", "admin").First(&role).Error; err != nil {
		return err
	}

	admin := domain.User{
		Email:  cfg.AdminEmail,
		RoleID: role.ID,
		Role:   role,
	}

	err := admin.HashPassword(cfg.AdminPassword)
	if err != nil {
		return err
	}

	var existingAdmin domain.User
	result := db.Where("email = ?", cfg.AdminEmail).First(&existingAdmin)
	if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
		if err := db.Create(&admin).Error; err != nil {
			return err
		}
		log.Printf("Админ создан")
	} else if result.Error == nil {
		log.Printf("Админ уже существует")
	} else {
		return result.Error
	}

	return nil
}
