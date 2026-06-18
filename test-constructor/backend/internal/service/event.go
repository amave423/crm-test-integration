package service

import (
	"test-constructor/internal/auth"
	"test-constructor/internal/client"
	"test-constructor/internal/dto"
)

type EventService interface {
	GetEvents(claims *auth.JWTClaims) (*dto.EventsListResponse, error)
	GetEventSpecializations(eventID int) (*dto.EventSpecializationsListResponse, error)
	GetAllEvents() ([]client.Event, error)
}

type eventService struct {
	crmClient client.CRMClient
}

func NewEventService(crmClient client.CRMClient) EventService {
	return &eventService{
		crmClient: crmClient,
	}
}

func (s *eventService) GetEvents(claims *auth.JWTClaims) (*dto.EventsListResponse, error) {
	events, err := s.crmClient.GetEvents()
	if err != nil {
		return nil, err
	}

	filtered := make([]dto.EventResponse, 0, len(events))
	for _, event := range events {
		if claims != nil && !claims.CanManageEvent(uint(event.ID)) {
			continue
		}
		if event.Name == "" {
			event.Name = event.Title
		}
		if event.StartDate == "" {
			event.StartDate = event.StartDateAlt
		}
		if event.EndDate == "" {
			event.EndDate = event.EndDateAlt
		}
		for i := range event.Specializations {
			if event.Specializations[i].Name == "" {
				event.Specializations[i].Name = event.Specializations[i].Title
			}
		}

		specializations := make([]dto.EventSpecializationResponse, len(event.Specializations))
		for j, spec := range event.Specializations {
			specializations[j] = dto.EventSpecializationResponse{
				ID:   spec.ID,
				Name: spec.Name,
			}
		}

		filtered = append(filtered, dto.EventResponse{
			ID:              uint(event.ID),
			Name:            event.Name,
			StartDate:       event.StartDate,
			EndDate:         event.EndDate,
			Specializations: specializations,
		})
	}

	return &dto.EventsListResponse{
		Events: filtered,
		Total:  len(filtered),
	}, nil
}

func (s *eventService) GetEventSpecializations(eventID int) (*dto.EventSpecializationsListResponse, error) {
	eventDetail, err := s.crmClient.GetEventSpecializations(eventID)
	if err != nil {
		return nil, err
	}

	specializations := make([]dto.EventSpecializationResponse, len(eventDetail.Specializations))
	for i, spec := range eventDetail.Specializations {
		if spec.Name == "" {
			spec.Name = spec.Title
		}
		specializations[i] = dto.EventSpecializationResponse{
			ID:   spec.ID,
			Name: spec.Name,
		}
	}

	return &dto.EventSpecializationsListResponse{
		EventID:         eventID,
		Specializations: specializations,
	}, nil
}

func (s *eventService) GetAllEvents() ([]client.Event, error) {
	return s.crmClient.GetEvents()
}
