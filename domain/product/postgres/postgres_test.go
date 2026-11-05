package postgres

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/programmer-bell/ddd/aggregate"
	"github.com/programmer-bell/ddd/domain/product"
)

func newTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	return db, mock
}

func newProducts(t *testing.T) []aggregate.Product {
	t.Helper()

	beer, err := aggregate.NewProduct("Beer", "Healthy Bear", 199)
	if err != nil {
		t.Fatal(err)
	}
	beer.SetQuantity(10)

	peanuts, err := aggregate.NewProduct("Peanuts", "Snacks", 0.98)
	if err != nil {
		t.Fatal(err)
	}

	return []aggregate.Product{beer, peanuts}
}

func productRows(products []aggregate.Product) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"id", "name", "description", "price", "quantity"})

	for _, p := range products {
		item := p.GetItem()
		rows.AddRow(p.GetID().String(), item.Name, item.Description, p.GetPrice(), p.GetQuantity())
	}

	return rows
}

func TestPostgresProductRepository_GetAll(t *testing.T) {
	products := newProducts(t)

	t.Run("products exist", func(t *testing.T) {
		db, mock := newTestDB(t)
		mock.ExpectQuery(getAllProductsQuery).WillReturnRows(productRows(products))

		got, err := New(db).GetAll()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != len(products) {
			t.Fatalf("expected %d products, got %d", len(products), len(got))
		}
		if got[0].GetQuantity() != 10 {
			t.Errorf("expected quantity 10, got %d", got[0].GetQuantity())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})

	t.Run("no products", func(t *testing.T) {
		db, mock := newTestDB(t)
		mock.ExpectQuery(getAllProductsQuery).WillReturnRows(productRows(nil))

		got, err := New(db).GetAll()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("expected no products, got %d", len(got))
		}
	})
}

func TestPostgresProductRepository_GetByID(t *testing.T) {
	products := newProducts(t)
	beer := products[0]

	t.Run("product exists", func(t *testing.T) {
		db, mock := newTestDB(t)
		mock.ExpectQuery(getProductQuery).
			WithArgs(beer.GetID().String()).
			WillReturnRows(productRows([]aggregate.Product{beer}))

		got, err := New(db).GetByID(beer.GetID())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.GetID() != beer.GetID() {
			t.Errorf("expected id %s, got %s", beer.GetID(), got.GetID())
		}
		if got.GetPrice() != beer.GetPrice() {
			t.Errorf("expected price %f, got %f", beer.GetPrice(), got.GetPrice())
		}
	})

	t.Run("product does not exist", func(t *testing.T) {
		db, mock := newTestDB(t)
		mock.ExpectQuery(getProductQuery).
			WithArgs(beer.GetID().String()).
			WillReturnError(sql.ErrNoRows)

		_, err := New(db).GetByID(beer.GetID())
		if !errors.Is(err, product.ErrProductNotFound) {
			t.Errorf("expected %v, got %v", product.ErrProductNotFound, err)
		}
	})
}

func TestPostgresProductRepository_Add(t *testing.T) {
	beer := newProducts(t)[0]

	t.Run("new product", func(t *testing.T) {
		db, mock := newTestDB(t)
		mock.ExpectExec(insertProductQuery).
			WithArgs(
				beer.GetID().String(),
				beer.GetItem().Name,
				beer.GetItem().Description,
				beer.GetPrice(),
				beer.GetQuantity(),
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		if err := New(db).Add(beer); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})

	t.Run("duplicate product", func(t *testing.T) {
		db, mock := newTestDB(t)
		mock.ExpectExec(insertProductQuery).
			WithArgs(
				beer.GetID().String(),
				beer.GetItem().Name,
				beer.GetItem().Description,
				beer.GetPrice(),
				beer.GetQuantity(),
			).
			WillReturnError(&pq.Error{Code: "23505"})

		err := New(db).Add(beer)
		if !errors.Is(err, product.ErrProductAlreadyExists) {
			t.Errorf("expected %v, got %v", product.ErrProductAlreadyExists, err)
		}
	})
}

func TestPostgresProductRepository_Update(t *testing.T) {
	beer := newProducts(t)[0]

	t.Run("existing product", func(t *testing.T) {
		db, mock := newTestDB(t)
		mock.ExpectExec(updateProductQuery).
			WithArgs(
				beer.GetItem().Name,
				beer.GetItem().Description,
				beer.GetPrice(),
				beer.GetQuantity(),
				beer.GetID().String(),
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		if err := New(db).Update(beer); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing product", func(t *testing.T) {
		db, mock := newTestDB(t)
		mock.ExpectExec(updateProductQuery).
			WithArgs(
				beer.GetItem().Name,
				beer.GetItem().Description,
				beer.GetPrice(),
				beer.GetQuantity(),
				beer.GetID().String(),
			).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := New(db).Update(beer)
		if !errors.Is(err, product.ErrProductNotFound) {
			t.Errorf("expected %v, got %v", product.ErrProductNotFound, err)
		}
	})
}

func TestPostgresProductRepository_Delete(t *testing.T) {
	beer := newProducts(t)[0]

	t.Run("existing product", func(t *testing.T) {
		db, mock := newTestDB(t)
		mock.ExpectExec(deleteProductQuery).
			WithArgs(beer.GetID().String()).
			WillReturnResult(sqlmock.NewResult(0, 1))

		if err := New(db).Delete(beer.GetID()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing product", func(t *testing.T) {
		db, mock := newTestDB(t)
		mock.ExpectExec(deleteProductQuery).
			WithArgs(beer.GetID().String()).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := New(db).Delete(beer.GetID())
		if !errors.Is(err, product.ErrProductNotFound) {
			t.Errorf("expected %v, got %v", product.ErrProductNotFound, err)
		}
	})
}

func TestPostgresProductRepository_RoundTripID(t *testing.T) {
	id := uuid.New()
	beer := newProducts(t)[0]
	beer.SetID(id)

	db, mock := newTestDB(t)
	mock.ExpectQuery(getProductQuery).
		WithArgs(id.String()).
		WillReturnRows(productRows([]aggregate.Product{beer}))

	got, err := New(db).GetByID(id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.GetID() != id {
		t.Errorf("expected product id %s, got %s", id, got.GetID())
	}
}
