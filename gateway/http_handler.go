package main

import (
	"errors"
	common "github.com/vietquan-37/go-microservice/commons"
	pb "github.com/vietquan-37/go-microservice/commons/api"
	"github.com/vietquan-37/go-microservice/gateway/gateway"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"net/http"
)

type handler struct {
	gateway gateway.OrdersGateway
}

func NewHandler(ordersGateway gateway.OrdersGateway) *handler {
	return &handler{
		gateway: ordersGateway,
	}
}
func (h *handler) registerRoutes(mux *http.ServeMux) {
	mux.Handle("/", http.FileServer(http.Dir("public")))
	mux.HandleFunc("POST /api/customer/{customerID}/orders", h.HandleCreateOrder)
	mux.HandleFunc("GET /api/customers/{customerID}/orders/{orderID}", h.HandleGetOrder)
}
func (h *handler) HandleCreateOrder(w http.ResponseWriter, r *http.Request) {
	customerID := r.PathValue("customerID")
	var items []*pb.Items
	if err := common.ReadJson(r, &items); err != nil {
		common.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	err := validateItems(items)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, err.Error())
	}
	o, err := h.gateway.CreateOrder(r.Context(), &pb.CreateOrderRequest{
		CustomerID: customerID,
		Items:      items,
	})
	rStatus := status.Convert(err)
	if rStatus != nil {
		if rStatus.Code() != codes.InvalidArgument {
			common.WriteError(w, http.StatusBadRequest, rStatus.Message())
			return
		}
		common.WriteError(w, http.StatusInternalServerError, rStatus.Message())
		return
	}

	common.WriteJson(w, http.StatusOK, o)
}
func (h *handler) HandleGetOrder(w http.ResponseWriter, r *http.Request) {
	customerID := r.PathValue("customerID")
	orderID := r.PathValue("orderID")
	o, err := h.gateway.GetOrder(r.Context(), &pb.GetOrderRequest{
		CustomerID: customerID,
		OrderID:    orderID,
	})
	rStatus := status.Convert(err)
	if rStatus != nil {
		if rStatus.Code() != codes.InvalidArgument {
			common.WriteError(w, http.StatusBadRequest, rStatus.Message())
			return
		}
		common.WriteError(w, http.StatusInternalServerError, rStatus.Message())
		return
	}

	common.WriteJson(w, http.StatusOK, o)
}
func validateItems(items []*pb.Items) error {
	if len(items) == 0 {
		return common.ErrNoItems
	}
	for _, item := range items {
		if item.Quantity <= 0 {
			return errors.New("quantity must greater than zero")
		}
		if item.ID == "" {
			return errors.New("id must have a value")
		}
	}
	return nil
}
