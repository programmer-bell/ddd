package mongodb

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/programmer-bell/ddd/aggregate"
	"github.com/programmer-bell/ddd/domain/customer"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func newCustomer(t *testing.T) aggregate.Customer {
	t.Helper()

	c, err := aggregate.NewCustomer("Bell")
	if err != nil {
		t.Fatal(err)
	}

	return c
}

func cursorResponse(mt *mtest.T, docs ...bson.D) bson.D {
	return mtest.CreateCursorResponse(0, mt.DB.Name()+"."+mt.Coll.Name(), mtest.FirstBatch, docs...)
}

func TestMongoRepository_Get(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().
		ClientType(mtest.Mock).
		DatabaseName("ddd").
		CollectionName(CustomerCollection))

	c := newCustomer(t)

	mt.Run("customer exists", func(mt *mtest.T) {
		mt.AddMockResponses(cursorResponse(mt, bson.D{
			{Key: "_id", Value: c.GetID().String()},
			{Key: "name", Value: c.GetName()},
		}))

		repo := New(mt.Coll)
		got, err := repo.Get(c.GetID())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.GetID() != c.GetID() {
			t.Errorf("expected id %s, got %s", c.GetID(), got.GetID())
		}
		if got.GetName() != c.GetName() {
			t.Errorf("expected name %s, got %s", c.GetName(), got.GetName())
		}
	})

	mt.Run("customer does not exist", func(mt *mtest.T) {
		mt.AddMockResponses(cursorResponse(mt))

		repo := New(mt.Coll)
		_, err := repo.Get(c.GetID())
		if !errors.Is(err, customer.ErrCustomerNotFound) {
			t.Errorf("expected %v, got %v", customer.ErrCustomerNotFound, err)
		}
	})
}

func TestMongoRepository_Add(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().
		ClientType(mtest.Mock).
		DatabaseName("ddd").
		CollectionName(CustomerCollection))

	c := newCustomer(t)

	mt.Run("new customer", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		repo := New(mt.Coll)
		if err := repo.Add(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	mt.Run("duplicate customer", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Index:   0,
			Code:    11000,
			Message: "duplicate key",
		}))

		repo := New(mt.Coll)
		err := repo.Add(c)
		if !errors.Is(err, customer.ErrFailedToAddCustomer) {
			t.Errorf("expected %v, got %v", customer.ErrFailedToAddCustomer, err)
		}
	})
}

func TestMongoRepository_Update(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().
		ClientType(mtest.Mock).
		DatabaseName("ddd").
		CollectionName(CustomerCollection))

	c := newCustomer(t)

	mt.Run("existing customer", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 1},
		))

		repo := New(mt.Coll)
		if err := repo.Update(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	mt.Run("missing customer", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 0},
			bson.E{Key: "nModified", Value: 0},
		))

		repo := New(mt.Coll)
		err := repo.Update(c)
		if !errors.Is(err, customer.ErrUpdateCustomer) {
			t.Errorf("expected %v, got %v", customer.ErrUpdateCustomer, err)
		}
	})
}

func TestMongoRepository_RoundTripID(t *testing.T) {
	id := uuid.New()
	c := newCustomer(t)
	c.SetID(id)

	doc := toDocument(c)
	if doc.ID != id.String() {
		t.Errorf("expected document id %s, got %s", id, doc.ID)
	}

	got, err := fromDocument(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.GetID() != id {
		t.Errorf("expected customer id %s, got %s", id, got.GetID())
	}
}
