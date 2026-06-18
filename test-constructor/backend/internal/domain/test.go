package domain

import (
	"gorm.io/gorm"
)

type Test struct {
	gorm.Model
	CreatorID   uint   `json:"creator_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	// Связи
	Questions []Question
	User      User `gorm:"foreignKey:CreatorID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"`
}
