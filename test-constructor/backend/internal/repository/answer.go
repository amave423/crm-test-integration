package repository

import (
	"test-constructor/internal/domain"

	"gorm.io/gorm"
)

type AnswerRepository interface {
	CreateBatch(answers []domain.Answer) error
	CreateBatchWithTx(tx *gorm.DB, answers []domain.Answer) error
}

type answerRepository struct {
	db *gorm.DB
}

func NewAnswerRepository(db *gorm.DB) AnswerRepository {
	return &answerRepository{db: db}
}

func (r *answerRepository) CreateBatch(answers []domain.Answer) error {
	return r.db.Create(&answers).Error
}

func (r *answerRepository) CreateBatchWithTx(tx *gorm.DB, answers []domain.Answer) error {
	return tx.Create(&answers).Error
}
