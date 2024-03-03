package repository

import (
	model "AuditService/Model"
	"gorm.io/gorm"
)

type auditEventRepositoryImpl struct {
	DB *gorm.DB
}

func NewAuditEventRepository(db *gorm.DB) *auditEventRepositoryImpl {
	return &auditEventRepositoryImpl{DB: db}
}

func (repo *auditEventRepositoryImpl) Save(event *model.AuditEvent) error {
	return repo.DB.Create(event).Error
}
