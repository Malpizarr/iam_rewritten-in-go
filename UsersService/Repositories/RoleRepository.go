package Repositories

import (
	"UsersService/model"
	"gorm.io/gorm"
)

type RoleRepository interface {
	FindByName(name string) (*model.GORMRole, error)
}

type GormRoleRepository struct {
	DB *gorm.DB
}

func (repo *GormRoleRepository) FindByName(name string) (*model.GORMRole, error) {
	var role model.GORMRole
	if err := repo.DB.Where("name = ?", name).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}
