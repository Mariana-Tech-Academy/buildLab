package service

import "Fhub/repository"

type MenuService struct {
	menuRepo repository.MenuRepository
}

func NewMenuService(repo repository.MenuRepository) *MenuService {
	return &MenuService{menuRepo: repo}
}
