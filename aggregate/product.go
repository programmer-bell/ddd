package aggregate

import (
	"errors"

	"github.com/google/uuid"
	"github.com/programmer-bell/ddd/entity"
)

var (
	ErrMissingValue = errors.New("missing important values")
)

type Product struct {
	item     *entity.Item
	price    float64
	quantity int
}

func NewProduct(name, description string, price float64) (Product, error) {
	if name == "" || description == "" {
		return Product{}, ErrMissingValue
	}
	return Product{
		item: &entity.Item{
			ID:          uuid.New(),
			Name:        name,
			Description: description,
		},
		price:    price,
		quantity: 0,
	}, nil
}

func (p Product) GetID() uuid.UUID {
	return p.item.ID
}

func (p *Product) SetID(id uuid.UUID) {
	if p.item == nil {
		p.item = &entity.Item{}
	}
	p.item.ID = id
}

func (p Product) GetItem() *entity.Item {
	return p.item
}

func (p Product) GetPrice() float64 {
	return p.price
}

func (p *Product) SetPrice(price float64) {
	p.price = price
}

func (p Product) GetQuantity() int {
	return p.quantity
}

func (p *Product) SetQuantity(quantity int) {
	p.quantity = quantity
}
