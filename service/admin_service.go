package service

import "Fhub/repository"

type AdminService struct {
	adminRepo repository.AdminRepository
}

func NewAdminService(repo repository.AdminRepository) *AdminService {
	return &AdminService{adminRepo: repo}
}
