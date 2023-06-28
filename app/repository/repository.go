package repository

import (
	"context"
	"errors"

	"sme-api/app/entity"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Repository :
type Repository struct {
	db *mongo.Database
}

// New :
func New(ctx context.Context, mongo *mongo.Client) *Repository {
	return &Repository{
		db: mongo.Database("csi-local"),
	}
}

// Create : a generic function to create entity
func (r Repository) Create(entityName entity.Collection, entity interface{}) (*mongo.InsertOneResult, error) {
	insertResult, err := r.db.Collection(entityName).InsertOne(context.Background(), entity)
	if err != nil {
		return nil, err
	}
	return insertResult, nil
}

// Check : a generic function to check existence of the document in the entityName table, does not return the result
func (r Repository) Check(entityName entity.Collection, id interface{}) (bool, error) {
	count, err := r.db.Collection(entityName).CountDocuments(
		context.Background(),
		bson.M{"_id": id},
		options.Count().SetLimit(1),
	)
	return (count != 0), err
}

// FindByName : a generic function to search for store, merchant colelction only.
func (r Repository) FindByName(entityName entity.Collection, id string, v interface{}) error {
	return r.db.Collection(entityName).FindOne(
		context.Background(),
		bson.M{"_id": id},
	).Decode(v)
}

// FindByObjectID : find document by using hex object id string
func (r Repository) FindByObjectID(entityName entity.Collection, hex string, v interface{}) error {
	objectID, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		return errors.New("error decoding ID")
	}
	return r.db.Collection(entityName).FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(v)
}

// Delete : a generic function to delete one document using it's _id value
func (r Repository) Delete(entityName entity.Collection, hex string) error {
	objectID, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		return errors.New("error decoding ID")
	}

	_, err = r.db.Collection(entityName).DeleteOne(
		context.Background(),
		bson.M{"_id": objectID},
	)
	return err
}

// HealthCheck :
func (r Repository) HealthCheck(c context.Context) error {
	return nil
}
