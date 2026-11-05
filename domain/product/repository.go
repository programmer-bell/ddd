package product

import (
	"github.com/google/uuid"
	"github.com/programmer-bell/ddd/aggregate"
)

type ProductRepository interface {
	GetAll() ([]aggregate.Product, error)
	GetByID(id uuid.UUID) (aggregate.Product, error)
	Add(product aggregate.Product) error
	Update(product aggregate.Product) error
	Delte(id uuid.UUID) error
}
