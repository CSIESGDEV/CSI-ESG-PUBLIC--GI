package repository

import (
	"context"
	"csi-api/app/entity"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FindLearningResourceFilter :
type FindLearningResourceFilter struct {
	Cursor    string
	Indicator string
	Type      string
	Link      string
	IDs       []string
	Status    entity.Status
}

// CreateLearningResource :
func (r Repository) CreateLearningResource(ctx context.Context, i entity.LearningResource) (*mongo.InsertOneResult, error) {
	return r.Create(entity.CollectionLearningResource, i)
}

// FindLearningResourceByID :
func (r Repository) FindLearningResourceByID(ctx context.Context, id string) (*entity.LearningResource, error) {
	learningResource := new(entity.LearningResource)
	err := r.FindByObjectID(entity.CollectionLearningResource, id, &learningResource)
	if err != nil {
		return nil, err
	}
	return learningResource, nil
}

// FindLearningResources :
func (r Repository) FindLearningResources(ctx context.Context, filter FindLearningResourceFilter) ([]*entity.LearningResource, string, error) {
	var learningResources []*entity.LearningResource

	query := bson.M{}
	sortQuery := bson.M{}

	var limit int64 = 50

	if filter.Cursor != "" {
		objectID, err := primitive.ObjectIDFromHex(filter.Cursor)
		if err != nil {
			return nil, "", err
		}
		query["_id"] = bson.M{"$gte": objectID}
	}

	if len(filter.IDs) > 0 {
		oIds := make([]primitive.ObjectID, 0)
		for i := 0; i < len(filter.IDs); i++ {
			oid, err := primitive.ObjectIDFromHex(filter.IDs[i])
			if err != nil {
				return learningResources, "", err
			}
			oIds = append(oIds, oid)
		}
		query["_id"] = bson.M{"$in": oIds}
	}

	if filter.Indicator != "" {
		query["indicator"] = primitive.Regex{Pattern: filter.Indicator, Options: "i"}
	}

	if filter.Type != "" {
		query["type"] = primitive.Regex{Pattern: filter.Type, Options: "i"}
	}

	if filter.Link != "" {
		query["link"] = filter.Link
	}

	if filter.Status != "" {
		query["status"] = filter.Status
	}

	nextCursor, err := r.db.Collection(entity.CollectionLearningResource).Find(
		ctx,
		query,
		options.Find().SetLimit(limit+1).SetSort(sortQuery))

	if err != nil {
		return nil, "", err
	}
	defer nextCursor.Close(ctx)

	for nextCursor.Next(ctx) {
		learningResource := new(entity.LearningResource)
		if err := nextCursor.Decode(learningResource); err != nil {
			return nil, "", err
		}

		learningResources = append(learningResources, learningResource)
	}

	if err := nextCursor.Err(); err != nil {
		return nil, "", err
	}

	if len(learningResources) > int(limit) {
		return learningResources[:len(learningResources)-1], learningResources[len(learningResources)-1].ID.Hex(), nil
	}
	return learningResources, strconv.FormatInt(nextCursor.ID(), 10), nil
}

// UpsertLearningResource :
func (r Repository) UpsertLearningResource(i *entity.LearningResource) (*mongo.UpdateResult, error) {
	return r.db.Collection(entity.CollectionLearningResource).UpdateOne(
		context.Background(),
		bson.M{"_id": i.ID},
		bson.M{"$set": i},
		options.Update().SetUpsert(true))
}

// DeleteLearningResourceByID : update status to deleted
func (r Repository) DeleteLearningResourceByID(i *entity.LearningResource) (*mongo.UpdateResult, error) {
	i.Status = entity.StatusDeleted
	i.DeletedAt = time.Now().UTC()
	return r.UpsertLearningResource(i)
}
