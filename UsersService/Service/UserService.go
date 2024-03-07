package Service

import (
	"UsersService/Repositories"
	"UsersService/Util"
	"UsersService/model"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"regexp"
)

type UserService struct {
	UserRepo Repositories.UserRepository
	RoleRepo Repositories.RoleRepository
}

func NewUserService(userRepo Repositories.UserRepository, roleRepo Repositories.RoleRepository) *UserService {
	return &UserService{
		UserRepo: userRepo,
		RoleRepo: roleRepo,
	}
}

func (s *UserService) CreateUser(newUser *model.GORMUser) (*model.GORMUser, error) {
	db := s.UserRepo.GetDB()

	err := db.Transaction(func(tx *gorm.DB) error {
		userRole, err := s.RoleRepo.FindByName("ROLE_USER")
		if err != nil {
			return err
		}

		newUser.Roles = append(newUser.Roles, *userRole)

		if _, err := s.UserRepo.FindByUsername(newUser.Username); err == nil {
			return errors.New("username already taken")
		}

		if _, err := s.UserRepo.FindByEmail(newUser.Email); err == nil {
			return errors.New("email already taken")
		}

		matched, _ := regexp.MatchString(`^[a-zA-Z0-9_+&*-]+(?:\.[a-zA-Z0-9_+&*-]+)*@(?:[a-zA-Z0-9-]+\.)+[a-zA-Z]{2,7}$`, newUser.Email)
		if !matched {
			return errors.New("invalid email format")
		}

		newUser.IsTwoFaEnabled = false
		newUser.IsEmailVerified = false
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
		newUser.Password = string(hashedPassword)

		if err := tx.Save(newUser).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return newUser, nil
}

func (s *UserService) SetVerified(user *model.GORMUser) error {
	user.IsEmailVerified = true
	_, err := s.UserRepo.Save(user)
	return err
}

func (s *UserService) GetUser(user *model.GORMUser) (*model.GORMUser, error) {
	user1, err := s.UserRepo.FindByUsername(user.Username)
	if err != nil {
		user1, err = s.UserRepo.FindByEmail(user.Email)
		if err != nil {
			return nil, errors.New("user not found")
		}
	}

	if user1.IsTwoFaEnabled {
		return nil, errors.New("2FA verification required")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user1.Password), []byte(user.Password))
	if err != nil {
		return nil, errors.New("invalid password")
	}

	return user1, nil
}

func (s *UserService) FindOrCreateUser(email, name, providerName, sub string) (*model.GORMUser, error) {
	existingUser, err := s.UserRepo.FindByEmail(email)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if existingUser != nil {
		providerExists := false
		for _, provider := range existingUser.OAuthProvider {
			if provider.ProviderName == providerName {
				providerExists = true
				break
			}
		}

		if !providerExists {
			newProvider := model.GORMOAuthProvider{
				ID:           sub + providerName,
				ProviderID:   sub,
				ProviderName: providerName,
				User:         existingUser,
			}
			existingUser.OAuthProvider = append(existingUser.OAuthProvider, newProvider)
			_, err := s.UserRepo.Save(existingUser)
			if err != nil {
				return nil, err
			}
		}

		if existingUser.IsTwoFaEnabled {
			return nil, errors.New("2FA verification required")
		}

		return existingUser, nil
	}

	generatedPassword := Util.GeneratePassword(10)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(generatedPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	newUser := &model.GORMUser{
		Email:    email,
		Username: name,
		Password: string(hashedPassword),
	}
	newUser.ID = uuid.New().String()

	newProvider := model.GORMOAuthProvider{
		ID:           sub + providerName,
		ProviderID:   sub,
		ProviderName: providerName,
		User:         newUser,
	}
	newUser.OAuthProvider = append(newUser.OAuthProvider, newProvider)

	createdUser, err := s.UserRepo.Create(newUser)
	if err != nil {
		return nil, err
	}

	return createdUser, nil
}

func (s *UserService) ConvertToDto(user *model.GORMUser) *model.OauthDTO {
	dto := &model.OauthDTO{
		Email: user.Email,
		Name:  user.Username,
	}

	if len(user.OAuthProvider) > 0 {
		dto.Provider = user.OAuthProvider[0].ProviderName
		dto.Sub = user.OAuthProvider[0].ID
	}

	return dto
}

func (s *UserService) UpdateUser(user *model.GORMUser) (*model.GORMUser, error) {
	result, err := s.UserRepo.Save(user)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *UserService) GetUserById(id string) (*model.GORMUser, error) {
	user, err := s.UserRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (s *UserService) GetUserByUsername(username string) (*model.GORMUser, error) {
	user, err := s.UserRepo.FindByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (s *UserService) DeleteUser(user *model.GORMUser) error {
	return s.UserRepo.Delete(user)
}

func (s *UserService) GetAllUsers() ([]*model.GORMUser, error) {
	return s.UserRepo.FindAll()
}
