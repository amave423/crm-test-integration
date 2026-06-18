package repository

import "gorm.io/gorm"

type TransactionManager interface {
	Begin() (*gorm.DB, error)
	Commit(tx *gorm.DB) error
	Rollback(tx *gorm.DB) error
}

type transactionManager struct {
	db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) TransactionManager {
	return &transactionManager{db: db}
}

func (tm *transactionManager) Begin() (*gorm.DB, error) {
	tx := tm.db.Begin()
	return tx, tx.Error
}

func (tm *transactionManager) Commit(tx *gorm.DB) error {
	return tx.Commit().Error
}

func (tm *transactionManager) Rollback(tx *gorm.DB) error {
	return tx.Rollback().Error
}
