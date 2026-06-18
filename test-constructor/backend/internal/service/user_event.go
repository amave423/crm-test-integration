package service

import (
	"errors"

	"test-constructor/internal/domain"
	"test-constructor/internal/dto"
	"test-constructor/internal/repository"
)

type UserEventService interface {
	CreateUserEvent(userID uint, req dto.CreateUserEventRequest) (*dto.UserEventResponse, error)
	GetUserEvents(userID uint) (*dto.UserEventsListResponse, error)
	CreateOrUpdateUserEvent(userEvent *domain.UserEvent) error
}

type userEventService struct {
	userEventRepo repository.UserEventRepository
}

func NewUserEventService(userEventRepo repository.UserEventRepository) UserEventService {
	return &userEventService{
		userEventRepo: userEventRepo,
	}
}

func (s *userEventService) CreateUserEvent(userID uint, req dto.CreateUserEventRequest) (*dto.UserEventResponse, error) {
	if req.EventID == 0 {
		return nil, errors.New("event_id обязателен")
	}
	if req.ApplicationID == 0 {
		return nil, errors.New("application_id обязателен")
	}

	userEvent := domain.UserEvent{
		UserID:           userID,
		EventID:          req.EventID,
		SpecializationID: req.SpecializationID,
		ApplicationID:    req.ApplicationID,
	}

	if err := s.userEventRepo.CreateOrUpdate(userEvent); err != nil {
		return nil, err
	}

	return &dto.UserEventResponse{
		ID:               userEvent.ID,
		EventID:          userEvent.EventID,
		SpecializationID: userEvent.SpecializationID,
		ApplicationID:    userEvent.ApplicationID,
	}, nil
}

func (s *userEventService) GetUserEvents(userID uint) (*dto.UserEventsListResponse, error) {
	userEvents, err := s.userEventRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	events := make([]dto.UserEventResponse, 0, len(userEvents))
	for _, ue := range userEvents {
		events = append(events, dto.UserEventResponse{
			ID:               ue.ID,
			EventID:          ue.EventID,
			SpecializationID: ue.SpecializationID,
			ApplicationID:    ue.ApplicationID,
		})
	}

	return &dto.UserEventsListResponse{
		Events: events,
	}, nil
}

func (s *userEventService) CreateOrUpdateUserEvent(userEvent *domain.UserEvent) error {
	if userEvent.UserID == 0 || userEvent.EventID == 0 {
		return errors.New("user_id и event_id обязательны")
	}

	return s.userEventRepo.CreateOrUpdate(*userEvent)
}
