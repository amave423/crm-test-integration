package domain

type Role struct {
	ID   uint   `gorm:"primarykey"`
	Name string `gorm:"uniqueIndex;not null"`
	Code string `gorm:"uniqueIndex;not null"`
}
