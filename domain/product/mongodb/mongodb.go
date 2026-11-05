package mongodb

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/programmer-bell/ddd/aggregate"
	"github.com/programmer-bell/ddd/domain/product"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const ProductCollection = "products"

// productDocument is the MongoDB representation of the Product aggregate.
type productDocument struct {
	ID          string  `bson:"_id"`
	Name        string  `bson:"name"`
	Description string  `bson:"description"`
	Price       float64 `bson:"price"`
	Quantity    int     `bson:"quantity"`
}

// MongoProductRepository stores products in a MongoDB collection.
type MongoProductRepository struct {
	collection *mongo.Collection
}

var _ product.ProductRepository = (*MongoProductRepository)(nil)

func New(collection *mongo.Collection) *MongoProductRepository {
	return &MongoProductRepository{collection: collection}
}

func (mpr *MongoProductRepository) GetAll() ([]aggregate.Product, error) {
	cursor, err := mpr.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, fmt.Errorf("mongodb: find products: %w", err)
	}
	defer cursor.Close(context.Background())

	var docs []productDocument
	if err := cursor.All(context.Background(), &docs); err != nil {
		return nil, fmt.Errorf("mongodb: decode products: %w", err)
	}

	products := make([]aggregate.Product, 0, len(docs))
	for _, doc := range docs {
		p, err := fromDocument(doc)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

func (mpr *MongoProductRepository) GetByID(id uuid.UUID) (aggregate.Product, error) {
	var doc productDocument

	err := mpr.collection.FindOne(context.Background(), bson.M{"_id": id.String()}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return aggregate.Product{}, product.ErrProductNotFound
	}
	if err != nil {
		return aggregate.Product{}, fmt.Errorf("mongodb: get product: %w", err)
	}

	return fromDocument(doc)
}

func (mpr *MongoProductRepository) Add(p aggregate.Product) error {
	_, err := mpr.collection.InsertOne(context.Background(), toDocument(p))
	if mongo.IsDuplicateKeyError(err) {
		return product.ErrProductAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("mongodb: add product: %w", err)
	}

	return nil
}

func (mpr *MongoProductRepository) Update(p aggregate.Product) error {
	result, err := mpr.collection.ReplaceOne(
		context.Background(),
		bson.M{"_id": p.GetID().String()},
		toDocument(p),
	)
	if err != nil {
		return fmt.Errorf("mongodb: update product: %w", err)
	}
	if result.MatchedCount == 0 {
		return product.ErrProductNotFound
	}

	return nil
}

func (mpr *MongoProductRepository) Delete(id uuid.UUID) error {
	result, err := mpr.collection.DeleteOne(context.Background(), bson.M{"_id": id.String()})
	if err != nil {
		return fmt.Errorf("mongodb: delete product: %w", err)
	}
	if result.DeletedCount == 0 {
		return product.ErrProductNotFound
	}

	return nil
}

func toDocument(p aggregate.Product) productDocument {
	item := p.GetItem()

	return productDocument{
		ID:          p.GetID().String(),
		Name:        item.Name,
		Description: item.Description,
		Price:       p.GetPrice(),
		Quantity:    p.GetQuantity(),
	}
}

func fromDocument(doc productDocument) (aggregate.Product, error) {
	id, err := uuid.Parse(doc.ID)
	if err != nil {
		return aggregate.Product{}, fmt.Errorf("mongodb: invalid product id %q: %w", doc.ID, err)
	}

	p, err := aggregate.NewProduct(doc.Name, doc.Description, doc.Price)
	if err != nil {
		return aggregate.Product{}, fmt.Errorf("mongodb: build product: %w", err)
	}
	p.SetID(id)
	p.SetQuantity(doc.Quantity)

	return p, nil
}
