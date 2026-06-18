package repository

import (
	"test-constructor/internal/domain"

	"gorm.io/gorm"
)

type QuestionRepository interface {
	CreateWithTx(tx *gorm.DB, question *domain.Question) error
	CreateBatchWithTx(tx *gorm.DB, questions []domain.Question) error
	FindByTestID(testID uint) ([]domain.Question, error)
	DeleteByTestID(testID uint) error
	DeleteByQuestionID(questionID uint) error
	GetMaxScoreByTestID(testID uint) (int, error)
}

type questionRepository struct {
	db *gorm.DB
}

func NewQuestionRepository(db *gorm.DB) QuestionRepository {
	return &questionRepository{db: db}
}

func (r *questionRepository) CreateWithTx(tx *gorm.DB, question *domain.Question) error {
	return tx.Create(question).Error
}

func (r *questionRepository) CreateBatchWithTx(tx *gorm.DB, questions []domain.Question) error {
	return tx.Create(&questions).Error
}

func (r *questionRepository) FindByTestID(testID uint) ([]domain.Question, error) {
	var questions []domain.Question
	err := r.db.Where("test_id = ?", testID).Order("order_number asc").Find(&questions).Error
	return questions, err
}

func (r *questionRepository) DeleteByTestID(testID uint) error {
	return r.db.Where("test_id = ?", testID).Delete(&domain.Question{}).Error
}

func (r *questionRepository) DeleteByQuestionID(questionID uint) error {
	return r.db.Delete(&domain.Question{}, questionID).Error
}

func (r *questionRepository) GetMaxScoreByTestID(testID uint) (int, error) {
	var maxScore int
	err := r.db.Model(&domain.Question{}).
		Where("test_id = ?", testID).
		Select("COALESCE(SUM(points), 0)").
		Scan(&maxScore).Error

	if err != nil {
		return 0, err
	}

	return maxScore, nil
}
