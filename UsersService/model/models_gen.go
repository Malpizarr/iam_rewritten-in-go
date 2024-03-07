package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GORMOAuthProvider struct {
	ID           string    `gorm:"type:varchar(36);primaryKey;" json:"ID"`
	ProviderID   string    `json:"providerId"`
	ProviderName string    `json:"providerName"`
	UserID       string    `gorm:"type:varchar(36);null" json:"userID"`
	User         *GORMUser `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
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
	Roles           []GORMRole          `gorm:"many2many:user_roles;foreignKey:ID;joinForeignKey:UserID;References:ID;JoinReferences:RoleID"`
	OAuthProvider   []GORMOAuthProvider `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (GORMUser) TableName() string {
	return "user"
}

func (u *GORMUser) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		u.ID = uuid.NewString()
	}
	return
}

type Mutation struct {
}

type Query struct {
}
