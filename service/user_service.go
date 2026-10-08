package service

import "Fhub/repository"

type UserService struct {
	UserRepo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{UserRepo: repo}

}
