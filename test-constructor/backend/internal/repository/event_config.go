package repository

import (
	"fmt"
	"test-constructor/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventConfigRepository interface {
	Create(config *domain.EventConfig) error
	CreateWithTx(tx *gorm.DB, config *domain.EventConfig) error
	FindByID(id uint) (*domain.EventConfig, error)
	FindByIDFull(id uint) (*domain.EventConfig, error)
	FindByEventAndSpecialization(eventID, specializationID uint, isExtra bool) ([]domain.EventConfig, error)
	FindByEventID(eventID uint) ([]domain.EventConfig, error)
	FindMainConfigsByEventAndSpec(eventID, specializationID uint) ([]domain.EventConfig, error)
	FindCommonConfigsByEventID(eventID uint) ([]domain.EventConfig, error)
	FindByTestLink(link string) (*domain.EventConfig, error)
	Update(config *domain.EventConfig) error
	UpdateWithTx(tx *gorm.DB, config *domain.EventConfig, updates domain.EventConfig) error
	Delete(id uint) error
	FindByEventSpecializationAndTest(tx *gorm.DB, eventID, specializationID, testID uint) (*domain.EventConfig, error)
	UpdateFieldsWithTx(tx *gorm.DB, configID uint, updates domain.EventConfig) error
	UpdateTestLinkWithTx(tx *gorm.DB, configID uint, testLink uuid.UUID) error
	FindByEventIDWithExtraRules(eventID, specializationID uint) ([]domain.EventConfig, error)
	FindByEventAndSpecializationAll(eventID, specializationID uint) ([]domain.EventConfig, error)
}

type eventConfigRepository struct {
	db *gorm.DB
}

func NewEventConfigRepository(db *gorm.DB) EventConfigRepository {
	return &eventConfigRepository{db: db}
}

func (r *eventConfigRepository) Create(config *domain.EventConfig) error {
	return r.db.Create(config).Error
}

func (r *eventConfigRepository) CreateWithTx(tx *gorm.DB, config *domain.EventConfig) error {
	return tx.Create(config).Error
}

func (r *eventConfigRepository) FindByID(id uint) (*domain.EventConfig, error) {
	var config domain.EventConfig
	err := r.db.
		Preload("ExtraThreshold.ExtraConfig").
		Preload("Test").
		Preload("User").
		First(&config, id).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *eventConfigRepository) FindByIDFull(id uint) (*domain.EventConfig, error) {
	var config domain.EventConfig
	err := r.db.
		Preload("ExtraThreshold.ExtraConfig.Test").
		Preload("Test.Questions").
		Preload("User").
		First(&config, id).Error
	if err != nil {
		return nil, fmt.Errorf("конфигурация не найдена: %w", err)
	}
	return &config, nil
}

func (r *eventConfigRepository) FindByEventAndSpecialization(eventID, specializationID uint, isExtra bool) ([]domain.EventConfig, error) {
	var configs []domain.EventConfig
	err := r.db.
		Where("event_id = ? AND specialization_id = ? AND is_extra = ?", eventID, specializationID, isExtra).
		Preload("Test").
		Preload("ExtraThreshold.ExtraConfig").
		Find(&configs).Error
	return configs, err
}

func (r *eventConfigRepository) FindByEventID(eventID uint) ([]domain.EventConfig, error) {
	var configs []domain.EventConfig
	err := r.db.Preload("ExtraThreshold").Where("event_id = ?", eventID).Order("config_id").Find(&configs).Error
	return configs, err
}

func (r *eventConfigRepository) FindMainConfigsByEventAndSpec(eventID, specializationID uint) ([]domain.EventConfig, error) {
	var configs []domain.EventConfig
	err := r.db.
		Where("event_id = ? AND is_extra = ? AND (specialization_id = ? OR specialization_id = 0)",
			eventID, false, specializationID).
		Preload("Test").
		Preload("ExtraThreshold.ExtraConfig").
		Find(&configs).Error
	return configs, err
}

func (r *eventConfigRepository) FindCommonConfigsByEventID(eventID uint) ([]domain.EventConfig, error) {
	var configs []domain.EventConfig
	err := r.db.
		Where("event_id = ? AND specialization_id = 0 AND is_extra = ?", eventID, false).
		Preload("Test").
		Preload("ExtraThreshold.ExtraConfig").
		Find(&configs).Error
	return configs, err
}

func (r *eventConfigRepository) FindByTestLink(link string) (*domain.EventConfig, error) {
	var config domain.EventConfig
	err := r.db.
		Where("test_link = ?", link).
		Preload("Test.Questions").
		Preload("ExtraThreshold").
		First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *eventConfigRepository) Update(config *domain.EventConfig) error {
	return r.db.Save(config).Error
}

func (r *eventConfigRepository) UpdateWithTx(tx *gorm.DB, config *domain.EventConfig, updates domain.EventConfig) error {
	return tx.Model(config).Updates(updates).Error
}

func (r *eventConfigRepository) UpdateTestLinkWithTx(tx *gorm.DB, configID uint, testLink uuid.UUID) error {
	return tx.Model(&domain.EventConfig{}).Where("config_id = ?", configID).Update("test_link", testLink).Error
}

func (r *eventConfigRepository) UpdateFieldsWithTx(tx *gorm.DB, configID uint, updates domain.EventConfig) error {
	return tx.Model(&domain.EventConfig{}).Where("config_id = ?", configID).Updates(updates).Error
}

func (r *eventConfigRepository) Delete(id uint) error {
	return r.db.Delete(&domain.EventConfig{}, id).Error
}

func (r *eventConfigRepository) FindByEventSpecializationAndTest(tx *gorm.DB, eventID, specializationID, testID uint) (*domain.EventConfig, error) {
	var config domain.EventConfig
	err := tx.Where("event_id = ? AND specialization_id = ? AND test_id = ?", eventID, specializationID, testID).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *eventConfigRepository) FindByEventIDWithExtraRules(eventID, specializationID uint) ([]domain.EventConfig, error) {
	var configs []domain.EventConfig
	query := r.db.
		Preload("Test").
		Preload("Test.Questions").
		Preload("ExtraThreshold").
		Where("event_id = ?", eventID).
		Order("config_id ASC")
	if specializationID > 0 {
		query = query.Where("specialization_id = ?", specializationID)
	}
	err := query.Find(&configs).Error
	return configs, err
}

func (r *eventConfigRepository) FindByEventAndSpecializationAll(eventID, specializationID uint) ([]domain.EventConfig, error) {
	var configs []domain.EventConfig
	query := r.db.
		Preload("ExtraThreshold").
		Where("event_id = ?", eventID)
	if specializationID > 0 {
		query = query.Where("specialization_id = ?", specializationID)
	}
	err := query.Find(&configs).Error
	return configs, err
}
