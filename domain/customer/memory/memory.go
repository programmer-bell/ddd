package memory

import (
	"sync"

	"github.com/google/uuid"
	"github.com/programmer-bell/ddd/aggregate"
)

type MemoryRepository struct {
	customers map[uuid.UUID]aggregate.Customer
	sync.Mutex
}

func New() *MemoryRepository {
	return &MemoryRepository{
		customers: make(map[uuid.UUID]aggregate.Customer),
	}
}

func (mr *MemoryRepository) Get(uuid.UUID) (aggregate.Customer, error) {
	return aggregate.Customer{}, nil
}

func (mr *MemoryRepository) Add(aggregate.Customer) error {
	return nil
}

func (mr *MemoryRepository) Update(aggregate.Customer) error {
	return nil
}
