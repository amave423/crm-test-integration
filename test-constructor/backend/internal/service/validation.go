package service

import (
	"fmt"
	"test-constructor/internal/dto"
	"test-constructor/internal/repository"
)

type ValidationService interface {
	ValidateThreshold(testID uint, threshold int) error
	ValidateEventConfig(req dto.CreateEventConfigRequest) error
}

type validationService struct {
	questionRepo repository.QuestionRepository
}

func NewValidationService(
	questionRepo repository.QuestionRepository,
) ValidationService {
	return &validationService{
		questionRepo: questionRepo,
	}
}

func (s *validationService) ValidateThreshold(testID uint, threshold int) error {
	maxScore, err := s.questionRepo.GetMaxScoreByTestID(testID)
	if err != nil {
		return fmt.Errorf("ошибка получения максимального балла: %w", err)
	}

	if maxScore == 0 {
		return fmt.Errorf("тест с ID %d не содержит вопросов или все вопросы имеют 0 баллов", testID)
	}

	if threshold > maxScore {
		return fmt.Errorf("порог (%d) не может быть выше максимального балла за тест (%d)",
			threshold, maxScore)
	}

	return nil
}

func (s *validationService) ValidateEventConfig(
	req dto.CreateEventConfigRequest,
) error {
	if err := s.ValidateThreshold(req.TestID, req.Threshold); err != nil {
		return err
	}

	for i, extra := range req.ExtraThreshold {
		if err := s.ValidateThreshold(extra.TestID, extra.Threshold); err != nil {
			return fmt.Errorf("дополнительный тест #%d: %w", i+1, err)
		}

		mainMaxScore, _ := s.questionRepo.GetMaxScoreByTestID(req.TestID)
		if extra.Threshold > mainMaxScore {
			return fmt.Errorf(
				"порог перехода (%d) к дополнительному тесту #%d не может быть выше "+
					"максимального балла основного теста (%d)",
				extra.Threshold, i+1, mainMaxScore,
			)
		}
	}

	return nil
}
