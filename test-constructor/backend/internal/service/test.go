package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"test-constructor/internal/domain"
	"test-constructor/internal/dto"
	"test-constructor/internal/repository"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type TestService interface {
	CreateTest(creatorID uint, req dto.CreateTestRequest) (*dto.CreateTestResponse, error)
	GetTests() (*dto.TestsListResponse, error)
	GetTestByID(id uint) (*dto.TestDetailResponse, error)
	DeleteTest(testID uint) error
}

type testService struct {
	testRepo     repository.TestRepository
	questionRepo repository.QuestionRepository
	txManager    repository.TransactionManager
}

func NewTestService(
	testRepo repository.TestRepository,
	questionRepo repository.QuestionRepository,
	txManager repository.TransactionManager,
) TestService {
	return &testService{
		testRepo:     testRepo,
		questionRepo: questionRepo,
		txManager:    txManager,
	}
}

func (s *testService) CreateTest(creatorID uint, req dto.CreateTestRequest) (*dto.CreateTestResponse, error) {
	if req.Title == "" {
		return nil, errors.New("название теста обязательно")
	}

	if len(req.Questions) == 0 {
		return nil, errors.New("тест должен содержать хотя бы один вопрос")
	}

	tx, err := s.txManager.Begin()
	if err != nil {
		return nil, errors.New("ошибка базы данных")
	}

	test := domain.Test{
		Title:       req.Title,
		Description: req.Description,
		CreatorID:   creatorID,
	}

	if err := s.testRepo.CreateWithTx(tx, &test); err != nil {
		s.txManager.Rollback(tx)
		return nil, fmt.Errorf("ошибка создания теста: %w", err)
	}

	for i, qReq := range req.Questions {
		qType, err := domain.ParseQType(qReq.Type)
		if err != nil {
			s.txManager.Rollback(tx)
			return nil, err
		}

		if err := s.validateQuestionOptions(qType, qReq.Options); err != nil {
			s.txManager.Rollback(tx)
			return nil, err
		}

		optionsJSON, err := json.Marshal(qReq.Options)
		if err != nil {
			s.txManager.Rollback(tx)
			return nil, fmt.Errorf("ошибка преобразования опций вопроса %d: %w", i+1, err)
		}

		question := domain.Question{
			TestID:      test.ID,
			Text:        qReq.Text,
			Points:      qReq.Points,
			Type:        qType,
			OrderNumber: qReq.OrderNumber,
			Options:     datatypes.JSON(optionsJSON),
		}

		if err := s.questionRepo.CreateWithTx(tx, &question); err != nil {
			s.txManager.Rollback(tx)
			return nil, fmt.Errorf("ошибка создания вопроса %d: %w", i+1, err)
		}
	}

	if err := s.txManager.Commit(tx); err != nil {
		return nil, errors.New("ошибка сохранения теста")
	}

	return &dto.CreateTestResponse{
		ID:      test.ID,
		Message: "Тест создан успешно",
	}, nil
}

func (s *testService) validateQuestionOptions(qType domain.QType, options domain.QuestionOptions) error {
	switch qType {
	case domain.SingleChoice, domain.MultipleChoice:
		if len(options.Choices) < 2 {
			return errors.New("вопрос с выбором должен содержать минимум 2 варианта ответа")
		}
		hasCorrect := false
		for _, choice := range options.Choices {
			if choice.IsTrue {
				hasCorrect = true
				break
			}
		}
		if !hasCorrect {
			return errors.New("должен быть хотя бы один правильный ответ")
		}
	case domain.Matching:
		if len(options.MatchingPairs) < 2 {
			return errors.New("вопрос на сопоставление должен содержать минимум 2 пары")
		}
	case domain.TextInput:
		if len(options.CorrectInput) == 0 {
			return errors.New("должен быть указан хотя бы один правильный ответ")
		}
	case domain.CorrectOrder:
		if len(options.Sequence) < 2 {
			return errors.New("вопрос на порядок должен содержать минимум 2 элемента")
		}
	}
	return nil
}

func (s *testService) GetTests() (*dto.TestsListResponse, error) {
	tests, err := s.testRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения тестов: %w", err)
	}

	response := &dto.TestsListResponse{
		Tests: make([]dto.TestResponse, len(tests)),
	}

	for i, t := range tests {
		creatorName := ""
		if t.User.ID != 0 {
			creatorName = fmt.Sprintf("%s %s", t.User.Name, t.User.Surname)
		}

		response.Tests[i] = dto.TestResponse{
			ID:          t.ID,
			CreatorID:   t.CreatorID,
			Title:       t.Title,
			Description: t.Description,
			CreatorName: creatorName,
		}
	}

	return response, nil
}

func (s *testService) GetTestByID(id uint) (*dto.TestDetailResponse, error) {
	test, err := s.testRepo.FindByIDWithQuestions(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("тест не найден")
		}
		return nil, fmt.Errorf("ошибка получения теста: %w", err)
	}

	response := &dto.TestDetailResponse{
		TestResponse: dto.TestResponse{
			ID:          test.ID,
			CreatorID:   test.CreatorID,
			Title:       test.Title,
			Description: test.Description,
		},
		Questions: make([]dto.QuestionResponse, len(test.Questions)),
	}

	if test.User.ID != 0 {
		response.CreatorName = fmt.Sprintf("%s %s", test.User.Name, test.User.Surname)
	}

	for i, q := range test.Questions {
		var options domain.QuestionOptions
		if err := json.Unmarshal(q.Options, &options); err != nil {
			return nil, fmt.Errorf("ошибка парсинга опций вопроса: %w", err)
		}

		response.Questions[i] = dto.QuestionResponse{
			ID:          q.ID,
			Text:        q.Text,
			Points:      q.Points,
			Type:        string(q.Type),
			OrderNumber: q.OrderNumber,
			Options:     options,
		}
	}

	return response, nil
}

func (s *testService) DeleteTest(testID uint) error {
	if err := s.testRepo.Delete(testID); err != nil {
		return fmt.Errorf("ошибка удаления теста: %w", err)
	}

	return nil
}
