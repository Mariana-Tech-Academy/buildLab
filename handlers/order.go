package handlers

import (
    "encoding/json"
    "errors"
    "net/http"
    "strconv"

    "Fhub/auth"
    "Fhub/repository"
    "Fhub/service"
    "Fhub/utils"

    "github.com/gorilla/mux"
    "gorm.io/gorm"
)


type OrderHandler struct {
	OrderService service.OrderService
}

func NewOrderHandler(service service.OrderService) *OrderHandler {
	return &OrderHandler{OrderService: service}

}


func (h *OrderHandler) Checkout(
    w http.ResponseWriter,
    r *http.Request,
) {
    claims, err := auth.ClaimsFromContext(r.Context())
    if err != nil {
        utils.Error(w, 401, "unauthorized")
        return
    }

    var delivery repository.DeliverySnapshot

    err = json.NewDecoder(
        http.MaxBytesReader(w, r.Body, 1<<20),
    ).Decode(&delivery)

    if err != nil {
        utils.Error(w, 400, "invalid JSON")
        return
    }

    order, err := h.OrderService.Checkout(
        claims.UserID,
        delivery,
    )

    if err != nil {
        switch {
        case errors.Is(err, service.ErrInvalidDelivery),
            errors.Is(err, repository.ErrEmptyCart),
            errors.Is(err, repository.ErrUnavailableItem):
            utils.Error(w, 400, err.Error())
        default:
            utils.Error(w, 500, "checkout failed")
        }
        return
    }

    utils.Success(w, 201, "order placed", order)
}

func (h *OrderHandler) List(
    w http.ResponseWriter,
    r *http.Request,
) {
    claims, err := auth.ClaimsFromContext(r.Context())
    if err != nil {
        utils.Error(w, 401, "unauthorized")
        return
    }

    orders, err := h.OrderService.List(claims.UserID)
    if err != nil {
        utils.Error(w, 500, "could not load orders")
        return
    }

    utils.Success(w, 200, "orders", orders)
}


func (h *OrderHandler) Get(
    w http.ResponseWriter,
    r *http.Request,
) {
    claims, err := auth.ClaimsFromContext(r.Context())
    if err != nil {
        utils.Error(w, 401, "unauthorized")
        return
    }

    id, err := strconv.ParseUint(
        mux.Vars(r)["id"], 10, 64,
    )
    if err != nil {
        utils.Error(w, 400, "invalid order ID")
        return
    }

    order, err := h.OrderService.Get(
        claims.UserID,
        uint(id),
    )

    if errors.Is(err, gorm.ErrRecordNotFound) {
        utils.Error(w, 404, "order not found")
        return
    }

    if err != nil {
        utils.Error(w, 500, "could not load order")
        return
    }

    utils.Success(w, 200, "order", order)
}
