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

// FindQuestionSetFilter :
type FindQuestionSetFilter struct {
	Cursor string
	Label  string
	IDs    []string
	Status entity.Status
}

// CreateQuestionSet :
func (r Repository) CreateQuestionSet(ctx context.Context, i entity.QuestionSet) (*mongo.InsertOneResult, error) {
	return r.Create(entity.CollectionQuestionSet, i)
}

// FindQuestionSetByID :
func (r Repository) FindQuestionSetByID(ctx context.Context, id string) (*entity.QuestionSet, error) {
	questionSet := new(entity.QuestionSet)
	err := r.FindByObjectID(entity.CollectionQuestionSet, id, &questionSet)
	if err != nil {
		return nil, err
	}
	return questionSet, nil
}

// FindQuestionSets :
func (r Repository) FindQuestionSets(ctx context.Context, filter FindQuestionSetFilter) ([]*entity.QuestionSet, string, error) {
	var questionSets []*entity.QuestionSet

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
				return questionSets, "", err
			}
			oIds = append(oIds, oid)
		}
		query["_id"] = bson.M{"$in": oIds}
	}

	if filter.Label != "" {
		query["label"] = primitive.Regex{Pattern: filter.Label, Options: "i"}
	}

	if filter.Status != "" {
		query["status"] = filter.Status
	}

	nextCursor, err := r.db.Collection(entity.CollectionQuestionSet).Find(
		ctx,
		query,
		options.Find().SetLimit(limit+1).SetSort(sortQuery))

	if err != nil {
		return nil, "", err
	}
	defer nextCursor.Close(ctx)

	for nextCursor.Next(ctx) {
		QuestionSet := new(entity.QuestionSet)
		if err := nextCursor.Decode(QuestionSet); err != nil {
			return nil, "", err
		}

		questionSets = append(questionSets, QuestionSet)
	}

	if err := nextCursor.Err(); err != nil {
		return nil, "", err
	}

	if len(questionSets) > int(limit) {
		return questionSets[:len(questionSets)-1], questionSets[len(questionSets)-1].ID.Hex(), nil
	}
	return questionSets, strconv.FormatInt(nextCursor.ID(), 10), nil
}

// UpsertQuestionSet :
func (r Repository) UpsertQuestionSet(i *entity.QuestionSet) (*mongo.UpdateResult, error) {
	return r.db.Collection(entity.CollectionQuestionSet).UpdateOne(
		context.Background(),
		bson.M{"_id": i.ID},
		bson.M{"$set": i},
		options.Update().SetUpsert(true))
}

// DeleteQuestionSetByID : update status to deleted
func (r Repository) DeleteQuestionSetByID(i *entity.QuestionSet) (*mongo.UpdateResult, error) {
	i.Status = entity.StatusDeleted
	i.DeletedAt = time.Now().UTC()
	return r.UpsertQuestionSet(i)
}
