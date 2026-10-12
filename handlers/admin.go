package handlers

import "Fhub/service"

type AdminHandler struct {
	AdminService *service.AdminService
}

func NewAdminHandler(service *service.AdminService) *AdminHandler {
	return &AdminHandler{AdminService: service}

}
