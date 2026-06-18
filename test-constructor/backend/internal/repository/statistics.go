package repository

import (
	"test-constructor/internal/domain"

	"gorm.io/gorm"
)

type StatisticsRepository interface {
	FindInterns(scopedEventIDs []uint) ([]domain.User, error)
	FindUserByID(userID uint) (*domain.User, error)
	FindCompletedAttemptsByUserID(userID uint, scopedEventIDs []uint) ([]domain.Attempt, error)
	FindCompletedAttemptsByConfigIDs(configIDs []uint) ([]domain.Attempt, error)
	FindConfigsByEventID(eventID uint, isExtra *bool) ([]domain.EventConfig, error)
}

type statisticsRepository struct {
	db *gorm.DB
}

func NewStatisticsRepository(db *gorm.DB) StatisticsRepository {
	return &statisticsRepository{db: db}
}

func (r *statisticsRepository) FindInterns(scopedEventIDs []uint) ([]domain.User, error) {
	var internRole domain.Role
	if err := r.db.Where("code = ?", "intern").First(&internRole).Error; err != nil {
		return nil, err
	}

	query := r.db.Where("role_id = ?", internRole.ID)

	if len(scopedEventIDs) > 0 {
		var scopedUserIDs []uint
		r.db.Model(&domain.Attempt{}).
			Joins("JOIN event_configs ON event_configs.config_id = attempts.config_id").
			Where("event_configs.event_id IN ?", scopedEventIDs).
			Distinct().
			Pluck("attempts.intern_id", &scopedUserIDs)

		if len(scopedUserIDs) == 0 {
			return []domain.User{}, nil
		}
		query = query.Where("id IN ?", scopedUserIDs)
	}

	var users []domain.User
	err := query.Order("surname, name").Find(&users).Error
	return users, err
}

func (r *statisticsRepository) FindUserByID(userID uint) (*domain.User, error) {
	var user domain.User
	err := r.db.First(&user, userID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *statisticsRepository) FindCompletedAttemptsByUserID(userID uint, scopedEventIDs []uint) ([]domain.Attempt, error) {
	query := r.db.Where("intern_id = ? AND end_time IS NOT NULL", userID)
	if len(scopedEventIDs) > 0 {
		query = query.Joins("JOIN event_configs ON event_configs.config_id = attempts.config_id").Where("event_configs.event_id IN ?", scopedEventIDs)
	}

	var attempts []domain.Attempt
	err := query.
		Preload("EventConfig").
		Preload("EventConfig.Test").
		Preload("EventConfig.Test.Questions").
		Preload("Answers").
		Preload("Answers.Question").
		Order("end_time DESC").
		Find(&attempts).Error
	return attempts, err
}

func (r *statisticsRepository) FindCompletedAttemptsByConfigIDs(configIDs []uint) ([]domain.Attempt, error) {
	var attempts []domain.Attempt
	err := r.db.Where("config_id IN ? AND end_time IS NOT NULL", configIDs).
		Preload("User").
		Preload("EventConfig").
		Preload("Answers").
		Preload("Answers.Question").
		Order("end_time DESC").
		Find(&attempts).Error
	return attempts, err
}

func (r *statisticsRepository) FindConfigsByEventID(eventID uint, isExtra *bool) ([]domain.EventConfig, error) {
	query := r.db.Where("event_id = ?", eventID)
	if isExtra != nil {
		query = query.Where("is_extra = ?", *isExtra)
	}

	var configs []domain.EventConfig
	err := query.Preload("Test.Questions").Find(&configs).Error
	return configs, err
}
