package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/programmer-bell/ddd/aggregate"
	"github.com/programmer-bell/ddd/domain/customer"
)

const (
	getCustomerQuery    = "SELECT name FROM customers WHERE id = $1"
	insertCustomerQuery = "INSERT INTO customers (id, name) VALUES ($1, $2)"
	updateCustomerQuery = "UPDATE customers SET name = $1 WHERE id = $2"
)

// PostgresRepository stores customers in PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

var _ customer.CustomerRepository = (*PostgresRepository)(nil)

func New(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (pr *PostgresRepository) Get(id uuid.UUID) (aggregate.Customer, error) {
	var name string

	err := pr.db.QueryRowContext(context.Background(), getCustomerQuery, id.String()).Scan(&name)
	if errors.Is(err, sql.ErrNoRows) {
		return aggregate.Customer{}, customer.ErrCustomerNotFound
	}
	if err != nil {
		return aggregate.Customer{}, fmt.Errorf("postgres: get customer: %w", err)
	}

	c, err := aggregate.NewCustomer(name)
	if err != nil {
		return aggregate.Customer{}, fmt.Errorf("postgres: build customer: %w", err)
	}
	c.SetID(id)

	return c, nil
}

func (pr *PostgresRepository) Add(c aggregate.Customer) error {
	_, err := pr.db.ExecContext(
		context.Background(),
		insertCustomerQuery,
		c.GetID().String(),
		c.GetName(),
	)
	if isDuplicateKey(err) {
		return fmt.Errorf("postgres: customer already exists: %w", customer.ErrFailedToAddCustomer)
	}
	if err != nil {
		return fmt.Errorf("postgres: add customer: %w", err)
	}

	return nil
}

func (pr *PostgresRepository) Update(c aggregate.Customer) error {
	result, err := pr.db.ExecContext(
		context.Background(),
		updateCustomerQuery,
		c.GetName(),
		c.GetID().String(),
	)
	if err != nil {
		return fmt.Errorf("postgres: update customer: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("postgres: update customer rows affected: %w", err)
	}
	if affected == 0 {
		return customer.ErrUpdateCustomer
	}

	return nil
}

// isDuplicateKey reports whether PostgreSQL rejected the write because of a
// unique constraint violation.
func isDuplicateKey(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
