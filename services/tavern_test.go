package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/programmer-bell/ddd/aggregate"
)

func TestTavern_NewTavernService(t *testing.T) {
	products := init_products(t)

	ts, err := NewTavernService(
		WithMemoryTavernCustomerRepository(),
		WithTavernProductRepository(products),
	)

	if err != nil {
		t.Fatal(err)
	}

	cust, err := aggregate.NewCustomer("Bell")
	if err != nil {
		t.Error(err)
	}

	err = ts.customers.Add(cust)
	if err != nil {
		t.Error(err)
	}

	order := []uuid.UUID{
		products[0].GetID(),
	}

	err = ts.Order(cust.GetID(), order)

	if err != nil {
		t.Error(err)
	}
}
