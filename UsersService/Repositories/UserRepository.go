package Repositories

import (
	"UsersService/model"
	"gorm.io/gorm"
)

type UserRepository interface {
	FindByID(id string) (*model.GORMUser, error)
	FindByEmail(email string) (*model.GORMUser, error)
	FindByUsername(username string) (*model.GORMUser, error)
	Save(user *model.GORMUser) (*model.GORMUser, error)

	Create(user *model.GORMUser) (*model.GORMUser, error)

	Delete(user *model.GORMUser) error

	FindAll() ([]*model.GORMUser, error)

	GetDB() *gorm.DB
}

type GormUserRepository struct {
	DB *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{DB: db}
}

func (repo *GormUserRepository) FindByID(id string) (*model.GORMUser, error) {
	var user model.GORMUser
	if err := repo.DB.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (repo *GormUserRepository) GetDB() *gorm.DB {
	return repo.DB
}

func (repo *GormUserRepository) FindByEmail(email string) (*model.GORMUser, error) {
	var user model.GORMUser
	if err := repo.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (repo *GormUserRepository) FindByUsername(username string) (*model.GORMUser, error) {
	var user model.GORMUser
	if err := repo.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (repo *GormUserRepository) Create(user *model.GORMUser) (*model.GORMUser, error) {
	if err := repo.DB.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (repo *GormUserRepository) Save(user *model.GORMUser) (*model.GORMUser, error) {
	if err := repo.DB.Save(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (repo *GormUserRepository) Delete(user *model.GORMUser) error {
	return repo.DB.Delete(user).Error
}

func (repo *GormUserRepository) FindAll() ([]*model.GORMUser, error) {
	var users []*model.GORMUser
	if err := repo.DB.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
