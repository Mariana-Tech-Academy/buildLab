package service

import "Fhub/repository"

type CategoryService interface {
}

type categoryService struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) CategoryService {
	return &categoryService{categoryRepo: repo}
}
