package service

import "Fhub/repository"

type MenuService struct {
	MenuRepo repository.MenuRepository
}

func NewMenuService(repo repository.MenuRepository) *MenuService {
	return &MenuService{MenuRepo: repo}
}