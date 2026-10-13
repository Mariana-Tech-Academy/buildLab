// package handlers

// import (
// 	"Fhub/service"
// 	"net/http"
// )

// type UserHandler struct {
// 	UserService service.UserService
// }

// func NewUserHandler(service service.UserService) *UserHandler {
// 	return &UserHandler{UserService: service}

// }

package handlers

import (
	"encoding/json"
	"net/http"
	"Fhub/auth"
	"Fhub/service"
	"Fhub/utils"

	
	
)

type UserHandler struct {
	UserService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{UserService: userService}
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.UserService.GetUserByID(int(claims.UserID))
	if err != nil {
		utils.Error(w, http.StatusNotFound, "User not found")
		return
	}
	profile := struct {
		ID        uint   `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Phone     string `json:"phone"`
		Role      string `json:"role"`
	}{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Phone:     user.Phone,
		Role:      string(user.Role),
	}

	utils.Success(
		w,
		http.StatusOK,
		"Current user retrieved successfully",
		profile,
	)
}
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var request struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Phone     string `json:"phone"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.UserService.UpdateProfile(    // Call the service layer to update the profile
		r.Context(),
		int64(claims.UserID),
		request.FirstName,
		request.LastName,
		request.Phone,
	)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	profile := struct {              // Create a response struct to return the updated profile
		ID        uint   `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Phone     string `json:"phone"`
		Role      string `json:"role"`
	}{
		ID:        user.ID,     // Populate the response struct with the updated user data
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Phone:     user.Phone,
		Role:      string(user.Role),
	}

	utils.Success(
		w,
		http.StatusOK,
		"Profile updated successfully",
		profile,
	)
}








