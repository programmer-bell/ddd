package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/programmer-bell/ddd/aggregate"
	"github.com/programmer-bell/ddd/domain/product"
)

const (
	getAllProductsQuery = "SELECT id, name, description, price, quantity FROM products"
	getProductQuery     = "SELECT id, name, description, price, quantity FROM products WHERE id = $1"
	insertProductQuery  = "INSERT INTO products (id, name, description, price, quantity) VALUES ($1, $2, $3, $4, $5)"
	updateProductQuery  = "UPDATE products SET name = $1, description = $2, price = $3, quantity = $4 WHERE id = $5"
	deleteProductQuery  = "DELETE FROM products WHERE id = $1"
)

// PostgresProductRepository stores products in PostgreSQL.
type PostgresProductRepository struct {
	db *sql.DB
}

var _ product.ProductRepository = (*PostgresProductRepository)(nil)

func New(db *sql.DB) *PostgresProductRepository {
	return &PostgresProductRepository{db: db}
}

func (pr *PostgresProductRepository) GetAll() ([]aggregate.Product, error) {
	rows, err := pr.db.QueryContext(context.Background(), getAllProductsQuery)
	if err != nil {
		return nil, fmt.Errorf("postgres: find products: %w", err)
	}
	defer rows.Close()

	products := make([]aggregate.Product, 0)
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: read products: %w", err)
	}

	return products, nil
}

func (pr *PostgresProductRepository) GetByID(id uuid.UUID) (aggregate.Product, error) {
	row := pr.db.QueryRowContext(context.Background(), getProductQuery, id.String())

	p, err := scanProduct(row)
	if errors.Is(err, sql.ErrNoRows) {
		return aggregate.Product{}, product.ErrProductNotFound
	}
	if err != nil {
		return aggregate.Product{}, err
	}

	return p, nil
}

func (pr *PostgresProductRepository) Add(p aggregate.Product) error {
	doc := toProductValues(p)

	_, err := pr.db.ExecContext(
		context.Background(),
		insertProductQuery,
		doc.id,
		doc.name,
		doc.description,
		doc.price,
		doc.quantity,
	)
	if isDuplicateKey(err) {
		return product.ErrProductAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("postgres: add product: %w", err)
	}

	return nil
}

func (pr *PostgresProductRepository) Update(p aggregate.Product) error {
	doc := toProductValues(p)

	result, err := pr.db.ExecContext(
		context.Background(),
		updateProductQuery,
		doc.name,
		doc.description,
		doc.price,
		doc.quantity,
		doc.id,
	)
	if err != nil {
		return fmt.Errorf("postgres: update product: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("postgres: update product rows affected: %w", err)
	}
	if affected == 0 {
		return product.ErrProductNotFound
	}

	return nil
}

func (pr *PostgresProductRepository) Delete(id uuid.UUID) error {
	result, err := pr.db.ExecContext(context.Background(), deleteProductQuery, id.String())
	if err != nil {
		return fmt.Errorf("postgres: delete product: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("postgres: delete product rows affected: %w", err)
	}
	if affected == 0 {
		return product.ErrProductNotFound
	}

	return nil
}

// scanner is implemented by both *sql.Row and *sql.Rows.
type scanner interface {
	Scan(dest ...interface{}) error
}

type productValues struct {
	id          string
	name        string
	description string
	price       float64
	quantity    int
}

func toProductValues(p aggregate.Product) productValues {
	item := p.GetItem()

	return productValues{
		id:          p.GetID().String(),
		name:        item.Name,
		description: item.Description,
		price:       p.GetPrice(),
		quantity:    p.GetQuantity(),
	}
}

func scanProduct(s scanner) (aggregate.Product, error) {
	var values productValues

	err := s.Scan(&values.id, &values.name, &values.description, &values.price, &values.quantity)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return aggregate.Product{}, err
		}
		return aggregate.Product{}, fmt.Errorf("postgres: scan product: %w", err)
	}

	id, err := uuid.Parse(values.id)
	if err != nil {
		return aggregate.Product{}, fmt.Errorf("postgres: invalid product id %q: %w", values.id, err)
	}

	p, err := aggregate.NewProduct(values.name, values.description, values.price)
	if err != nil {
		return aggregate.Product{}, fmt.Errorf("postgres: build product: %w", err)
	}
	p.SetID(id)
	p.SetQuantity(values.quantity)

	return p, nil
}

// isDuplicateKey reports whether PostgreSQL rejected the write because of a
// unique constraint violation.
func isDuplicateKey(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
