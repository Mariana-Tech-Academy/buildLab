package handlers

import (
    "errors"
    "net/http"
    "strconv"

    "Fhub/auth"
    "Fhub/repository"
    "Fhub/service"
    "Fhub/utils"

    "github.com/gorilla/mux"
)

type CartHandler struct {
	CartService service.CartService
}

func NewCartHandler(service service.CartService) *CartHandler {
	return &CartHandler{CartService: service}

}

// CART-04: Remove an item from the cart
func (h *CartHandler) RemoveItem(
    w http.ResponseWriter,
    r *http.Request,
) {
    claims, err := auth.ClaimsFromContext(r.Context())

    if err != nil {
        utils.Error(
            w,
            http.StatusUnauthorized,
            "unauthenticated",
        )
        return
    }

    // Get cart item ID from URL
    id, err := strconv.ParseUint(
        mux.Vars(r)["id"],
        10,
        32,
    )

    if err != nil || id == 0 {
        utils.Error(
            w,
            http.StatusBadRequest,
            "invalid cart item id",
        )
        return
    }

    // Remove item using service
    result, err := h.CartService.RemoveItem(
        claims.UserID,
        uint(id),
    )

    if errors.Is(err, repository.ErrCartItemNotFound) {
        utils.Error(
            w,
            http.StatusNotFound,
            "cart item not found",
        )
        return
    }

    if err != nil {
        utils.Error(
            w,
            http.StatusInternalServerError,
            "could not remove cart item",
        )
        return
    }

    utils.Success(
        w,
        http.StatusOK,
        "cart item removed",
        result,
    )
}

// CART-04: Clear the entire cart
func (h *CartHandler) Clear(
    w http.ResponseWriter,
    r *http.Request,
) {
    claims, err := auth.ClaimsFromContext(r.Context())

    if err != nil {
        utils.Error(
            w,
            http.StatusUnauthorized,
            "unauthenticated",
        )
        return
    }

    // Clear cart using service
    result, err := h.CartService.Clear(claims.UserID)

    if err != nil {
        utils.Error(
            w,
            http.StatusInternalServerError,
            "could not clear cart",
        )
        return
    }

    utils.Success(
        w,
        http.StatusOK,
        "cart cleared",
        result,
    )
}
