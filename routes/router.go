package routes

import (
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func Router(db *gorm.DB) *mux.Router {
	r := mux.NewRouter()

	UserRoutes(r, db)
	RestaurantRoutes(r, db)
	CuisineRoutes(r, db)
	MenuRoutes(r, db)
	CategoryRoutes(r, db)
	CartRoutes(r, db)
	OrderRoutes(r, db)
	AdminRoutes(r, db)

	return r
}