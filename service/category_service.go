package service

import "Fhub/repository"

type CategoryService struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) *CategoryService {
	return &CategoryService{categoryRepo: repo}
}
