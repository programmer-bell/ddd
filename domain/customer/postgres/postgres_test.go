package postgres

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/programmer-bell/ddd/aggregate"
	"github.com/programmer-bell/ddd/domain/customer"
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

func newCustomer(t *testing.T) aggregate.Customer {
	t.Helper()

	c, err := aggregate.NewCustomer("Bell")
	if err != nil {
		t.Fatal(err)
	}

	return c
}

func TestPostgresRepository_Get(t *testing.T) {
	c := newCustomer(t)

	t.Run("customer exists", func(t *testing.T) {
		db, mock := newTestDB(t)
		mock.ExpectQuery(getCustomerQuery).
			WithArgs(c.GetID().String()).
			WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow(c.GetName()))

		got, err := New(db).Get(c.GetID())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.GetID() != c.GetID() {
			t.Errorf("expected id %s, got %s", c.GetID(), got.GetID())
		}
		if got.GetName() != c.GetName() {
			t.Errorf("expected name %s, got %s", c.GetName(), got.GetName())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})

	t.Run("customer does not exist", func(t *testing.T) {
		db, mock := newTestDB(t)
		mock.ExpectQuery(getCustomerQuery).
			WithArgs(c.GetID().String()).
			WillReturnError(sql.ErrNoRows)

		_, err := New(db).Get(c.GetID())
		if !errors.Is(err, customer.ErrCustomerNotFound) {
			t.Errorf("expected %v, got %v", customer.ErrCustomerNotFound, err)
		}
	})
}

func TestPostgresRepository_Add(t *testing.T) {
	c := newCustomer(t)

	t.Run("new customer", func(t *testing.T) {
		db, mock := newTestDB(t)
		mock.ExpectExec(insertCustomerQuery).
			WithArgs(c.GetID().String(), c.GetName()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		if err := New(db).Add(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})

	t.Run("duplicate customer", func(t *testing.T) {
		db, mock := newTestDB(t)
		mock.ExpectExec(insertCustomerQuery).
			WithArgs(c.GetID().String(), c.GetName()).
			WillReturnError(&pq.Error{Code: "23505"})

		err := New(db).Add(c)
		if !errors.Is(err, customer.ErrFailedToAddCustomer) {
			t.Errorf("expected %v, got %v", customer.ErrFailedToAddCustomer, err)
		}
	})
}

func TestPostgresRepository_Update(t *testing.T) {
	c := newCustomer(t)

	t.Run("existing customer", func(t *testing.T) {
		db, mock := newTestDB(t)
		mock.ExpectExec(updateCustomerQuery).
			WithArgs(c.GetName(), c.GetID().String()).
			WillReturnResult(sqlmock.NewResult(0, 1))

		if err := New(db).Update(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing customer", func(t *testing.T) {
		db, mock := newTestDB(t)
		mock.ExpectExec(updateCustomerQuery).
			WithArgs(c.GetName(), c.GetID().String()).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := New(db).Update(c)
		if !errors.Is(err, customer.ErrUpdateCustomer) {
			t.Errorf("expected %v, got %v", customer.ErrUpdateCustomer, err)
		}
	})
}

func TestPostgresRepository_RoundTripID(t *testing.T) {
	id := uuid.New()
	c := newCustomer(t)
	c.SetID(id)

	db, mock := newTestDB(t)
	mock.ExpectQuery(getCustomerQuery).
		WithArgs(id.String()).
		WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow(c.GetName()))

	got, err := New(db).Get(id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.GetID() != id {
		t.Errorf("expected customer id %s, got %s", id, got.GetID())
	}
}
