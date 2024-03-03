package repository

import (
	model "AuditService/Model"
)

type AuditEventRepository interface {
	Save(event *model.AuditEvent) error
}
