package handlers

import (
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"Fhub/auth"
	"Fhub/utils"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const MaxRestaurantImageSize = 5 * 1024 * 1024

var ErrInvalidImage = errors.New("invalid restaurant image")

func validateRestaurantImage(file io.ReadSeeker) (string, error) {
	_, format, err := image.DecodeConfig(file)
	if err != nil {
		return "", ErrInvalidImage
	}

	extensions := map[string]string{
		"jpeg": ".jpg",
		"png":  ".png",
	}
	extension, allowed := extensions[format]
	if !allowed {
		return "", ErrInvalidImage
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	return extension, nil
}

func (h *RestaurantHandler) UploadRestaurantImage(w http.ResponseWriter, r *http.Request) {
	restaurantID, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil || restaurantID == 0 {
		utils.Error(w, http.StatusBadRequest, "invalid restaurant ID")
		return
	}

	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.RestaurantService.CheckOwnership(r.Context(), uint(restaurantID), claims.UserID, claims.Role); err != nil {
		writeRestaurantError(w, err)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, MaxRestaurantImageSize+(1<<20))
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid or oversized upload")
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "image is required")
		return
	}
	defer file.Close()

	extension, err := validateRestaurantImage(file)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid image type or size")
		return
	}

	uploadDir := filepath.Join("uploads", "restaurants")
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		utils.Error(w, http.StatusInternalServerError, "cannot create upload folder")
		return
	}

	filename := uuid.NewString() + extension
	imagePath := filepath.Join(uploadDir, filename)

	destination, err := os.OpenFile(imagePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "cannot save image")
		return
	}

	// Enforce the size limit while writing; do not trust multipart header.Size.
	written, copyErr := io.Copy(destination, io.LimitReader(file, MaxRestaurantImageSize+1))
	closeErr := destination.Close()
	if copyErr != nil || closeErr != nil || written > MaxRestaurantImageSize {
		os.Remove(imagePath)
		if written > MaxRestaurantImageSize {
			utils.Error(w, http.StatusBadRequest, "invalid image type or size")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "image upload failed")
		return
	}

	imageURL := fmt.Sprintf("/uploads/restaurants/%s", filename)

	previousURL, err := h.RestaurantService.UpdateRestaurantImage(
		r.Context(),
		uint(restaurantID),
		claims.UserID,
		claims.Role,
		imageURL,
	)
	if err != nil {
		os.Remove(imagePath)
		writeRestaurantError(w, err)
		return
	}

	removeLocalRestaurantImage(previousURL)

	utils.Success(w, http.StatusOK, "restaurant image updated successfully", map[string]string{
		"image_url": imageURL,
	})
}

func removeLocalRestaurantImage(imageURL string) {
	if imageURL == "" || !strings.HasPrefix(imageURL, "/uploads/restaurants/") {
		return
	}
	path := filepath.Join(".", filepath.FromSlash(strings.TrimPrefix(imageURL, "/")))
	_ = os.Remove(path)
}
