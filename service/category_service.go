package service

import "Fhub/repository"

type CategoryService struct {
	CategoryRepo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) *CategoryService {
	return &CategoryService{CategoryRepo: repo}

}