package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Hiroki111/go-backend-example/internal/common"
	"github.com/Hiroki111/go-backend-example/internal/domain"
	"github.com/Hiroki111/go-backend-example/internal/repository"
	"github.com/Hiroki111/go-backend-example/internal/service"
)

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(UserIDKey).(uint)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error: "internal error",
		})
		return
	}

	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid request body",
		})
		return
	}

	ctx := r.Context()
	err := h.service.CreateOrder(ctx, userId, req.ProductID)
	if err != nil {
		if errors.Is(err, repository.ErrItemNotFound) {
			writeJSON(w, http.StatusNotFound, ErrorResponse{
				Error: "product not found",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error: "internal error",
		})
		return
	}

	writeJSON(w, http.StatusCreated, CreateOrderResponse{Message: "success"})
}

func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	orderBy := r.URL.Query().Get("orderBy")
	sortIn := r.URL.Query().Get("sortIn")
	productIDs := r.URL.Query().Get("product_ids")
	page := r.URL.Query().Get("page")
	limit := r.URL.Query().Get("limit")

	pageInt, err := parseOptionalInt(page, 1)
	if err != nil || pageInt <= 0 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "page must be a positive integer",
		})
		return
	}

	limitInt, err := parseOptionalInt(limit, common.DefaultPageLimit)
	if err != nil || limitInt <= 0 || limitInt > common.MaxPageLimit {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "limit must be a positive integer and not exceed " + strconv.Itoa(common.MaxPageLimit),
		})
		return
	}

	productIDsSlice := make([]string, 0)
	if productIDs != "" {
		productIDsSlice = strings.Split(productIDs, ",")
	}

	productIDIntegers := make([]uint, 0)
	for _, productIDString := range productIDsSlice {
		id, err := strconv.ParseUint(productIDString, 10, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{
				Error: "product_ids must be integers",
			})
			return
		}
		productIDIntegers = append(productIDIntegers, uint(id))
	}

	ctx := r.Context()
	params := service.GetOrderParameters{
		OrderBy:    orderBy,
		SortIn:     sortIn,
		ProductIDs: productIDIntegers,
		Page:       pageInt,
		Limit:      limitInt,
	}
	orders, total, err := h.service.GetOrdersWithTotalCount(ctx, params)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error: "failed to get orders",
		})
		return
	}
	items := make([]OrderItem, len(orders))
	for i, order := range orders {
		items[i] = OrderItem{
			ID:          order.ID,
			ProductName: order.Product.Name,
			PriceCents:  order.PriceCents,
		}
	}

	writeJSON(w, http.StatusOK, GetOrdersResponse{
		Items:   items,
		Page:    pageInt,
		Limit:   limitInt,
		Total:   int(total),
		HasNext: pageInt*limitInt < int(total),
	})
}

func (h *Handler) GetOrderById(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")

	id64, err := strconv.ParseInt(idString, 10, 64)
	if err != nil || id64 <= 0 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid ID",
		})
		return
	}
	id := uint(id64)

	role, roleRetrieved := r.Context().Value(RoleKey).(domain.UserRole)
	userId, userIdRetrieved := r.Context().Value(UserIDKey).(uint)
	if !roleRetrieved || !userIdRetrieved {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error: "internal error",
		})
		return
	}

	order, err := h.service.GetOrderById(id)
	if err != nil {
		if err == repository.ErrItemNotFound {
			writeJSON(w, http.StatusNotFound, ErrorResponse{
				Error: "item not found",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error: "internal error",
		})
		return
	}

	if role == domain.CustomerRole && order.UserID != userId {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error: "forbidden",
		})
		return
	}

	orderItem := OrderItem{
		ID:           order.ID,
		CustomerName: order.User.UserName,
		ProductName:  order.Product.Name,
		PriceCents:   order.PriceCents,
	}
	writeJSON(w, http.StatusOK, GetOrderResponse{
		Item: orderItem,
	})
}

func (h *Handler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")

	id64, err := strconv.ParseInt(idString, 10, 64)
	if err != nil || id64 <= 0 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid ID",
		})
		return
	}
	id := uint(id64)

	var payload UpdateOrderRequest

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid request body",
		})
		return
	}

	order, err := h.service.UpdateOrder(repository.UpdateOrderInput{ID: id, PriceCents: payload.PriceCents})
	if err != nil {
		if errors.Is(err, repository.ErrItemNotFound) {
			writeJSON(w, http.StatusNotFound, ErrorResponse{
				Error: "item not found",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error: "internal error",
		})
		return
	}

	item := OrderItem{
		ID:           order.ID,
		CustomerName: order.User.UserName,
		ProductName:  order.Product.Name,
		PriceCents:   order.PriceCents,
	}
	writeJSON(w, http.StatusOK, UpdateOrderResponse{Item: item})
}

func (h *Handler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")

	id64, err := strconv.ParseInt(idString, 10, 64)
	if err != nil || id64 <= 0 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid ID",
		})
		return
	}
	id := uint(id64)

	err = h.service.DeleteOrder(id)
	if err != nil {
		if errors.Is(err, repository.ErrItemNotFound) {
			writeJSON(w, http.StatusNotFound, ErrorResponse{
				Error: "order not found",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error: "internal error",
		})
		return
	}

	writeJSON(w, http.StatusOK, DeleteOrderResponse{Message: "success"})
}
