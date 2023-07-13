package repository

import (
	"context"
	"csi-api/app/entity"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FindQuestionFilter :
type FindQuestionFilter struct {
	Cursor        string
	QuestionSetID string
	Dimension     string
	SubCategory   string
	Indicator     string
	QuestionType  string
	IDs           []string
	Status        entity.Status
}

// CreateQuestion :
func (r Repository) CreateQuestion(ctx context.Context, i entity.Question) (*mongo.InsertOneResult, error) {
	return r.Create(entity.CollectionQuestion, i)
}

// FindQuestionByID :
func (r Repository) FindQuestionByID(ctx context.Context, id string) (*entity.Question, error) {
	question := new(entity.Question)
	err := r.FindByObjectID(entity.CollectionQuestion, id, &question)
	if err != nil {
		return nil, err
	}
	return question, nil
}

// FindQuestions :
func (r Repository) FindQuestions(ctx context.Context, filter FindQuestionFilter) ([]*entity.Question, string, error) {
	var questions []*entity.Question

	query := bson.M{}
	sortQuery := bson.M{}

	var limit int64 = 60

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
				return questions, "", err
			}
			oIds = append(oIds, oid)
		}
		query["_id"] = bson.M{"$in": oIds}
	}

	if filter.QuestionSetID != "" {
		oid, err := primitive.ObjectIDFromHex(filter.QuestionSetID)
		if err != nil {
			return questions, "", err
		}
		query["questionSetID"] = oid
	}

	if filter.Dimension != "" {
		query["dimension"] = filter.Dimension
	}

	if filter.SubCategory != "" {
		query["subCategory"] = filter.SubCategory
	}

	if filter.Indicator != "" {
		query["indicator"] = filter.Indicator
	}

	if filter.QuestionType != "" {
		query["questionType"] = filter.QuestionType
	}

	if filter.Status != "" {
		query["status"] = filter.Status
	}

	nextCursor, err := r.db.Collection(entity.CollectionQuestion).Find(
		ctx,
		query,
		options.Find().SetLimit(limit+1).SetSort(sortQuery))

	if err != nil {
		return nil, "", err
	}
	defer nextCursor.Close(ctx)

	for nextCursor.Next(ctx) {
		Question := new(entity.Question)
		if err := nextCursor.Decode(Question); err != nil {
			return nil, "", err
		}

		questions = append(questions, Question)
	}

	if err := nextCursor.Err(); err != nil {
		return nil, "", err
	}

	if len(questions) > int(limit) {
		return questions[:len(questions)-1], questions[len(questions)-1].ID.Hex(), nil
	}
	return questions, strconv.FormatInt(nextCursor.ID(), 10), nil
}

// UpsertQuestion :
func (r Repository) UpsertQuestion(i *entity.Question) (*mongo.UpdateResult, error) {
	return r.db.Collection(entity.CollectionQuestion).UpdateOne(
		context.Background(),
		bson.M{"_id": i.ID},
		bson.M{"$set": i},
		options.Update().SetUpsert(true))
}

// DeleteQuestionByID : update status to deleted
func (r Repository) DeleteQuestionByID(i *entity.Question) (*mongo.UpdateResult, error) {
	i.Status = entity.StatusDeleted
	i.DeletedAt = time.Now().UTC()
	return r.UpsertQuestion(i)
}

// BulkInsertQuestions :
func (r Repository) BulkInsertQuestions(ctx context.Context, entries []*entity.Question) (*mongo.InsertManyResult, error) {
	var docs []interface{}
	for _, t := range entries {
		docs = append(docs, t)
	}

	opts := options.InsertMany().SetOrdered(false)
	res, err := r.db.Collection(entity.CollectionQuestion).InsertMany(ctx, docs, opts)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// BulkWriteQuestions :
func (r Repository) BulkWriteQuestions(ctx context.Context, entries []*entity.Question) (interface{}, error) {
	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		var operations []mongo.WriteModel
		for _, r := range entries {
			operation := mongo.NewUpdateOneModel()
			operation.SetFilter(bson.D{{Key: "_id", Value: r.ID}})
			operation.SetUpdate(bson.D{{Key: "$set", Value: r}})
			operation.SetUpsert(false)
			operations = append(operations, operation)
		}
		bulkOption := options.BulkWriteOptions{}
		bulkOption.SetOrdered(true)
		// Important: You must pass sessCtx as the Context parameter to the operations for them to be executed in the
		// transaction.
		var updateResult *mongo.BulkWriteResult
		var err error
		if updateResult, err = r.db.Collection(entity.CollectionQuestion).BulkWrite(ctx, operations, &bulkOption); err != nil {
			return nil, err
		}
		return updateResult, nil
	}
	session, err := r.db.Client().StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)
	result, err := session.WithTransaction(ctx, callback)
	if err != nil {
		return nil, err
	}
	fmt.Printf("result: %v\n", result)
	return result, nil
}
