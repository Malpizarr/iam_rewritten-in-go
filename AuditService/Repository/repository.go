package repository

import (
	model "AuditService/Model"
	"gorm.io/gorm"
)

type AuditEventRepositoryImpl struct {
	DB *gorm.DB
}

func NewAuditEventRepository(db *gorm.DB) *AuditEventRepositoryImpl {
	return &AuditEventRepositoryImpl{DB: db}
}

func (repo *AuditEventRepositoryImpl) Save(event *model.AuditEvent) error {
	return repo.DB.Create(event).Error
}
