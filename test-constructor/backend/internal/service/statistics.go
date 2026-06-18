package service

import (
	"fmt"
	"sort"
	"test-constructor/internal/domain"
	"test-constructor/internal/dto"
	"test-constructor/internal/repository"
)

type StatisticsService interface {
	GetInternList(scopedEventIDs []uint) (*dto.GetUsersResponse, error)
	GetUserStatistics(userID uint, scopedEventIDs []uint) (*dto.UserStatisticsResponse, error)
	GetEventStatistics(eventID uint, filter *dto.StatisticsFilter) (*dto.StatisticsResponse, error)
}

type statisticsService struct {
	statsRepo    repository.StatisticsRepository
	eventService EventService
}

func NewStatisticsService(
	statsRepo repository.StatisticsRepository,
	eventService EventService,
) StatisticsService {
	return &statisticsService{
		statsRepo:    statsRepo,
		eventService: eventService,
	}
}

func (s *statisticsService) GetInternList(scopedEventIDs []uint) (*dto.GetUsersResponse, error) {
	users, err := s.statsRepo.FindInterns(scopedEventIDs)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка стажёров: %w", err)
	}

	interns := make([]dto.UserInfo, len(users))
	for i, user := range users {
		interns[i] = dto.UserInfo{
			ID:      user.ID,
			Name:    user.Name,
			Surname: user.Surname,
			Email:   user.Email,
		}
	}

	return &dto.GetUsersResponse{
		Users: interns,
	}, nil
}

func (s *statisticsService) GetUserStatistics(userID uint, scopedEventIDs []uint) (*dto.UserStatisticsResponse, error) {
	user, err := s.statsRepo.FindUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("пользователь не найден: %w", err)
	}

	attempts, err := s.statsRepo.FindCompletedAttemptsByUserID(userID, scopedEventIDs)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения попыток: %w", err)
	}

	eventNames := s.loadEventNames()
	attemptDetails := make([]dto.UserAttemptDetail, 0, len(attempts))
	for _, attempt := range attempts {
		cfg := attempt.EventConfig
		maxScore := 0
		for _, question := range cfg.Test.Questions {
			maxScore += question.Points
		}

		eventName := eventNames[cfg.EventID]
		if eventName == "" {
			eventName = fmt.Sprintf("Мероприятие #%d", cfg.EventID)
		}

		attemptDetails = append(attemptDetails, dto.UserAttemptDetail{
			AttemptID: attempt.AttemptID,
			TestTitle: cfg.Test.Title,
			EventName: eventName,
			IsExtra:   false,
			Score:     attempt.Score,
			MaxScore:  maxScore,
			Passed:    attempt.Passed,
			Questions: buildQuestionStats(attempt.Answers),
		})
	}

	return &dto.UserStatisticsResponse{
		UserID:    user.ID,
		FirstName: user.Name,
		LastName:  user.Surname,
		Email:     user.Email,
		Attempts:  attemptDetails,
	}, nil
}

func (s *statisticsService) GetEventStatistics(eventID uint, filter *dto.StatisticsFilter) (*dto.StatisticsResponse, error) {
	var isExtra *bool
	if filter != nil {
		isExtra = filter.IsExtra
	}

	configs, err := s.statsRepo.FindConfigsByEventID(eventID, isExtra)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения конфигураций: %w", err)
	}

	if len(configs) == 0 {
		return &dto.StatisticsResponse{
			Attempts: []dto.UserAttemptInfo{},
		}, nil
	}

	configIDs := make([]uint, len(configs))
	configMap := make(map[uint]domain.EventConfig)
	questionMap := make(map[uint]map[uint]domain.Question)

	for i, cfg := range configs {
		configIDs[i] = cfg.ConfigID
		configMap[cfg.ConfigID] = cfg
		questionMap[cfg.ConfigID] = make(map[uint]domain.Question)
		for _, q := range cfg.Test.Questions {
			questionMap[cfg.ConfigID][q.ID] = q
		}
	}

	attempts, err := s.statsRepo.FindCompletedAttemptsByConfigIDs(configIDs)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения попыток: %w", err)
	}

	userAttempts := make([]dto.UserAttemptInfo, 0, len(attempts))
	for _, attempt := range attempts {
		cfg, cfgExists := configMap[attempt.ConfigID]
		if !cfgExists {
			continue
		}

		maxScore := 0
		for _, question := range cfg.Test.Questions {
			maxScore += question.Points
		}

		timeSpent := 0
		if attempt.EndTime != nil {
			duration := attempt.EndTime.Sub(attempt.StartTime)
			timeSpent = int(duration.Minutes())
		}

		questions := buildQuestionStats(attempt.Answers)

		attemptInfo := dto.UserAttemptInfo{
			UserID:    attempt.InternID,
			FirstName: attempt.User.Name,
			LastName:  attempt.User.Surname,
			Email:     attempt.User.Email,
			Score:     attempt.Score,
			MaxScore:  maxScore,
			Passed:    attempt.Passed,
			IsExtra:   attempt.EventConfig.IsExtra,
			TimeSpent: timeSpent,
			Questions: questions,
		}

		userAttempts = append(userAttempts, attemptInfo)
	}

	return &dto.StatisticsResponse{
		Attempts: userAttempts,
	}, nil
}

func (s *statisticsService) loadEventNames() map[uint]string {
	eventNames := make(map[uint]string)
	events, err := s.eventService.GetAllEvents()
	if err != nil {
		return eventNames
	}

	for _, event := range events {
		if event.ID > 0 && event.Name != "" {
			eventNames[uint(event.ID)] = event.Name
		}
	}

	return eventNames
}

func buildQuestionStats(answers []domain.Answer) []dto.QuestionStatInfo {
	stats := make([]dto.QuestionStatInfo, 0, len(answers))
	for _, answer := range answers {
		question := answer.Question
		stats = append(stats, dto.QuestionStatInfo{
			Text:         question.Text,
			Points:       answer.Points,
			MaxPoints:    question.Points,
			IsCorrect:    answer.IsCorrect,
			QuestionType: string(question.Type),
			OrderNumber:  question.OrderNumber,
		})
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].OrderNumber < stats[j].OrderNumber
	})

	return stats
}
