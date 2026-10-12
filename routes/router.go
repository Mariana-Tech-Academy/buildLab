package routes

import (
	"net/http"

	"Fhub/auth"
	"Fhub/middleware"
	"Fhub/utils"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func Router(db *gorm.DB, tokens *auth.JWTManager) *mux.Router {
	r := mux.NewRouter()
	blacklist := middleware.NewTokenBlacklist()

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		utils.Error(w, http.StatusNotFound, "route not found")
	})
	r.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		utils.Error(w, http.StatusMethodNotAllowed, "method not allowed")
	})

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		utils.Success(w, http.StatusOK, "ok", nil)
	}).Methods(http.MethodGet)

	UserRoutes(api, db, tokens, blacklist)
	RestaurantRoutes(api, db, tokens, blacklist)
	MenuRoutes(api, db, tokens, blacklist)
	CategoryRoutes(api, db, tokens, blacklist)
	CartRoutes(api, db, tokens, blacklist)
	OrderRoutes(api, db, tokens, blacklist)
	AdminRoutes(api, db, tokens, blacklist)

	return r
}

// protected returns a subrouter whose routes require a valid JWT.
func protected(router *mux.Router, tokens *auth.JWTManager, blacklist *middleware.TokenBlacklist) *mux.Router {
	sub := router.NewRoute().Subrouter()
	sub.Use(middleware.Authenticate(tokens, blacklist))
	return sub
}
