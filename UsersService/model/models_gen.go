package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GORMOAuthProvider struct {
	ID           string    `gorm:"type:varchar(36);primaryKey;" json:"ID"`
	ProviderID   string    `json:"providerId"`
	ProviderName string    `json:"providerName"`
	UserID       string    `gorm:"type:varchar(255);"`
	User         *GORMUser `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
}

func (GORMOAuthProvider) TableName() string {
	return "oauth_provider"
}

type GORMRole struct {
	ID   int64  `gorm:"primaryKey;autoIncrement:true" json:"id"`
	Name string `gorm:"size:255" json:"name"`
}

func (GORMRole) TableName() string {
	return "role"
}

type GORMUser struct {
	ID              string              `json:"id" gorm:"primaryKey;size:36;"`
	Username        string              `json:"username" gorm:"unique;size:50;not null"`
	Email           string              `json:"email" gorm:"unique;not null"`
	Password        string              `json:"password" gorm:"not null"`
	TotpSecret      *string             `json:"totpSecret,omitempty"`
	IsTwoFaEnabled  bool                `json:"isTwoFaEnabled" gorm:"default:false"`
	IsEmailVerified bool                `json:"isEmailVerified" gorm:"default:false"`
	Roles           []GORMRole          `json:"roles,omitempty" gorm:"many2many:user_roles;"`
	OAuthProvider   []GORMOAuthProvider `json:"providers,omitempty" gorm:"foreignKey:UserID"`
}

func (GORMUser) TableName() string {
	return "user"
}

func (u *GORMUser) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return
}

type Mutation struct {
}

type Query struct {
}
