package services

import (
	"log"

	"github.com/google/uuid"
	"github.com/programmer-bell/ddd/aggregate"
	"github.com/programmer-bell/ddd/domain/customer"
	"github.com/programmer-bell/ddd/domain/customer/memory"
	"github.com/programmer-bell/ddd/domain/product"
	prodmem "github.com/programmer-bell/ddd/domain/product/memory"
	"github.com/programmer-bell/ddd/valueobject"
)

type BillingConfiguration func(bs *BillingService) error

type BillingService struct {
	customers customer.CustomerRepository
	products  product.ProductRepository
}

func NewBillingService(cfgs ...BillingConfiguration) (*BillingService, error) {
	bs := &BillingService{}
	var err error

	for _, cfg := range cfgs {
		err = cfg(bs)
		if err != nil {
			return nil, err
		}
	}

	return bs, nil
}

func WithBillingCustomerRepository(cr customer.CustomerRepository) BillingConfiguration {
	return func(bs *BillingService) error {
		bs.customers = cr
		return nil
	}
}

func WithMemoryBillingCustomerRepository() BillingConfiguration {
	cr := memory.New()
	return WithBillingCustomerRepository(cr)
}

func WithBillingProductRepository(products []aggregate.Product) BillingConfiguration {
	return func(bs *BillingService) error {
		pr := prodmem.New()

		for _, p := range products {
			if err := pr.Add(p); err != nil {
				return err
			}
		}

		bs.products = pr
		return nil
	}
}

func (b *BillingService) Charge(customerID uuid.UUID, productsID []uuid.UUID) (valueobject.Transaction, error) {
	c, err := b.customers.Get(customerID)
	if err != nil {
		return valueobject.Transaction{}, err
	}

	var products []aggregate.Product
	var total float64

	for _, id := range productsID {
		p, err := b.products.GetByID(id)

		if err != nil {
			return valueobject.Transaction{}, err
		}

		products = append(products, p)
		total += p.GetPrice()
	}

	tx := valueobject.NewTransaction(int(total), c.GetID(), uuid.Nil)

	log.Printf("customer %s has been billed %f for %d products", c.GetID(), total, len(products))
	return tx, nil
}
