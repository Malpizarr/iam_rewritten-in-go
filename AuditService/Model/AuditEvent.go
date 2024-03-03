package model

import (
	"time"
)

type AuditEvent struct {
	ID            uint      `gorm:"primaryKey"`
	EventType     string    `gorm:"size:255"`
	Username      string    `gorm:"size:255"`
	EventDateTime time.Time `gorm:"type:timestamp"`
	Details       string    `gorm:"type:text"`
	IpAddress     string    `gorm:"size:255"`
}

func (AuditEvent) TableName() string {
	return "audit_event"
}
