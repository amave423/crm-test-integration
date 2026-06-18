package repository

import (
	"test-constructor/internal/domain"

	"gorm.io/gorm"
)

type TestRepository interface {
	CreateWithTx(tx *gorm.DB, test *domain.Test) error
	FindByID(id uint) (*domain.Test, error)
	FindByIDWithQuestions(id uint) (*domain.Test, error)
	FindAll() ([]domain.Test, error)
	Delete(id uint) error
	DeleteWithTx(tx *gorm.DB, id uint) error
}

type testRepository struct {
	db *gorm.DB
}

func NewTestRepository(db *gorm.DB) TestRepository {
	return &testRepository{db: db}
}

func (r *testRepository) CreateWithTx(tx *gorm.DB, test *domain.Test) error {
	return tx.Create(test).Error
}

func (r *testRepository) FindByID(id uint) (*domain.Test, error) {
	var test domain.Test
	err := r.db.Preload("User").First(&test, id).Error
	if err != nil {
		return nil, err
	}
	return &test, nil
}

func (r *testRepository) FindByIDWithQuestions(id uint) (*domain.Test, error) {
	var test domain.Test
	err := r.db.Preload("Questions").Preload("User").First(&test, id).Error
	if err != nil {
		return nil, err
	}
	return &test, nil
}

func (r *testRepository) FindAll() ([]domain.Test, error) {
	var tests []domain.Test
	err := r.db.Preload("User").Find(&tests).Error
	return tests, err
}

func (r *testRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Test{}, id).Error
}

func (r *testRepository) DeleteWithTx(tx *gorm.DB, id uint) error {
	return tx.Delete(&domain.Test{}, id).Error
}
