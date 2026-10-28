package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"Fhub/auth"
	"Fhub/models"
	"Fhub/service"
	"Fhub/utils"

	"github.com/gorilla/mux"
)

// OrderHandler handles HTTP requests related to orders.
type OrderHandler struct {
	OrderService service.OrderService
}

// orderCustomerResponse contains the customer's basic details.
type orderCustomerResponse struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

// orderItemResponse contains the details of each ordered item.
type orderItemResponse struct {
	Name      string `json:"name"`
	UnitPrice int64  `json:"unit_price"`
	Quantity  int    `json:"quantity"`
	LineTotal int64  `json:"line_total"`
}

// restaurantOrderResponse contains the information shown to the manager.
type restaurantOrderResponse struct {
	ID                   uint                  `json:"id"`
	OrderNumber          string                `json:"order_number"`
	Status               models.OrderStatus    `json:"status"`
	CreatedAt            time.Time             `json:"created_at"`
	Customer             orderCustomerResponse `json:"customer"`
	Items                []orderItemResponse   `json:"items"`
	Subtotal             int64                 `json:"subtotal"`
	DeliveryFee          int64                 `json:"delivery_fee"`
	Total                int64                 `json:"total"`
	DeliveryName         string                `json:"delivery_name"`
	DeliveryPhone        string                `json:"delivery_phone"`
	DeliveryAddress      string                `json:"delivery_address"`
	DeliveryCity         string                `json:"delivery_city"`
	DeliveryState        string                `json:"delivery_state"`
	DeliveryInstructions string                `json:"delivery_instructions"`
}

// newRestaurantOrderResponse converts an order into a safe JSON response.
func newRestaurantOrderResponse(order models.Order) restaurantOrderResponse {

	// Create an empty slice to store the order items.
	items := make([]orderItemResponse, 0, len(order.Items))

	// Loop through each order item; _ ignores the index, and item holds the current item.
	for _, item := range order.Items {

		// Add the item's name, price, quantity and total to the response.
		items = append(items, orderItemResponse{
			Name:      item.Name,
			UnitPrice: item.UnitPrice,
			Quantity:  item.Quantity,
			LineTotal: item.LineTotal,
		})
	}

	// Return only the information the restaurant manager needs.
	return restaurantOrderResponse{

		// Basic order information.
		ID:          order.ID,
		OrderNumber: order.OrderNumber,
		Status:      order.Status,
		CreatedAt:   order.CreatedAt,

		// Customer details without exposing sensitive account information.
		Customer: orderCustomerResponse{
			FirstName: order.Customer.FirstName,
			LastName:  order.Customer.LastName,
			Phone:     order.Customer.Phone,
		},

		// Ordered items and payment totals.
		Items:       items,
		Subtotal:    order.Subtotal,
		DeliveryFee: order.DeliveryFee,
		Total:       order.Total,

		// Delivery information provided during checkout.
		DeliveryName:         order.DeliveryName,
		DeliveryPhone:        order.DeliveryPhone,
		DeliveryAddress:      order.DeliveryAddress,
		DeliveryCity:         order.DeliveryCity,
		DeliveryState:        order.DeliveryState,
		DeliveryInstructions: order.DeliveryInstructions,
	}
}

// NewOrderHandler creates a new order handler and connects it to the order service.
func NewOrderHandler(service service.OrderService) *OrderHandler {
	return &OrderHandler{
		OrderService: service,
	}
}

// ValidateDeliveryInfo receives delivery information from the customer,
// checks that the required fields are provided,
// and returns a JSON response.
func (h *OrderHandler) ValidateDeliveryInfo(w http.ResponseWriter, r *http.Request) {

	// Create a variable to store the customer's delivery information.
	var input service.DeliveryInfoInput

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	validationErrors := h.OrderService.ValidateDeliveryInfo(&input)
	if len(validationErrors) > 0 {
		utils.JSON(w, http.StatusUnprocessableEntity, utils.Response{
			Success: false,
			Message: "delivery information is invalid",
			Data:    validationErrors,
		})
		return
	}

	// Return HTTP 200 OK with a success message and the validated delivery details.
	utils.Success(w, http.StatusOK, "delivery information is valid", input)
}

// GetRestaurantOrders retrieves orders belonging to the manager's restaurant.
func (h *OrderHandler) GetRestaurantOrders(w http.ResponseWriter, r *http.Request) {

	// Get the authenticated manager's details from the JWT.
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get the restaurant ID from the URL.
	// mux.Vars gets the restaurant ID from the URL's {id} parameter.
	restaurantID, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil || restaurantID == 0 {
		utils.Error(w, http.StatusBadRequest, "invalid restaurant ID")
		return
	}

	status := r.URL.Query().Get("status")
	page, limit, message := restaurantListPage(r)
	if message != "" {
		utils.Error(w, http.StatusBadRequest, message)
		return
	}

	result, err := h.OrderService.GetRestaurantOrders(r.Context(), claims.UserID, claims.Role, uint(restaurantID), status, page, limit)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidOrderStatus), errors.Is(err, service.ErrInvalidOrderPage):
			utils.Error(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrOrderAccessDenied):
			utils.Error(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrRestaurantNotFound):
			utils.Error(w, http.StatusNotFound, "restaurant not found")
		default:
			log.Printf("get restaurant orders: %v", err)
			utils.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	items := make([]restaurantOrderResponse, 0, len(result.Orders))
	for _, order := range result.Orders {
		items = append(items, newRestaurantOrderResponse(order))
	}

	utils.Success(w, http.StatusOK, "restaurant orders retrieved successfully", restaurantOrdersResponse{
		Items: items,
		Page:  result.Page,
		Limit: result.Limit,
		Total: result.Total,
	})
}

type restaurantOrdersResponse struct {
	Items []restaurantOrderResponse `json:"items"`
	Page  int                       `json:"page"`
	Limit int                       `json:"limit"`
	Total int64                     `json:"total"`
}

func (h *OrderHandler) GetCheckoutSummary(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	summary, err := h.OrderService.GetCheckoutSummary(
		r.Context(),
		claims.UserID,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmptyCart):
			utils.Error(w, http.StatusUnprocessableEntity, "cannot checkout with an empty cart")
		case errors.Is(err, service.ErrUnavailableItem):
			utils.Error(w, http.StatusConflict, "cart contains an unavailable item")
		case errors.Is(err, service.ErrRestaurantNotFound):
			utils.Error(w, http.StatusNotFound, "restaurant not found")
		default:
			log.Printf("checkout summary: %v", err)
			utils.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	utils.Success(w, http.StatusOK, "checkout summary", summary)
}
