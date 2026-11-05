package customer

import (
	"errors"

	"github.com/google/uuid"
	"github.com/programmer-bell/ddd/aggregate"
)

var (
	ErrCustomerNotFound    = errors.New("the customer not found")
	ErrFailedToAddCustomer = errors.New("failed to add customer")
	ErrUpdateCustomer      = errors.New("failed to update the customer")
)

type CustomerRepository interface {
	Get(uuid.UUID) (aggregate.Customer, error)
	Add(aggregate.Customer) error
	Update(aggregate.Customer) error
}
