package repository

import (
	"test-constructor/internal/domain"

	"gorm.io/gorm"
)

type ExtraThresholdRepository interface {
	Create(threshold *domain.ExtraThreshold) error
	CreateWithTx(tx *gorm.DB, threshold *domain.ExtraThreshold) error
	FindByConfigID(configID uint) ([]domain.ExtraThreshold, error)
	FindByExtraConfigID(extraConfigID uint) (*domain.ExtraThreshold, error)
	FindReplacementsForConfigID(configID uint) ([]domain.ExtraThreshold, error)
	DeleteByConfigID(configID uint) error
	DeleteByConfigIDWithTx(tx *gorm.DB, configID uint) error
}

type extraThresholdRepository struct {
	db *gorm.DB
}

func NewExtraThresholdRepository(db *gorm.DB) ExtraThresholdRepository {
	return &extraThresholdRepository{db: db}
}

func (r *extraThresholdRepository) Create(threshold *domain.ExtraThreshold) error {
	return r.db.Create(threshold).Error
}

func (r *extraThresholdRepository) CreateWithTx(tx *gorm.DB, threshold *domain.ExtraThreshold) error {
	return tx.Create(threshold).Error
}

func (r *extraThresholdRepository) FindByConfigID(configID uint) ([]domain.ExtraThreshold, error) {
	var thresholds []domain.ExtraThreshold
	err := r.db.Where("config_id = ?", configID).
		Preload("ExtraConfig").
		Find(&thresholds).Error
	return thresholds, err
}

func (r *extraThresholdRepository) FindByExtraConfigID(extraConfigID uint) (*domain.ExtraThreshold, error) {
	var threshold domain.ExtraThreshold
	err := r.db.Where("extra_config_id = ?", extraConfigID).
		First(&threshold).Error
	if err != nil {
		return nil, err
	}
	return &threshold, nil
}

func (r *extraThresholdRepository) FindReplacementsForConfigID(configID uint) ([]domain.ExtraThreshold, error) {
	var thresholds []domain.ExtraThreshold
	err := r.db.Where("config_id = ?", configID).
		Preload("ExtraConfig.Test").
		Order("threshold DESC").
		Find(&thresholds).Error
	return thresholds, err
}

func (r *extraThresholdRepository) DeleteByConfigID(configID uint) error {
	return r.db.Where("config_id = ?", configID).Delete(&domain.ExtraThreshold{}).Error
}

func (r *extraThresholdRepository) DeleteByConfigIDWithTx(tx *gorm.DB, configID uint) error {
	return tx.Where("config_id = ?", configID).Delete(&domain.ExtraThreshold{}).Error
}
