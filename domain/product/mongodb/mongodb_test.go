package mongodb

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/programmer-bell/ddd/aggregate"
	"github.com/programmer-bell/ddd/domain/product"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

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

func productDocuments(products []aggregate.Product) []bson.D {
	docs := make([]bson.D, 0, len(products))

	for _, p := range products {
		item := p.GetItem()
		docs = append(docs, bson.D{
			{Key: "_id", Value: p.GetID().String()},
			{Key: "name", Value: item.Name},
			{Key: "description", Value: item.Description},
			{Key: "price", Value: p.GetPrice()},
			{Key: "quantity", Value: p.GetQuantity()},
		})
	}

	return docs
}

func TestMongoProductRepository_GetAll(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().
		ClientType(mtest.Mock).
		DatabaseName("ddd").
		CollectionName(ProductCollection))

	products := newProducts(t)

	mt.Run("products exist", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(
			0,
			mt.DB.Name()+"."+mt.Coll.Name(),
			mtest.FirstBatch,
			productDocuments(products)...,
		))

		repo := New(mt.Coll)
		got, err := repo.GetAll()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != len(products) {
			t.Fatalf("expected %d products, got %d", len(products), len(got))
		}
		if got[0].GetQuantity() != 10 {
			t.Errorf("expected quantity 10, got %d", got[0].GetQuantity())
		}
	})

	mt.Run("no products", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(
			0,
			mt.DB.Name()+"."+mt.Coll.Name(),
			mtest.FirstBatch,
		))

		repo := New(mt.Coll)
		got, err := repo.GetAll()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("expected no products, got %d", len(got))
		}
	})
}

func TestMongoProductRepository_GetByID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().
		ClientType(mtest.Mock).
		DatabaseName("ddd").
		CollectionName(ProductCollection))

	products := newProducts(t)
	beer := products[0]
	doc := productDocuments(products)[0]

	mt.Run("product exists", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(
			0,
			mt.DB.Name()+"."+mt.Coll.Name(),
			mtest.FirstBatch,
			doc,
		))

		repo := New(mt.Coll)
		got, err := repo.GetByID(beer.GetID())
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

	mt.Run("product does not exist", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(
			0,
			mt.DB.Name()+"."+mt.Coll.Name(),
			mtest.FirstBatch,
		))

		repo := New(mt.Coll)
		_, err := repo.GetByID(uuid.New())
		if !errors.Is(err, product.ErrProductNotFound) {
			t.Errorf("expected %v, got %v", product.ErrProductNotFound, err)
		}
	})
}

func TestMongoProductRepository_Add(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().
		ClientType(mtest.Mock).
		DatabaseName("ddd").
		CollectionName(ProductCollection))

	p := newProducts(t)[0]

	mt.Run("new product", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		repo := New(mt.Coll)
		if err := repo.Add(p); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	mt.Run("duplicate product", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Index:   0,
			Code:    11000,
			Message: "duplicate key",
		}))

		repo := New(mt.Coll)
		err := repo.Add(p)
		if !errors.Is(err, product.ErrProductAlreadyExists) {
			t.Errorf("expected %v, got %v", product.ErrProductAlreadyExists, err)
		}
	})
}

func TestMongoProductRepository_Update(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().
		ClientType(mtest.Mock).
		DatabaseName("ddd").
		CollectionName(ProductCollection))

	p := newProducts(t)[0]

	mt.Run("existing product", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 1},
		))

		repo := New(mt.Coll)
		if err := repo.Update(p); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	mt.Run("missing product", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 0},
			bson.E{Key: "nModified", Value: 0},
		))

		repo := New(mt.Coll)
		err := repo.Update(p)
		if !errors.Is(err, product.ErrProductNotFound) {
			t.Errorf("expected %v, got %v", product.ErrProductNotFound, err)
		}
	})
}

func TestMongoProductRepository_Delete(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().
		ClientType(mtest.Mock).
		DatabaseName("ddd").
		CollectionName(ProductCollection))

	p := newProducts(t)[0]

	mt.Run("existing product", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
		))

		repo := New(mt.Coll)
		if err := repo.Delete(p.GetID()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	mt.Run("missing product", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 0},
		))

		repo := New(mt.Coll)
		err := repo.Delete(p.GetID())
		if !errors.Is(err, product.ErrProductNotFound) {
			t.Errorf("expected %v, got %v", product.ErrProductNotFound, err)
		}
	})
}
