package handlers

import "Fhub/service"

type CategoryHandler struct {
	CategoryService service.CategoryService
}


func NewCategoryHandler( service service.CategoryService) *CategoryHandler {
      return &CategoryHandler{CategoryService:service}

}