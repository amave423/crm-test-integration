package migrations

import (
	"errors"
	"log"
	"test-constructor/internal/domain"

	"gorm.io/gorm"
)

func SeedRoles(db *gorm.DB) error {
	predefinedRoles := []domain.Role{
		{Name: "Администратор", Code: "admin"},
		{Name: "Организатор", Code: "manager"},
		{Name: "Стажёр", Code: "intern"},
	}

	for _, role := range predefinedRoles {
		var existingRole domain.Role
		result := db.Where("code = ?", role.Code).First(&existingRole)

		if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
			if err := db.Create(&role).Error; err != nil {
				return err
			}
			log.Printf("Роль создана: %s", role.Code)
		} else if result.Error == nil {
			log.Printf("Роль уже существует: %s", role.Code)
		} else {
			return result.Error
		}
	}

	return nil
}
