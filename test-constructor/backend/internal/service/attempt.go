package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"test-constructor/internal/client"
	"test-constructor/internal/domain"
	"test-constructor/internal/dto"
	"test-constructor/internal/repository"
)

const (
	crmStatusTestingFailed = "Не прошел тестирование"
	crmStatusChatLinkSent  = "Отправлена ссылка на орг. чат"
)

type AttemptService interface {
	StartAttempt(userID uint, link string, req dto.StartAttemptRequest) (*dto.StartAttemptResponse, int, error)
	FinishAttempt(userID uint, req dto.FinishAttemptRequest) (*dto.FinishAttemptResponse, error)
	GetActiveAttempt(userID uint) (*dto.StartAttemptResponse, error)
}

type attemptService struct {
	attemptRepo        repository.AttemptRepository
	answerRepo         repository.AnswerRepository
	eventConfigRepo    repository.EventConfigRepository
	extraThresholdRepo repository.ExtraThresholdRepository
	questionRepo       repository.QuestionRepository
	userEventRepo      repository.UserEventRepository
	txManager          repository.TransactionManager
	crmClient          client.CRMClient
}

func NewAttemptService(
	attemptRepo repository.AttemptRepository,
	answerRepo repository.AnswerRepository,
	eventConfigRepo repository.EventConfigRepository,
	extraThresholdRepo repository.ExtraThresholdRepository,
	questionRepo repository.QuestionRepository,
	userEventRepo repository.UserEventRepository,
	txManager repository.TransactionManager,
	crmClient client.CRMClient,
) AttemptService {
	return &attemptService{
		attemptRepo:        attemptRepo,
		answerRepo:         answerRepo,
		eventConfigRepo:    eventConfigRepo,
		extraThresholdRepo: extraThresholdRepo,
		questionRepo:       questionRepo,
		userEventRepo:      userEventRepo,
		txManager:          txManager,
		crmClient:          crmClient,
	}
}

func (s *attemptService) StartAttempt(userID uint, link string, req dto.StartAttemptRequest) (*dto.StartAttemptResponse, int, error) {
	config, err := s.eventConfigRepo.FindByTestLink(link)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 404, errors.New("test was not found")
		}
		return nil, 500, fmt.Errorf("database error: %w", err)
	}

	if config.IsExtra {
		if !s.hasAccessToExtraConfig(userID, config.ConfigID) {
			return nil, 403, errors.New("дополнительный тест пока недоступен")
		}
	}

	existingActiveAttempt, err := s.attemptRepo.FindActiveByUser(userID)
	resumeActiveAttempt := false
	if err == nil {
		if existingActiveAttempt.ConfigID == config.ConfigID {
			resumeActiveAttempt = true
		} else {
			now := time.Now()
			existingActiveAttempt.EndTime = &now
			existingActiveAttempt.Passed = false
			existingActiveAttempt.Score = 0
			if err := s.attemptRepo.Update(existingActiveAttempt); err != nil {
				return nil, 500, fmt.Errorf("failed to close previous active attempt: %w", err)
			}
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, 500, fmt.Errorf("database error: %w", err)
	}

	if !resumeActiveAttempt {
		existingAttempt, err := s.attemptRepo.FindByUserAndConfig(userID, config.ConfigID)
		if err == nil && existingAttempt.EndTime != nil {
			return nil, 409, errors.New("you already passed this test for this application")
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 500, fmt.Errorf("database error: %w", err)
		}
	}

	publicQuestions, err := s.preparePublicQuestions(config.TestID)
	if err != nil {
		return nil, 500, err
	}

	applicationID := req.ApplicationID
	if resumeActiveAttempt && applicationID == 0 {
		applicationID = existingActiveAttempt.ApplicationID
	}
	s.syncTestingStartedStatus(applicationID)

	if resumeActiveAttempt {
		return &dto.StartAttemptResponse{
			ConfigID:      config.ConfigID,
			TestID:        config.TestID,
			ApplicationID: existingActiveAttempt.ApplicationID,
			Title:         config.Test.Title,
			Description:   config.Test.Description,
			TimeLimit:     config.TimeLimit,
			Threshold:     config.Threshold,
			Questions:     publicQuestions,
		}, 200, nil
	}

	maxScore, _ := s.questionRepo.GetMaxScoreByTestID(config.TestID)
	attempt := domain.Attempt{
		ConfigID:      config.ConfigID,
		ApplicationID: req.ApplicationID,
		InternID:      userID,
		StartTime:     time.Now(),
		MaxScore:      maxScore,
	}
	if err := s.attemptRepo.Create(&attempt); err != nil {
		return nil, 500, fmt.Errorf("failed to create attempt: %w", err)
	}

	if req.ApplicationID > 0 {
		expiresAt := attempt.StartTime.Add(time.Duration(config.TimeLimit) * time.Second)
		sessionID := fmt.Sprintf("%d", attempt.AttemptID)
		if err := s.crmClient.CreateTestSession(req.ApplicationID, config.TestID, sessionID, expiresAt); err != nil {
			fmt.Printf("CRM test session sync failed: %v\n", err)
		}
	}

	return &dto.StartAttemptResponse{
		ConfigID:      config.ConfigID,
		TestID:        config.TestID,
		ApplicationID: req.ApplicationID,
		Title:         config.Test.Title,
		Description:   config.Test.Description,
		TimeLimit:     config.TimeLimit,
		Threshold:     config.Threshold,
		Questions:     publicQuestions,
	}, 201, nil
}

func (s *attemptService) FinishAttempt(userID uint, req dto.FinishAttemptRequest) (*dto.FinishAttemptResponse, error) {
	attempt, err := s.attemptRepo.FindActiveByUser(userID)
	if err != nil {
		return nil, errors.New("active attempt was not found")
	}

	config, err := s.eventConfigRepo.FindByIDFull(attempt.ConfigID)
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки конфигурации: %w", err)
	}

	test := config.Test
	answersByQuestion := make(map[uint]dto.UserAnswer, len(req.UserAnswers))
	for _, answerInfo := range req.UserAnswers {
		answersByQuestion[answerInfo.QuestionID] = answerInfo.Answer
	}

	userPoints := 0
	maxPoints := 0
	answers := make([]domain.Answer, 0, len(test.Questions))

	for _, question := range test.Questions {
		maxPoints += question.Points

		var options domain.QuestionOptions
		if err := json.Unmarshal(question.Options, &options); err != nil {
			return nil, fmt.Errorf("question format error: %w", err)
		}

		answer, hasAnswer := answersByQuestion[question.ID]
		correct := hasAnswer && isAnswerCorrect(question, options, answer)
		answerPoints := 0
		if correct {
			answerPoints = question.Points
			userPoints += question.Points
		}

		answerJSON, err := json.Marshal(answer)
		if err != nil {
			return nil, fmt.Errorf("answer marshal error: %w", err)
		}

		answers = append(answers, domain.Answer{
			AttemptID:    attempt.AttemptID,
			QuestionID:   question.ID,
			InternAnswer: datatypes.JSON(answerJSON),
			IsCorrect:    correct,
			Points:       answerPoints,
		})
	}

	tx, err := s.txManager.Begin()
	if err != nil {
		return nil, errors.New("ошибка начала транзакции")
	}

	if err := s.answerRepo.CreateBatchWithTx(tx, answers); err != nil {
		s.txManager.Rollback(tx)
		return nil, fmt.Errorf("ошибка сохранения ответов: %w", err)
	}

	now := time.Now()
	attempt.EndTime = &now
	attempt.Score = userPoints
	attempt.MaxScore = maxPoints
	attempt.Passed = userPoints >= config.Threshold

	if err := s.attemptRepo.UpdateWithTx(tx, attempt); err != nil {
		s.txManager.Rollback(tx)
		return nil, fmt.Errorf("ошибка обновления попытки: %w", err)
	}

	if err := s.txManager.Commit(tx); err != nil {
		return nil, errors.New("ошибка сохранения результатов")
	}

	resultText := config.FailText
	if attempt.Passed {
		resultText = config.SuccessText
	}

	applicationStatus := ""
	allCompleted := false
	if attempt.ApplicationID > 0 {
		applicationStatus = s.resolveCRMApplicationStatus(attempt)
		allCompleted = applicationStatus != ""

		crmResult := client.CRMResultData{
			SessionID:         fmt.Sprintf("%d", attempt.AttemptID),
			Score:             userPoints,
			MaxScore:          maxPoints,
			IsPassed:          attempt.Passed,
			CompletedAt:       now.Format("2006-01-02T15:04:05Z"),
			StartedAt:         attempt.StartTime.Format("2006-01-02T15:04:05Z"),
			ApplicationStatus: applicationStatus,
		}

		if err := s.crmClient.SendTestResult(attempt.ApplicationID, crmResult); err != nil {
			fmt.Printf("CRM result sync failed: %v\n", err)
			if applicationStatus != "" {
				if err := s.crmClient.SendApplicationStatus(attempt.ApplicationID, applicationStatus); err != nil {
					fmt.Printf("CRM status sync failed: %v\n", err)
				}
			}
		}
	}

	return &dto.FinishAttemptResponse{
		Result:        resultText,
		Score:         userPoints,
		MaxTestPoints: maxPoints,
		Passed:        attempt.Passed,
		AllCompleted:  allCompleted,
	}, nil
}

func (s *attemptService) GetActiveAttempt(userID uint) (*dto.StartAttemptResponse, error) {
	attempt, err := s.attemptRepo.FindActiveByUser(userID)
	if err != nil {
		return nil, errors.New("активная попытка не найдена")
	}

	config, err := s.eventConfigRepo.FindByID(attempt.ConfigID)
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки конфигурации: %w", err)
	}

	publicQuestions, err := s.preparePublicQuestions(config.TestID)
	if err != nil {
		return nil, err
	}

	return &dto.StartAttemptResponse{
		ConfigID:      config.ConfigID,
		TestID:        config.TestID,
		ApplicationID: attempt.ApplicationID,
		Title:         config.Test.Title,
		Description:   config.Test.Description,
		TimeLimit:     config.TimeLimit,
		Threshold:     config.Threshold,
		Questions:     publicQuestions,
	}, nil
}

func (s *attemptService) hasAccessToExtraConfig(userID, extraConfigID uint) bool {
	extraThreshold, err := s.extraThresholdRepo.FindByExtraConfigID(extraConfigID)
	if err != nil {
		return false
	}

	attempt, err := s.attemptRepo.FindByUserAndConfig(userID, extraThreshold.ConfigID)
	if err != nil || attempt.EndTime == nil {
		return false
	}

	return !attempt.Passed && attempt.Score >= extraThreshold.Threshold
}

func (s *attemptService) syncTestingStartedStatus(applicationID uint) {
	if applicationID == 0 {
		return
	}
	fmt.Printf("CRM testing started status synced for application %d\n", applicationID)
}

func (s *attemptService) resolveCRMApplicationStatus(attempt *domain.Attempt) string {
	if attempt.ApplicationID == 0 {
		return ""
	}

	configs, err := s.eventConfigRepo.FindByEventAndSpecializationAll(attempt.EventConfig.EventID, attempt.EventConfig.SpecializationID)
	if err != nil {
		return ""
	}

	configByTestID := make(map[uint]domain.EventConfig, len(configs))
	extraTestIDs := make(map[uint]struct{})
	configIDs := make([]uint, 0, len(configs))
	for _, config := range configs {
		configByTestID[config.TestID] = config
		configIDs = append(configIDs, config.ConfigID)
		for _, rule := range config.ExtraThreshold {
			extraTestIDs[rule.TestID] = struct{}{}
		}
	}

	finishedAttempts, err := s.attemptRepo.FindCompletedByUserAndConfigIDs(attempt.InternID, attempt.ApplicationID, configIDs)
	if err != nil {
		return ""
	}

	attemptByConfigID := make(map[uint]domain.Attempt, len(finishedAttempts))
	for _, finishedAttempt := range finishedAttempts {
		attemptByConfigID[finishedAttempt.ConfigID] = finishedAttempt
	}

	hasRequiredConfigs := false
	testingFailed := false

	for _, config := range configs {
		if _, isExtra := extraTestIDs[config.TestID]; isExtra {
			continue
		}
		hasRequiredConfigs = true

		mainAttempt, completed := attemptByConfigID[config.ConfigID]
		if !completed {
			return ""
		}
		if mainAttempt.Passed {
			continue
		}

		applicableExtraTests := make([]uint, 0)
		for _, rule := range config.ExtraThreshold {
			if mainAttempt.Score >= rule.Threshold {
				applicableExtraTests = append(applicableExtraTests, rule.TestID)
			}
		}
		if len(applicableExtraTests) == 0 {
			testingFailed = true
			continue
		}

		extraPassed := false
		extraPending := false
		for _, extraTestID := range applicableExtraTests {
			extraConfig, exists := configByTestID[extraTestID]
			if !exists {
				continue
			}
			extraAttempt, completed := attemptByConfigID[extraConfig.ConfigID]
			if !completed {
				extraPending = true
				continue
			}
			if extraAttempt.Passed {
				extraPassed = true
				break
			}
		}

		if extraPassed {
			continue
		}
		if extraPending {
			return ""
		}
		testingFailed = true
	}

	if !hasRequiredConfigs || testingFailed {
		return crmStatusTestingFailed
	}
	return crmStatusChatLinkSent
}

func (s *attemptService) preparePublicQuestions(testID uint) ([]dto.QuestionPublic, error) {
	questions, err := s.questionRepo.FindByTestID(testID)
	if err != nil {
		return nil, err
	}

	public := make([]dto.QuestionPublic, len(questions))
	for i, q := range questions {
		var opts domain.QuestionOptions
		json.Unmarshal(q.Options, &opts)

		pub := dto.QuestionPublic{
			QuestionID:  q.ID,
			Text:        q.Text,
			Points:      q.Points,
			OrderNumber: q.OrderNumber,
			Type:        q.Type,
		}

		switch q.Type {
		case domain.SingleChoice, domain.MultipleChoice:
			choices := make([]dto.PublicChoice, len(opts.Choices))
			for j, c := range opts.Choices {
				choices[j] = dto.PublicChoice{Text: c.Text, Index: j}
			}
			rand.Shuffle(len(choices), func(i, j int) {
				choices[i], choices[j] = choices[j], choices[i]
			})
			pub.Options.Choices = choices
		case domain.Matching:
			left := make([]string, len(opts.MatchingPairs))
			right := make([]string, len(opts.MatchingPairs))
			for j, p := range opts.MatchingPairs {
				left[j] = p.LeftColumn
				right[j] = p.RightColumn
			}
			rand.Shuffle(len(right), func(i, j int) {
				right[i], right[j] = right[j], right[i]
			})
			pub.Options.Matching = &dto.PublicMatching{LeftColumn: left, RightColumn: right}
		case domain.CorrectOrder:
			texts := make([]string, len(opts.Sequence))
			for j, s := range opts.Sequence {
				texts[j] = s.Text
			}
			rand.Shuffle(len(texts), func(i, j int) {
				texts[i], texts[j] = texts[j], texts[i]
			})
			pub.Options.Sequence = texts
		case domain.TextInput:
			pub.Options.CaseSensitive = opts.CaseSensitive
		}
		public[i] = pub
	}
	return public, nil
}

func isAnswerCorrect(question domain.Question, options domain.QuestionOptions, answer dto.UserAnswer) bool {
	switch question.Type {
	case domain.SingleChoice, domain.MultipleChoice:
		for i, choice := range options.Choices {
			selected := i < len(answer.Choices) && answer.Choices[i]
			if choice.IsTrue != selected {
				return false
			}
		}
		return true

	case domain.Matching:
		if len(answer.MatchingPairs) != len(options.MatchingPairs) {
			return false
		}
		selected := make(map[string]string, len(answer.MatchingPairs))
		for _, pair := range answer.MatchingPairs {
			selected[pair.Left] = pair.Right
		}
		for _, pair := range options.MatchingPairs {
			if selected[pair.LeftColumn] != pair.RightColumn {
				return false
			}
		}
		return true

	case domain.CorrectOrder:
		if len(answer.Sequence) != len(options.Sequence) {
			return false
		}
		expected := make([]domain.SequenceItem, len(options.Sequence))
		copy(expected, options.Sequence)
		selected := make([]domain.SequenceItem, len(answer.Sequence))
		for i, s := range answer.Sequence {
			selected[i] = domain.SequenceItem{Text: s.Text, Order: s.Order}
		}
		sort.Slice(expected, func(i, j int) bool { return expected[i].Order < expected[j].Order })
		sort.Slice(selected, func(i, j int) bool { return selected[i].Order < selected[j].Order })
		for i := range expected {
			if expected[i].Text != selected[i].Text {
				return false
			}
		}
		return true

	case domain.TextInput:
		input := answer.UserInput
		for _, expected := range options.CorrectInput {
			if options.CaseSensitive {
				if input == expected {
					return true
				}
			} else if strings.EqualFold(input, expected) {
				return true
			}
		}
		return false
	}

	return false
}
