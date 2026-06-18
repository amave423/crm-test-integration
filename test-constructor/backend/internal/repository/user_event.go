package repository

import (
	"test-constructor/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserEventRepository interface {
	Create(userEvent *domain.UserEvent) error
	FindByUserAndEventAndSpec(userID, eventID, specializationID uint) (*domain.UserEvent, error)
	FindByUserAndEvent(userID, eventID uint) (*domain.UserEvent, error)
	FindByUserID(userID uint) ([]domain.UserEvent, error)
	Update(userEvent *domain.UserEvent) error
	CreateOrUpdate(userEvent domain.UserEvent) error
	Delete(id uint) error
}

type userEventRepository struct {
	db *gorm.DB
}

func NewUserEventRepository(db *gorm.DB) UserEventRepository {
	return &userEventRepository{db: db}
}

func (r *userEventRepository) Create(userEvent *domain.UserEvent) error {
	return r.db.Create(userEvent).Error
}

func (r *userEventRepository) FindByUserAndEventAndSpec(userID, eventID, specializationID uint) (*domain.UserEvent, error) {
	var ue domain.UserEvent
	err := r.db.Where("user_id = ? AND event_id = ? AND specialization_id = ?", userID, eventID, specializationID).First(&ue).Error
	if err != nil {
		return nil, err
	}
	return &ue, nil
}

func (r *userEventRepository) FindByUserAndEvent(userID, eventID uint) (*domain.UserEvent, error) {
	var ue domain.UserEvent
	err := r.db.Where("user_id = ? AND event_id = ?", userID, eventID).First(&ue).Error
	if err != nil {
		return nil, err
	}
	return &ue, nil
}

func (r *userEventRepository) FindByUserID(userID uint) ([]domain.UserEvent, error) {
	var userEvents []domain.UserEvent
	err := r.db.Where("user_id = ?", userID).Order("id DESC").Find(&userEvents).Error
	return userEvents, err
}

func (r *userEventRepository) Update(userEvent *domain.UserEvent) error {
	return r.db.Save(userEvent).Error
}

func (r *userEventRepository) CreateOrUpdate(userEvent domain.UserEvent) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "event_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"application_id", "specialization_id"}),
	}).Create(&userEvent).Error
}

func (r *userEventRepository) Delete(id uint) error {
	return r.db.Delete(&domain.UserEvent{}, id).Error
}
