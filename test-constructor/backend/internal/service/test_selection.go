package service

import (
	"test-constructor/internal/domain"
	"test-constructor/internal/dto"
	"test-constructor/internal/repository"
)

type TestSelectionService interface {
	GetSelection(userID, eventID, specializationID, applicationID uint) (*dto.TestSelectionResponse, error)
}

type testSelectionService struct {
	eventConfigRepo    repository.EventConfigRepository
	userEventRepo      repository.UserEventRepository
	attemptRepo        repository.AttemptRepository
	extraThresholdRepo repository.ExtraThresholdRepository
}

func NewTestSelectionService(
	eventConfigRepo repository.EventConfigRepository,
	userEventRepo repository.UserEventRepository,
	attemptRepo repository.AttemptRepository,
	extraThresholdRepo repository.ExtraThresholdRepository,
) TestSelectionService {
	return &testSelectionService{
		eventConfigRepo:    eventConfigRepo,
		userEventRepo:      userEventRepo,
		attemptRepo:        attemptRepo,
		extraThresholdRepo: extraThresholdRepo,
	}
}

func (s *testSelectionService) GetSelection(userID, eventID, specializationID, applicationID uint) (*dto.TestSelectionResponse, error) {
	configs, err := s.eventConfigRepo.FindByEventIDWithExtraRules(eventID, specializationID)
	if err != nil {
		return nil, err
	}

	configIDs := make([]uint, 0, len(configs))
	rulesByTestID := make(map[uint][]extraTestRule)
	for _, config := range configs {
		configIDs = append(configIDs, config.ConfigID)
		for _, rule := range config.ExtraThreshold {
			rulesByTestID[rule.TestID] = append(rulesByTestID[rule.TestID], extraTestRule{
				SourceConfigID: config.ConfigID,
				Threshold:      rule.Threshold,
				Message:        rule.Message,
			})
		}
	}

	attemptsByConfig := make(map[uint]domain.Attempt)
	if len(configIDs) > 0 {
		attempts, err := s.attemptRepo.FindByUserAndConfigIDs(userID, configIDs, applicationID)
		if err != nil {
			return nil, err
		}
		for _, attempt := range attempts {
			attemptsByConfig[attempt.ConfigID] = attempt
		}
	}

	tests := make([]dto.TestSelectionInfo, 0, len(configs))
	allCompleted := len(configs) > 0
	for _, config := range configs {
		rules := rulesByTestID[config.TestID]
		isExtra := len(rules) > 0
		status := "available"
		message := ""

		if isExtra {
			status = "locked"
			for _, rule := range rules {
				sourceAttempt, exists := attemptsByConfig[rule.SourceConfigID]
				if exists && sourceAttempt.EndTime != nil && !sourceAttempt.Passed && sourceAttempt.Score >= rule.Threshold {
					status = "available"
					message = rule.Message
					break
				}
			}
		}

		maxScore := 0
		for _, question := range config.Test.Questions {
			maxScore += question.Points
		}

		info := dto.TestSelectionInfo{
			ConfigID:    config.ConfigID,
			TestID:      config.TestID,
			TestLink:    config.TestLink.String(),
			Title:       config.Test.Title,
			Description: config.Test.Description,
			TimeLimit:   config.TimeLimit,
			Status:      status,
			MaxScore:    maxScore,
			IsExtra:     isExtra,
			Message:     message,
		}

		if attempt, exists := attemptsByConfig[config.ConfigID]; exists {
			info.AttemptID = attempt.AttemptID
			info.Score = attempt.Score
			if attempt.MaxScore > 0 {
				info.MaxScore = attempt.MaxScore
			}
			info.Passed = attempt.Passed
			if attempt.EndTime == nil {
				info.Status = "in_progress"
			} else {
				info.Status = "completed"
			}
		}

		if !isExtra && info.Status != "completed" {
			allCompleted = false
		}
		if isExtra && info.Status != "locked" && info.Status != "completed" {
			allCompleted = false
		}
		tests = append(tests, info)
	}

	eventPassed := s.checkEventPassed(userID, eventID, specializationID)

	return &dto.TestSelectionResponse{
		EventID:          eventID,
		SpecializationID: specializationID,
		ApplicationID:    applicationID,
		Tests:            tests,
		AllCompleted:     allCompleted,
		EventPassed:      eventPassed,
	}, nil
}

func (s *testSelectionService) checkEventPassed(userID, eventID, specializationID uint) bool {
	mainConfigs, err := s.eventConfigRepo.FindMainConfigsByEventAndSpec(eventID, specializationID)
	if err != nil || len(mainConfigs) == 0 {
		return false
	}

	for _, mainCfg := range mainConfigs {
		mainAttempt, err := s.attemptRepo.FindByUserAndConfig(userID, mainCfg.ConfigID)
		if err != nil || mainAttempt.EndTime == nil {
			return false
		}

		if mainAttempt.Passed {
			continue
		}

		if !s.hasPassedReplacement(userID, mainCfg.ConfigID) {
			return false
		}
	}
	return true
}

func (s *testSelectionService) hasPassedReplacement(userID, mainConfigID uint) bool {
	replacements, _ := s.extraThresholdRepo.FindReplacementsForConfigID(mainConfigID)
	mainAttempt, _ := s.attemptRepo.FindByUserAndConfig(userID, mainConfigID)

	for _, replacement := range replacements {
		if mainAttempt.Score < replacement.Threshold {
			continue
		}
		extraAttempt, err := s.attemptRepo.FindByUserAndConfig(userID, replacement.ExtraConfigID)
		if err == nil && extraAttempt.EndTime != nil && extraAttempt.Passed {
			return true
		}
	}
	return false
}

type extraTestRule struct {
	SourceConfigID uint
	Threshold      int
	Message        string
}
