
package handlers

import (
    "errors"
    "image"
    _ "image/jpeg"
    _ "image/png"
    "io"
    "mime/multipart"
	"Fhub/auth"
"Fhub/service"
"Fhub/utils"

"fmt"
"net/http"
"os"
"path/filepath"
"strconv"

"github.com/gorilla/mux"
"github.com/google/uuid"
)

const MaxRestaurantImageSize = 5 * 1024 * 1024

var ErrInvalidImage = errors.New("invalid restaurant image")

func validateRestaurantImage(
    file multipart.File,
    size int64,
) (string, error) {

    if size <= 0 || size > MaxRestaurantImageSize {
        return "", ErrInvalidImage
    }

    // Check the actual image format.
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

    // Return the file pointer to the beginning.
    if _, err := file.Seek(0, io.SeekStart); err != nil {
        return "", err
    }

    return extension, nil
}

func (h *RestaurantHandler) UploadRestaurantImage(
    w http.ResponseWriter,
    r *http.Request,
) {
    // Step 1: Get restaurant ID.
    restaurantID, err := strconv.ParseUint(
        mux.Vars(r)["id"], 10, 64,
    )
    if err != nil || restaurantID == 0 {
        utils.Error(w, http.StatusBadRequest, "invalid restaurant ID")
        return
    }

    // Step 2: Get authenticated manager.
    claims, err := auth.ClaimsFromContext(r.Context())
    if err != nil {
        utils.Error(w, http.StatusUnauthorized, "unauthorized")
        return
    }

    managerID := claims.UserID

    // Step 3: Check restaurant ownership.
    err = h.RestaurantService.CheckOwnership(
        uint(restaurantID),
        managerID,
    )
    if err != nil {
        switch {
        case errors.Is(err, service.ErrRestaurantForbidden):
            utils.Error(w, http.StatusForbidden, "not your restaurant")
        case errors.Is(err, service.ErrRestaurantNotFound):
            utils.Error(w, http.StatusNotFound, "restaurant not found")
        default:
            utils.Error(w, http.StatusInternalServerError, "ownership check failed")
        }
        return
    }

    // Step 4: Limit the total request body.
    r.Body = http.MaxBytesReader(
        w, r.Body, MaxRestaurantImageSize+(1<<20),
    )

    if err := r.ParseMultipartForm(1 << 20); err != nil {
        utils.Error(w, http.StatusBadRequest, "invalid or oversized upload")
        return
    }
    if r.MultipartForm != nil {
        defer r.MultipartForm.RemoveAll()
    }

    file, header, err := r.FormFile("image")
    if err != nil {
        utils.Error(w, http.StatusBadRequest, "image is required")
        return
    }
    defer file.Close()

    extension, err := validateRestaurantImage(file, header.Size)
    if err != nil {
        utils.Error(w, http.StatusBadRequest, "invalid image type or size")
        return
    }

    // Step 5: Save the image.
    uploadDir := filepath.Join("uploads", "restaurants")

    if err := os.MkdirAll(uploadDir, 0755); err != nil {
        utils.Error(w, http.StatusInternalServerError, "cannot create upload folder")
        return
    }

    filename := uuid.NewString() + extension
    imagePath := filepath.Join(uploadDir, filename)

    destination, err := os.OpenFile(
        imagePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644,
    )
    if err != nil {
        utils.Error(w, http.StatusInternalServerError, "cannot save image")
        return
    }

    _, copyErr := io.Copy(destination, file)
    closeErr := destination.Close()

    if copyErr != nil || closeErr != nil {
        os.Remove(imagePath)
        utils.Error(w, http.StatusInternalServerError, "image upload failed")
        return
    }

    
imageURL := fmt.Sprintf("/uploads/restaurants/%s", filename)

// Save the image URL to PostgreSQL.
err = h.RestaurantService.UpdateRestaurantImage(
    uint(restaurantID),
    managerID,
    imageURL,
)

if err != nil {
    // Remove the uploaded file if the database update fails.
    os.Remove(imagePath)

    switch {
    case errors.Is(err, service.ErrRestaurantForbidden):
        utils.Error(w, http.StatusForbidden, "not your restaurant")

    case errors.Is(err, service.ErrRestaurantNotFound):
        utils.Error(w, http.StatusNotFound, "restaurant not found")

    default:
        utils.Error(w, http.StatusInternalServerError, "failed to update restaurant image")
    }
    return
}

// Image uploaded and database updated successfully.
utils.Success(
    w,
    http.StatusOK,
    "restaurant image updated successfully",
    map[string]string{
        "image_url": imageURL,
    },
)

}
