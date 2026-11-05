package valueobject

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	amount   int
	from     uuid.UUID
	to       uuid.UUID
	createAt time.Time
}

func NewTransaction(amount int, from, to uuid.UUID) Transaction {
	return Transaction{
		amount:   amount,
		from:     from,
		to:       to,
		createAt: time.Now(),
	}
}

func (t Transaction) Amount() int {
	return t.amount
}

func (t Transaction) From() uuid.UUID {
	return t.from
}

func (t Transaction) To() uuid.UUID {
	return t.to
}

func (t Transaction) CreateAt() time.Time {
	return t.createAt
}
