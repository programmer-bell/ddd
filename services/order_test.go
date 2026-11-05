package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/programmer-bell/ddd/aggregate"
)

func init_products(t *testing.T) []aggregate.Product {
	beer, err := aggregate.NewProduct("Beer", "Healthy Bear", 199)
	if err != nil {
		t.Fatal(err)
	}

	peenuts, err := aggregate.NewProduct("Peanuts", "Snacks", 0.98)

	if err != nil {
		t.Fatal(err)
	}

	wine, err := aggregate.NewProduct("wine", "nasty drink", 0.99)
	if err != nil {
		t.Fatal(err)
	}

	return []aggregate.Product{
		beer, peenuts, wine,
	}

}

func TestOrder_NewOrderservice(t *testing.T) {
	products := init_products(t)

	os, err := NewOrderservice(
		WithMemorycustomerRepository(),
		WithMemoryProductRepository(products),
	)

	if err != nil {
		t.Fatal(err)
	}

	cust, err := aggregate.NewCustomer("Bell")
	if err != nil {
		t.Error(err)
	}

	err = os.customers.Add(cust)
	if err != nil {
		t.Error(err)
	}

	order := []uuid.UUID{
		products[0].GetID(),
	}

	err = os.CreateOrder(cust.GetID(), order)

	if err != nil {
		t.Error(err)
	}
}
