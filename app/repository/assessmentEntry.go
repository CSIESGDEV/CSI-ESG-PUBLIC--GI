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

// FindAssessmentEntryFilter :
type FindAssessmentEntryFilter struct {
	Cursor        string
	AssessmentID  string
	QuestionSetID string
	SMEID         string
	QuestionType  string
	IDs           []string
	RespondStatus entity.RespondStatus
}

// SubmitAssessmentFilter :
type SubmitAssessmentFilter struct {
	AssessmentID string
	SMEIDs       []string
}

// FindAssessmentEntryByID :
func (r Repository) FindAssessmentEntryByID(ctx context.Context, id string) (*entity.AssessmentEntry, error) {
	assessmentEntry := new(entity.AssessmentEntry)
	err := r.FindByObjectID(entity.CollectionAssessmentEntry, id, &assessmentEntry)
	if err != nil {
		return nil, err
	}
	return assessmentEntry, nil
}

// FindAssessmentEntries :
func (r Repository) FindAssessmentEntries(ctx context.Context, filter FindAssessmentEntryFilter) ([]*entity.AssessmentEntry, string, error) {
	var assessmentEntries []*entity.AssessmentEntry

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
				return assessmentEntries, "", err
			}
			oIds = append(oIds, oid)
		}
		query["_id"] = bson.M{"$in": oIds}
	}

	if filter.AssessmentID != "" {
		oid, err := primitive.ObjectIDFromHex(filter.AssessmentID)
		if err != nil {
			return assessmentEntries, "", err
		}
		query["assessmentID"] = oid
	}

	if filter.SMEID != "" {
		oid, err := primitive.ObjectIDFromHex(filter.SMEID)
		if err != nil {
			return assessmentEntries, "", err
		}
		query["smeID"] = oid
	}

	if filter.QuestionSetID != "" {
		oid, err := primitive.ObjectIDFromHex(filter.QuestionSetID)
		if err != nil {
			return assessmentEntries, "", err
		}
		query["questionSetID"] = oid
	}

	if filter.QuestionType != "" {
		query["questionType"] = filter.QuestionType
	}

	if filter.RespondStatus != "" {
		query["respondStatus"] = filter.RespondStatus
	}

	nextCursor, err := r.db.Collection(entity.CollectionAssessmentEntry).Find(
		ctx,
		query,
		options.Find().SetLimit(limit+1).SetSort(sortQuery))

	if err != nil {
		return nil, "", err
	}
	defer nextCursor.Close(ctx)

	for nextCursor.Next(ctx) {
		assessmentEntry := new(entity.AssessmentEntry)
		if err := nextCursor.Decode(assessmentEntry); err != nil {
			return nil, "", err
		}

		assessmentEntries = append(assessmentEntries, assessmentEntry)
	}

	if err := nextCursor.Err(); err != nil {
		return nil, "", err
	}

	if len(assessmentEntries) > int(limit) {
		return assessmentEntries[:len(assessmentEntries)-1], assessmentEntries[len(assessmentEntries)-1].ID.Hex(), nil
	}
	return assessmentEntries, strconv.FormatInt(nextCursor.ID(), 10), nil
}

// BulkInsertAssessmentEntries :
func (r Repository) BulkInsertAssessmentEntries(ctx context.Context, entries []*entity.AssessmentEntry) (*mongo.InsertManyResult, error) {
	var docs []interface{}
	for _, t := range entries {
		docs = append(docs, t)
	}

	opts := options.InsertMany().SetOrdered(false)
	res, err := r.db.Collection(entity.CollectionAssessmentEntry).InsertMany(ctx, docs, opts)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// BulkWriteAssessmentEntries :
func (r Repository) BulkWriteAssessmentEntries(ctx context.Context, entries []*entity.AssessmentEntry) (interface{}, error) {
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
		if updateResult, err = r.db.Collection(entity.CollectionAssessmentEntry).BulkWrite(ctx, operations, &bulkOption); err != nil {
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

// BulkDeleteAssessmentEntries :
func (r Repository) BulkDeleteAssessmentEntries(ctx context.Context, AssessmentIDs []string) (*mongo.DeleteResult, error) {
	query := bson.M{}

	if len(AssessmentIDs) > 0 {
		ids := make([]primitive.ObjectID, 0)
		for i := 0; i < len(AssessmentIDs); i++ {
			id, err := primitive.ObjectIDFromHex(AssessmentIDs[i])
			if err != nil {
				return nil, err
			}
			ids = append(ids, id)
		}
		query["assessmentID"] = bson.M{"$in": ids}
	}

	return r.db.Collection(entity.CollectionAssessmentEntry).DeleteMany(ctx, query)
}

// SubmitAssessment :
func (r Repository) SubmitAssessment(ctx context.Context, filter SubmitAssessmentFilter) (*mongo.UpdateResult, error) {
	query := bson.M{
		"assessmentID":  filter.AssessmentID,
		"respondStatus": bson.M{"$eq": entity.ResponseStatusInProgress},
	}
	if len(filter.SMEIDs) > 0 {
		ids := make([]primitive.ObjectID, 0)
		for i := 0; i < len(filter.SMEIDs); i++ {
			id, err := primitive.ObjectIDFromHex(filter.SMEIDs[i])
			if err != nil {
				return nil, err
			}
			ids = append(ids, id)
		}
		query["smeID"] = bson.M{"$in": ids}
	}

	return r.db.Collection(entity.CollectionAssessmentEntry).UpdateMany(ctx, query, bson.M{
		"$set": bson.M{
			"respondStatus": entity.ResponseStatusToSubmitted,
			"submittedAt":   time.Now().UTC(),
		},
	}, options.Update())
}
