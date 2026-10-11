
package service

import (
	"context"
	"errors"
	"strings"

	"Fhub/auth"
	"Fhub/models"
	"Fhub/repository"
)

type UserService interface {
	GetUserByID(id int) (models.User, error)
	UpdateProfile(
		ctx context.Context,
		id int64,
		firstName, lastName, phone string,
	) (*models.User, error)

	
}

type userService struct {
	userRepo repository.UserRepository
	tokens   *auth.JWTManager
}

func NewUserService(repo repository.UserRepository,tokens *auth.JWTManager,) UserService {
	return &userService{userRepo: repo,tokens:tokens,}
}

func (s *userService) GetUserByID(id int) (models.User, error) {
	return s.userRepo.GetUserByID(id)
}


func (s *userService) UpdateProfile(
	ctx context.Context,
	id int64,
	firstName, lastName, phone string,
) (*models.User, error) {
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)
	phone = strings.TrimSpace(phone)

	if firstName == "" || lastName == "" || phone == "" {
		return nil, errors.New("first name, last name and phone are required")
	}

	return s.userRepo.UpdateProfile(
		ctx, id, firstName, lastName, phone,
	)
}

   