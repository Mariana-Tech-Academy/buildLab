package service

import "Fhub/repository"

type AdminService struct {
	AdminRepo repository.AdminRepository

}

func NewAdminService(repo repository.AdminRepository) *AdminService {
  return  &AdminService{AdminRepo:repo}

}