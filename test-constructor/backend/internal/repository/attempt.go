package repository

import (
	"test-constructor/internal/domain"

	"gorm.io/gorm"
)

type AttemptRepository interface {
	Create(attempt *domain.Attempt) error
	CreateWithTx(tx *gorm.DB, attempt *domain.Attempt) error
	FindByID(id uint) (*domain.Attempt, error)
	FindActiveByUser(userID uint) (*domain.Attempt, error)
	FindByUserAndConfig(userID, configID uint) (*domain.Attempt, error)
	FindByConfigIDsAndUser(configIDs []uint, userID uint) ([]domain.Attempt, error)
	FindByUserAndConfigIDs(userID uint, configIDs []uint, applicationID uint) ([]domain.Attempt, error)
	FindCompletedByUserAndConfigIDs(userID, applicationID uint, configIDs []uint) ([]domain.Attempt, error)
	FindByConfigID(configID uint) ([]domain.Attempt, error)
	FindCompletedByEventAndSpec(userID, eventID, specializationID uint) ([]domain.Attempt, error)
	Update(attempt *domain.Attempt) error
	UpdateWithTx(tx *gorm.DB, attempt *domain.Attempt) error
	Delete(id uint) error
}

type attemptRepository struct {
	db *gorm.DB
}

func NewAttemptRepository(db *gorm.DB) AttemptRepository {
	return &attemptRepository{db: db}
}

func (r *attemptRepository) Create(attempt *domain.Attempt) error {
	return r.db.Create(attempt).Error
}

func (r *attemptRepository) CreateWithTx(tx *gorm.DB, attempt *domain.Attempt) error {
	return tx.Create(attempt).Error
}

func (r *attemptRepository) FindByID(id uint) (*domain.Attempt, error) {
	var attempt domain.Attempt
	err := r.db.
		Preload("EventConfig").
		Preload("EventConfig.Test").
		Preload("Answers").
		Preload("Answers.Question").
		First(&attempt, id).Error
	if err != nil {
		return nil, err
	}
	return &attempt, nil
}

func (r *attemptRepository) FindActiveByUser(userID uint) (*domain.Attempt, error) {
	var attempt domain.Attempt
	err := r.db.
		Preload("EventConfig").
		Preload("EventConfig.ExtraThreshold").
		Preload("EventConfig.Test").
		Preload("EventConfig.Test.Questions").
		Where("intern_id = ? AND end_time IS NULL", userID).
		First(&attempt).Error
	if err != nil {
		return nil, err
	}
	return &attempt, nil
}

func (r *attemptRepository) FindByUserAndConfig(userID, configID uint) (*domain.Attempt, error) {
	var attempt domain.Attempt
	err := r.db.Where("intern_id = ? AND config_id = ?", userID, configID).First(&attempt).Error
	if err != nil {
		return nil, err
	}
	return &attempt, nil
}

func (r *attemptRepository) FindByConfigIDsAndUser(configIDs []uint, userID uint) ([]domain.Attempt, error) {
	var attempts []domain.Attempt
	err := r.db.Where("config_id IN ? AND intern_id = ?", configIDs, userID).Find(&attempts).Error
	return attempts, err
}

func (r *attemptRepository) FindByUserAndConfigIDs(userID uint, configIDs []uint, applicationID uint) ([]domain.Attempt, error) {
	query := r.db.Where("intern_id = ? AND config_id IN ?", userID, configIDs).Order("start_time ASC")
	if applicationID > 0 {
		query = query.Where("application_id = ?", applicationID)
	}
	var attempts []domain.Attempt
	err := query.Find(&attempts).Error
	return attempts, err
}

func (r *attemptRepository) FindCompletedByUserAndConfigIDs(userID, applicationID uint, configIDs []uint) ([]domain.Attempt, error) {
	var attempts []domain.Attempt
	err := r.db.
		Where("intern_id = ? AND application_id = ? AND config_id IN ? AND end_time IS NOT NULL", userID, applicationID, configIDs).
		Find(&attempts).Error
	return attempts, err
}

func (r *attemptRepository) FindByConfigID(configID uint) ([]domain.Attempt, error) {
	var attempts []domain.Attempt
	err := r.db.
		Where("config_id = ?", configID).
		Preload("User").
		Preload("EventConfig").
		Preload("Answers").
		Preload("Answers.Question").
		Find(&attempts).Error
	return attempts, err
}

func (r *attemptRepository) FindCompletedByEventAndSpec(userID, eventID, specializationID uint) ([]domain.Attempt, error) {
	var attempts []domain.Attempt
	err := r.db.
		Table("attempts").
		Joins("JOIN event_configs ON event_configs.config_id = attempts.config_id").
		Where("attempts.intern_id = ? AND event_configs.event_id = ? AND event_configs.specialization_id = ? AND attempts.end_time IS NOT NULL", userID, eventID, specializationID).
		Find(&attempts).Error
	return attempts, err
}

func (r *attemptRepository) Update(attempt *domain.Attempt) error {
	return r.db.Save(attempt).Error
}

func (r *attemptRepository) UpdateWithTx(tx *gorm.DB, attempt *domain.Attempt) error {
	return tx.Save(attempt).Error
}

func (r *attemptRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Attempt{}, id).Error
}
