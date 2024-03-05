package repository

import (
	"AuthService/data"

	"gorm.io/gorm"
)

type VerificationTokenRepository struct {
	DB *gorm.DB
}

func NewVerificationTokenRepository(db *gorm.DB) *VerificationTokenRepository {
	return &VerificationTokenRepository{
		DB: db,
	}
}

func (r *VerificationTokenRepository) FindByToken(token string) (*data.VerificationToken, error) {
	var verificationToken data.VerificationToken
	err := r.DB.Where("token = ?", token).First(&verificationToken).Error
	if err != nil {
		return nil, err
	}
	return &verificationToken, nil
}

func (r *VerificationTokenRepository) Save(token *data.VerificationToken) error {
	return r.DB.Save(token).Error
}
