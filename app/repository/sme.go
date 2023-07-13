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

// FindSMEFilter :
type FindSMEFilter struct {
	Cursor      string
	CompanyName string
	SSMNumbers  []string
	IDs         []string
	LinkedWiths []string
	Status      entity.Status
}

// CreateSME :
func (r Repository) CreateSME(ctx context.Context, i entity.SME) (*mongo.InsertOneResult, error) {
	return r.Create(entity.CollectionSME, i)
}

// FindSMEByID :
func (r Repository) FindSMEByID(ctx context.Context, id string) (*entity.SME, error) {
	sme := new(entity.SME)
	err := r.FindByObjectID(entity.CollectionSME, id, &sme)
	if err != nil {
		return nil, err
	}
	return sme, nil
}

// FindSMEs :
func (r Repository) FindSMEs(ctx context.Context, filter FindSMEFilter) ([]*entity.SME, string, error) {
	var SMEs []*entity.SME

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
				return SMEs, "", err
			}
			oIds = append(oIds, oid)
		}
		query["_id"] = bson.M{"$in": oIds}
	}

	if len(filter.LinkedWiths) > 0 {
		oIds := make([]primitive.ObjectID, 0)
		for i := 0; i < len(filter.LinkedWiths); i++ {
			oid, err := primitive.ObjectIDFromHex(filter.LinkedWiths[i])
			if err != nil {
				return SMEs, "", err
			}
			oIds = append(oIds, oid)
		}
		query["linkedWiths.corporateID"] = bson.M{"$in": oIds}
	}

	if filter.CompanyName != "" {
		query["companyName"] = primitive.Regex{Pattern: filter.CompanyName, Options: "i"}
	}

	if len(filter.SSMNumbers) > 0 {
		query["ssmNumber"] = bson.M{"$in": filter.SSMNumbers}
	}

	if filter.Status != "" {
		query["status"] = filter.Status
	}

	nextCursor, err := r.db.Collection(entity.CollectionSME).Find(
		ctx,
		query,
		options.Find().SetLimit(limit+1).SetSort(sortQuery))

	if err != nil {
		return nil, "", err
	}
	defer nextCursor.Close(ctx)

	for nextCursor.Next(ctx) {
		SME := new(entity.SME)
		if err := nextCursor.Decode(SME); err != nil {
			return nil, "", err
		}

		SMEs = append(SMEs, SME)
	}

	if err := nextCursor.Err(); err != nil {
		return nil, "", err
	}

	if len(SMEs) > int(limit) {
		return SMEs[:len(SMEs)-1], SMEs[len(SMEs)-1].ID.Hex(), nil
	}
	return SMEs, strconv.FormatInt(nextCursor.ID(), 10), nil
}

// UpsertSME :
func (r Repository) UpsertSME(i *entity.SME) (*mongo.UpdateResult, error) {
	return r.db.Collection(entity.CollectionSME).UpdateOne(
		context.Background(),
		bson.M{"_id": i.ID},
		bson.M{"$set": i},
		options.Update().SetUpsert(true))
}

// DeleteSMEByID : update status to deleted
func (r Repository) DeleteSMEByID(i *entity.SME) (*mongo.UpdateResult, error) {
	i.Status = entity.StatusDeleted
	i.DeletedAt = time.Now().UTC()
	return r.UpsertSME(i)
}
