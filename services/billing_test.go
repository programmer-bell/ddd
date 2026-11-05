package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/programmer-bell/ddd/aggregate"
)

func TestBilling_NewBillingService(t *testing.T) {
	products := init_products(t)

	bs, err := NewBillingService(
		WithMemoryBillingCustomerRepository(),
		WithBillingProductRepository(products),
	)

	if err != nil {
		t.Fatal(err)
	}

	cust, err := aggregate.NewCustomer("Bell")
	if err != nil {
		t.Error(err)
	}

	err = bs.customers.Add(cust)
	if err != nil {
		t.Error(err)
	}

	order := []uuid.UUID{
		products[0].GetID(),
	}

	tx, err := bs.Charge(cust.GetID(), order)

	if err != nil {
		t.Error(err)
	}

	if tx.Amount() <= 0 {
		t.Errorf("expected positive billing amount, got %d", tx.Amount())
	}

	if tx.From() != cust.GetID() {
		t.Errorf("expected transaction from %s, got %s", cust.GetID(), tx.From())
	}
}
