

package service

import "Fhub/repository"

type CuisineService struct {
	CuisineRepo repository.CuisineRepository
}

func NewCuisineService(repo repository.CuisineRepository ) *CuisineService {
	return &CuisineService{CuisineRepo: repo}

}