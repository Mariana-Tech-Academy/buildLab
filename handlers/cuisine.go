package handlers

import "Fhub/service"

type CuisineHandler struct {
	CuisineService service.CuisineService
}


func NewCuisineHandler( service service.CuisineService) *CuisineHandler {
      return &CuisineHandler{CuisineService:service}

}