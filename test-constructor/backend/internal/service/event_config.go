package service

import (
	"errors"
	"fmt"
	"test-constructor/internal/client"

	"test-constructor/internal/domain"
	"test-constructor/internal/dto"
	"test-constructor/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventConfigService interface {
	CreateOrUpdateConfig(creatorID uint, req dto.CreateEventConfigRequest) (*dto.CreateEventConfigResponse, bool, error)
	UpdateConfig(configID, creatorID uint, req dto.UpdateEventConfigRequest) (*dto.UpdateEventConfigResponse, error)
	GetConfig(id uint) (*dto.EventConfigResponse, error)
	GetConfigsByEventID(eventID uint) (*dto.EventConfigsResponse, error)
}

type eventConfigService struct {
	questionRepo       repository.QuestionRepository
	testRepo           repository.TestRepository
	eventConfigRepo    repository.EventConfigRepository
	extraThresholdRepo repository.ExtraThresholdRepository
	txManager          repository.TransactionManager
	crmClient          client.CRMClient
	validationService  ValidationService
}

func NewEventConfigService(
	questionRepo repository.QuestionRepository,
	testRepo repository.TestRepository,
	eventConfigRepo repository.EventConfigRepository,
	extraThresholdRepo repository.ExtraThresholdRepository,
	txManager repository.TransactionManager,
	crmClient client.CRMClient,
	validationService ValidationService,
) EventConfigService {
	return &eventConfigService{
		questionRepo:       questionRepo,
		testRepo:           testRepo,
		eventConfigRepo:    eventConfigRepo,
		extraThresholdRepo: extraThresholdRepo,
		txManager:          txManager,
		crmClient:          crmClient,
		validationService:  validationService,
	}
}

func (s *eventConfigService) CreateOrUpdateConfig(creatorID uint, req dto.CreateEventConfigRequest) (*dto.CreateEventConfigResponse, bool, error) {
	if req.EventID < 1 || req.TestID < 1 {
		return nil, false, errors.New("ID должен быть положительным")
	}

	if req.Threshold < 1 {
		return nil, false, errors.New("пороговое значение должно быть положительным")
	}

	tx, err := s.txManager.Begin()
	if err != nil {
		return nil, false, errors.New("ошибка базы данных")
	}

	var eventConfig domain.EventConfig
	existingConfig, err := s.eventConfigRepo.FindByEventSpecializationAndTest(tx, req.EventID, req.SpecializationID, req.TestID)
	created := false

	if err == nil {
		eventConfig = *existingConfig
		updates := domain.EventConfig{
			CreatorID:   creatorID,
			SuccessText: req.SuccessText,
			FailText:    req.FailText,
			TimeLimit:   req.TimeLimit,
			Threshold:   req.Threshold,
		}
		if err := s.eventConfigRepo.UpdateWithTx(tx, &eventConfig, updates); err != nil {
			s.txManager.Rollback(tx)
			return nil, false, fmt.Errorf("ошибка обновления настройки: %w", err)
		}
		if err := s.extraThresholdRepo.DeleteByConfigIDWithTx(tx, eventConfig.ConfigID); err != nil {
			s.txManager.Rollback(tx)
			return nil, false, fmt.Errorf("ошибка удаления старых порогов: %w", err)
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		eventConfig = domain.EventConfig{
			EventID:          req.EventID,
			SpecializationID: req.SpecializationID,
			TestID:           req.TestID,
			CreatorID:        creatorID,
			SuccessText:      req.SuccessText,
			FailText:         req.FailText,
			TimeLimit:        req.TimeLimit,
			TestLink:         uuid.New(),
			Threshold:        req.Threshold,
		}
		if err := s.eventConfigRepo.CreateWithTx(tx, &eventConfig); err != nil {
			s.txManager.Rollback(tx)
			return nil, false, fmt.Errorf("ошибка создания настройки: %w", err)
		}
		created = true
	} else {
		s.txManager.Rollback(tx)
		return nil, false, fmt.Errorf("ошибка поиска настройки: %w", err)
	}

	for _, eThreshold := range req.ExtraThreshold {
		extraThreshold := domain.ExtraThreshold{
			ConfigID:  eventConfig.ConfigID,
			Threshold: eThreshold.Threshold,
			Message:   eThreshold.Message,
			TestID:    eThreshold.TestID,
		}
		if err := s.extraThresholdRepo.CreateWithTx(tx, &extraThreshold); err != nil {
			s.txManager.Rollback(tx)
			return nil, false, err
		}
	}

	if eventConfig.TestLink == uuid.Nil {
		eventConfig.TestLink = uuid.New()
		if err := s.eventConfigRepo.UpdateTestLinkWithTx(tx, eventConfig.ConfigID, eventConfig.TestLink); err != nil {
			s.txManager.Rollback(tx)
			return nil, false, fmt.Errorf("ошибка создания ссылки на тест: %w", err)
		}
	}

	if err := s.txManager.Commit(tx); err != nil {
		return nil, false, errors.New("ошибка сохранения изменений")
	}

	return &dto.CreateEventConfigResponse{
		ConfigID: eventConfig.ConfigID,
		TestLink: eventConfig.TestLink.String(),
		Created:  created,
	}, created, nil
}

func (s *eventConfigService) UpdateConfig(configID, creatorID uint, req dto.UpdateEventConfigRequest) (*dto.UpdateEventConfigResponse, error) {
	if req.EventID < 1 || req.TestID < 1 {
		return nil, errors.New("ID должен быть положительным")
	}

	if req.Threshold < 1 {
		return nil, errors.New("пороговое значение должно быть положительным")
	}

	tx, err := s.txManager.Begin()
	if err != nil {
		return nil, errors.New("ошибка базы данных")
	}

	existingConfig, err := s.eventConfigRepo.FindByID(configID)
	if err != nil {
		s.txManager.Rollback(tx)
		return nil, errors.New("конфигурация не найдена")
	}

	updates := domain.EventConfig{
		EventID:          req.EventID,
		CreatorID:        creatorID,
		SpecializationID: req.SpecializationID,
		TestID:           req.TestID,
		SuccessText:      req.SuccessText,
		FailText:         req.FailText,
		TimeLimit:        req.TimeLimit,
		Threshold:        req.Threshold,
	}

	if err := s.eventConfigRepo.UpdateFieldsWithTx(tx, configID, updates); err != nil {
		s.txManager.Rollback(tx)
		return nil, fmt.Errorf("ошибка обновления настройки: %w", err)
	}

	if err := s.extraThresholdRepo.DeleteByConfigIDWithTx(tx, configID); err != nil {
		s.txManager.Rollback(tx)
		return nil, fmt.Errorf("ошибка удаления старых порогов: %w", err)
	}

	for _, eThreshold := range req.ExtraThreshold {
		extraThreshold := domain.ExtraThreshold{
			ConfigID:  configID,
			Threshold: eThreshold.Threshold,
			Message:   eThreshold.Message,
			TestID:    eThreshold.TestID,
		}
		if err := s.extraThresholdRepo.CreateWithTx(tx, &extraThreshold); err != nil {
			s.txManager.Rollback(tx)
			return nil, fmt.Errorf("ошибка создания порога: %w", err)
		}
	}

	if err := s.txManager.Commit(tx); err != nil {
		return nil, errors.New("ошибка сохранения изменений")
	}

	return &dto.UpdateEventConfigResponse{
		ConfigID: configID,
		TestLink: existingConfig.TestLink.String(),
	}, nil
}

func (s *eventConfigService) GetConfigsByEventID(eventID uint) (*dto.EventConfigsResponse, error) {
	configs, err := s.eventConfigRepo.FindByEventID(eventID)
	if err != nil {
		return nil, err
	}

	response := &dto.EventConfigsResponse{
		Configs: make([]dto.EventConfigResponse, 0, len(configs)),
	}

	for _, config := range configs {
		extra := make([]dto.ExtraThresholdResponse, 0, len(config.ExtraThreshold))
		for _, item := range config.ExtraThreshold {
			extra = append(extra, dto.ExtraThresholdResponse{
				Threshold: item.Threshold,
				Message:   item.Message,
				TestID:    item.TestID,
			})
		}

		response.Configs = append(response.Configs, dto.EventConfigResponse{
			ConfigID:         config.ConfigID,
			EventID:          config.EventID,
			SpecializationID: config.SpecializationID,
			TestID:           config.TestID,
			SuccessText:      config.SuccessText,
			FailText:         config.FailText,
			TimeLimit:        config.TimeLimit,
			Threshold:        config.Threshold,
			TestLink:         config.TestLink.String(),
			ExtraThreshold:   extra,
		})
	}

	return response, nil
}

func (s *eventConfigService) GetConfig(id uint) (*dto.EventConfigResponse, error) {
	config, err := s.eventConfigRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("конфигурация не найдена")
	}

	extra := make([]dto.ExtraThresholdResponse, 0, len(config.ExtraThreshold))
	for _, item := range config.ExtraThreshold {
		extra = append(extra, dto.ExtraThresholdResponse{
			Threshold: item.Threshold,
			Message:   item.Message,
			TestID:    item.TestID,
		})
	}

	return &dto.EventConfigResponse{
		ConfigID:         config.ConfigID,
		EventID:          config.EventID,
		SpecializationID: config.SpecializationID,
		TestID:           config.TestID,
		SuccessText:      config.SuccessText,
		FailText:         config.FailText,
		TimeLimit:        config.TimeLimit,
		Threshold:        config.Threshold,
		TestLink:         config.TestLink.String(),
		ExtraThreshold:   extra,
	}, nil
}
