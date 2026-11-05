package product

import (
	"errors"

	"github.com/google/uuid"
	"github.com/programmer-bell/ddd/aggregate"
)

var (
	ErrProductNotFound      = errors.New("products not found")
	ErrProductAlreadyExists = errors.New("there is already such products")
)

type ProductRepository interface {
	GetAll() ([]aggregate.Product, error)
	GetByID(id uuid.UUID) (aggregate.Product, error)
	Add(product aggregate.Product) error
	Update(product aggregate.Product) error
	Delete(id uuid.UUID) error
}
