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

// FindAssessmentFilter :
type FindAssessmentFilter struct {
	Cursor      string
	SMEID       string
	SharedWiths []string
	IDs         []string
	Status      entity.Status
}

// CreateAssessment :
func (r Repository) CreateAssessment(ctx context.Context, i entity.Assessment) (*mongo.InsertOneResult, error) {
	return r.Create(entity.CollectionAssessment, i)
}

// FindAssessmentByID :
func (r Repository) FindAssessmentByID(ctx context.Context, id string) (*entity.Assessment, error) {
	assessment := new(entity.Assessment)
	err := r.FindByObjectID(entity.CollectionAssessment, id, &assessment)
	if err != nil {
		return nil, err
	}
	return assessment, nil
}

// FindAssessments :
func (r Repository) FindAssessments(ctx context.Context, filter FindAssessmentFilter) ([]*entity.Assessment, string, error) {
	var assessments []*entity.Assessment

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
				return assessments, "", err
			}
			oIds = append(oIds, oid)
		}
		query["_id"] = bson.M{"$in": oIds}
	}

	if filter.SMEID != "" {
		oid, err := primitive.ObjectIDFromHex(filter.SMEID)
		if err != nil {
			return assessments, "", err
		}
		query["smeID"] = oid
	}

	if len(filter.SharedWiths) > 0 {
		oIds := make([]primitive.ObjectID, 0)
		for i := 0; i < len(filter.SharedWiths); i++ {
			oid, err := primitive.ObjectIDFromHex(filter.SharedWiths[i])
			if err != nil {
				return assessments, "", err
			}
			oIds = append(oIds, oid)
		}
		query["sharedWith.corporateID"] = bson.M{"$in": oIds}
	}

	if filter.Status != "" {
		query["status"] = filter.Status
	}

	nextCursor, err := r.db.Collection(entity.CollectionAssessment).Find(
		ctx,
		query,
		options.Find().SetLimit(limit+1).SetSort(sortQuery))

	if err != nil {
		return nil, "", err
	}
	defer nextCursor.Close(ctx)

	for nextCursor.Next(ctx) {
		assessment := new(entity.Assessment)
		if err := nextCursor.Decode(assessment); err != nil {
			return nil, "", err
		}

		assessments = append(assessments, assessment)
	}

	if err := nextCursor.Err(); err != nil {
		return nil, "", err
	}

	if len(assessments) > int(limit) {
		return assessments[:len(assessments)-1], assessments[len(assessments)-1].ID.Hex(), nil
	}
	return assessments, strconv.FormatInt(nextCursor.ID(), 10), nil
}

// UpsertAssessment :
func (r Repository) UpsertAssessment(i *entity.Assessment) (*mongo.UpdateResult, error) {
	return r.db.Collection(entity.CollectionAssessment).UpdateOne(
		context.Background(),
		bson.M{"_id": i.ID},
		bson.M{"$set": i},
		options.Update().SetUpsert(true))
}

// DeleteAssessmentByID : update status to deleted
func (r Repository) DeleteAssessmentByID(i *entity.Assessment) (*mongo.UpdateResult, error) {
	i.Status = entity.StatusDeleted
	i.DeletedAt = time.Now().UTC()
	return r.UpsertAssessment(i)
}
