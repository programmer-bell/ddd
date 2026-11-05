package aggregate

import (
	"errors"

	"github.com/google/uuid"
	"github.com/programmer-bell/ddd/entity"
	"github.com/programmer-bell/ddd/valueobject"
)

var (
	ErrInvalidPerson = errors.New("Customer needs a valid name")
)

type Customer struct {
	person       *entity.Person
	products     []*entity.Item
	transactions []valueobject.Transaction
}

func NewCustomer(name string) (Customer, error) {
	if name == "" {
		return Customer{}, ErrInvalidPerson
	}

	person := &entity.Person{
		Name: name,
		ID:   uuid.New(),
	}

	return Customer{
		person:       person,
		products:     make([]*entity.Item, 0),
		transactions: make([]valueobject.Transaction, 0),
	}, nil
}
