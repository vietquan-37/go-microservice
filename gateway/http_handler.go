package main

import (
	"errors"
	"fmt"
	common "github.com/vietquan-37/go-microservice/commons"
	pb "github.com/vietquan-37/go-microservice/commons/api"
	"github.com/vietquan-37/go-microservice/gateway/gateway"
	"go.opentelemetry.io/otel"
	codes2 "go.opentelemetry.io/otel/codes"
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
	tr := otel.Tracer("http")
	ctx, span := tr.Start(r.Context(), fmt.Sprintf("%s %s", r.Method, r.RequestURI))
	defer span.End()
	err := validateItems(items)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	o, err := h.gateway.CreateOrder(ctx, &pb.CreateOrderRequest{
		CustomerID: customerID,
		Items:      items,
	})
	rStatus := status.Convert(err)
	if rStatus != nil {
		span.SetStatus(codes2.Error, err.Error())
		if rStatus.Code() == codes.InvalidArgument {
			common.WriteError(w, http.StatusBadRequest, rStatus.Message())
			return
		}
		common.WriteError(w, http.StatusInternalServerError, rStatus.Message())
		return
	}
	rsp := CreateOrderRequest{
		Order:         o,
		RedirectToUrl: fmt.Sprintf("http://%s/success.html?customerID=%s&orderID=%s", r.Host, customerID, o.ID),
	}

	common.WriteJson(w, http.StatusOK, rsp)
}
func (h *handler) HandleGetOrder(w http.ResponseWriter, r *http.Request) {
	customerID := r.PathValue("customerID")
	orderID := r.PathValue("orderID")
	tr := otel.Tracer("http")
	ctx, span := tr.Start(r.Context(), fmt.Sprintf("%s %s", r.Method, r.RequestURI))
	defer span.End()
	o, err := h.gateway.GetOrder(ctx, &pb.GetOrderRequest{
		CustomerID: customerID,
		OrderID:    orderID,
	})
	rStatus := status.Convert(err)
	if rStatus != nil {
		span.SetStatus(codes2.Error, err.Error())
		if rStatus.Code() == codes.InvalidArgument {
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
		if item.ID == 0 {
			return errors.New("id must have a value")
		}
	}
	return nil
}
