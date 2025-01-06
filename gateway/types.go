package main

import pb "github.com/vietquan-37/go-microservice/commons/api"

type CreateOrderRequest struct {
	Order         *pb.Order `json:"order"`
	RedirectToUrl string    `json:"redirectToUrl"`
}
