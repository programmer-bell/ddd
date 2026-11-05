package services

import (
	"log"

	"github.com/google/uuid"
	"github.com/programmer-bell/ddd/domain/customer"
	"github.com/programmer-bell/ddd/domain/customer/memory"
)

type OrderConfiguration func(os *OrderService) error

type OrderService struct {
	customers customer.CustomerRepository
}

func NewOrderservice(cfgs ...OrderConfiguration) (*OrderService, error) {
	os := &OrderService{}
	var err error

	for _, cfg := range cfgs {
		err = cfg(os)
		if err != nil {
			return nil, err
		}
	}

	return os, nil
}

func WithCustomerRepository(cr customer.CustomerRepository) OrderConfiguration {
	return func(os *OrderService) error {
		os.customers = cr
		return nil
	}
}

func WithMemorycustomerRepository() OrderConfiguration {
	cr := memory.New()
	return WithCustomerRepository(cr)
}

func (o *OrderService) CreateOrder(customerID uuid.UUID, productsID []uuid.UUID) error {
	c, err := o.customers.Get(customerID)
	if err != nil {
		return err
	}
	log.Println(c)
	return nil
}
