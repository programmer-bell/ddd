package services

import (
	"log"

	"github.com/google/uuid"
	"github.com/programmer-bell/ddd/aggregate"
	"github.com/programmer-bell/ddd/domain/customer"
	"github.com/programmer-bell/ddd/domain/customer/memory"
	"github.com/programmer-bell/ddd/domain/product"
	prodmem "github.com/programmer-bell/ddd/domain/product/memory"
)

type OrderConfiguration func(os *OrderService) error

type OrderService struct {
	customers customer.CustomerRepository
	products  product.ProductRepository
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

func WithMemoryProductRepository(products []aggregate.Product) OrderConfiguration {
	return func(os *OrderService) error {
		pr := prodmem.New()

		for _, p := range products {
			if err := pr.Add(p); err != nil {
				return err
			}
		}

		os.products = pr
		return nil
	}
}

func (o *OrderService) CreateOrder(customerID uuid.UUID, productsID []uuid.UUID) error {
	c, err := o.customers.Get(customerID)
	if err != nil {
		return err
	}

	var products []aggregate.Product
	var total float64

	for _, id := range productsID {
		p, err := o.products.GetByID(id)

		if err != nil {
			return err
		}

		products = append(products, p)
		total += p.GetPrice()
	}

	log.Printf("customer %s has ordered %d products", c.GetID(), len(products))
	return nil
}
