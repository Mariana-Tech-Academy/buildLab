package service

import (
	"Fhub/models"
	"Fhub/repository"
)


type MenuService interface {


Get(uint) (*models.MenuItem, error)
    List(uint) ([]models.MenuItem, error)
    SetAvailability(uint, uint, bool, bool) (*models.MenuItem, error)
    CheckAvailable(uint) error
    Search(filters repository.MenuSearchFilters) ([]models.MenuItem, error)
}

type menuService struct {
	menuRepo repository.MenuRepository
}

func NewMenuService(repo repository.MenuRepository) MenuService {
	return &menuService{menuRepo: repo}
}
 
func (s *menuService) Get(id uint) (*models.MenuItem, error) {
    return s.menuRepo.Get(id)
}

func (s *menuService) List(id uint) ([]models.MenuItem, error) {
    return s.menuRepo.List(id)
}

func (s *menuService) SetAvailability(
    id, managerID uint,
    isAdmin, available bool,
) (*models.MenuItem, error) {
    return s.menuRepo.SetAvailability(
        id, managerID, isAdmin, available,
    )
}

func (s *menuService) CheckAvailable(id uint) error {
    return s.menuRepo.CheckAvailable(id)
}

func (s *menuService) Search(
    filters repository.MenuSearchFilters,
) ([]models.MenuItem, error) {
    return s.menuRepo.Search(filters)
}
