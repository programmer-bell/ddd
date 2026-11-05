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

type TavernConfiguration func(ts *TavernService) error

type TavernService struct {
	customers customer.CustomerRepository
	products  product.ProductRepository
}

func NewTavernService(cfgs ...TavernConfiguration) (*TavernService, error) {
	ts := &TavernService{}
	var err error

	for _, cfg := range cfgs {
		err = cfg(ts)
		if err != nil {
			return nil, err
		}
	}

	return ts, nil
}

func WithTavernCustomerRepository(cr customer.CustomerRepository) TavernConfiguration {
	return func(ts *TavernService) error {
		ts.customers = cr
		return nil
	}
}

func WithMemoryTavernCustomerRepository() TavernConfiguration {
	cr := memory.New()
	return WithTavernCustomerRepository(cr)
}

func WithTavernProductRepository(products []aggregate.Product) TavernConfiguration {
	return func(ts *TavernService) error {
		pr := prodmem.New()

		for _, p := range products {
			if err := pr.Add(p); err != nil {
				return err
			}
		}

		ts.products = pr
		return nil
	}
}

func (t *TavernService) Order(customerID uuid.UUID, productsID []uuid.UUID) error {
	c, err := t.customers.Get(customerID)
	if err != nil {
		return err
	}

	var products []aggregate.Product
	var total float64

	for _, id := range productsID {
		p, err := t.products.GetByID(id)

		if err != nil {
			return err
		}

		products = append(products, p)
		total += p.GetPrice()
	}

	log.Printf("customer %s has ordered %d products for total %f", c.GetID(), len(products), total)
	return nil
}
