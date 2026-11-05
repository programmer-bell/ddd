package mongodb

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/programmer-bell/ddd/aggregate"
	"github.com/programmer-bell/ddd/domain/customer"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const CustomerCollection = "customers"

// customerDocument is the MongoDB representation of the Customer aggregate.
// The aggregate keeps its fields private, so the document is mapped explicitly.
type customerDocument struct {
	ID   string `bson:"_id"`
	Name string `bson:"name"`
}

// MongoRepository stores customers in a MongoDB collection.
type MongoRepository struct {
	collection *mongo.Collection
}

var _ customer.CustomerRepository = (*MongoRepository)(nil)

func New(collection *mongo.Collection) *MongoRepository {
	return &MongoRepository{collection: collection}
}

func (mr *MongoRepository) Get(id uuid.UUID) (aggregate.Customer, error) {
	var doc customerDocument

	err := mr.collection.FindOne(context.Background(), bson.M{"_id": id.String()}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return aggregate.Customer{}, customer.ErrCustomerNotFound
	}
	if err != nil {
		return aggregate.Customer{}, fmt.Errorf("mongodb: get customer: %w", err)
	}

	return fromDocument(doc)
}

func (mr *MongoRepository) Add(c aggregate.Customer) error {
	_, err := mr.collection.InsertOne(context.Background(), toDocument(c))
	if mongo.IsDuplicateKeyError(err) {
		return fmt.Errorf("mongodb: customer already exists: %w", customer.ErrFailedToAddCustomer)
	}
	if err != nil {
		return fmt.Errorf("mongodb: add customer: %w", err)
	}

	return nil
}

func (mr *MongoRepository) Update(c aggregate.Customer) error {
	result, err := mr.collection.ReplaceOne(
		context.Background(),
		bson.M{"_id": c.GetID().String()},
		toDocument(c),
	)
	if err != nil {
		return fmt.Errorf("mongodb: update customer: %w", err)
	}
	if result.MatchedCount == 0 {
		return customer.ErrUpdateCustomer
	}

	return nil
}

func toDocument(c aggregate.Customer) customerDocument {
	return customerDocument{
		ID:   c.GetID().String(),
		Name: c.GetName(),
	}
}

func fromDocument(doc customerDocument) (aggregate.Customer, error) {
	id, err := uuid.Parse(doc.ID)
	if err != nil {
		return aggregate.Customer{}, fmt.Errorf("mongodb: invalid customer id %q: %w", doc.ID, err)
	}

	c, err := aggregate.NewCustomer(doc.Name)
	if err != nil {
		return aggregate.Customer{}, fmt.Errorf("mongodb: build customer: %w", err)
	}
	c.SetID(id)

	return c, nil
}
