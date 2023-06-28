package repository

import (
	"context"
	"fmt"
	"sme-api/app/entity"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FindSMEUserFilter :
type FindSMEUserFilter struct {
	Cursor    string
	FirstName string
	LastName  string
	CompanyID string
	IDs       []string
	Emails    []string
	Status    entity.UserStatus
}

// CreateSMEUser :
func (r Repository) CreateSMEUser(ctx context.Context, i entity.SMEUser) (*mongo.InsertOneResult, error) {
	return r.Create(entity.CollectionSMEUser, i)
}

// FindSMEUserByID :
func (r Repository) FindSMEUserByID(ctx context.Context, id string) (*entity.SMEUser, error) {
	user := new(entity.SMEUser)
	err := r.FindByObjectID(entity.CollectionSMEUser, id, &user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// FindSMEUserByEmail :
func (r Repository) FindSMEUserByEmail(ctx context.Context, email string) (*entity.SMEUser, error) {
	fmt.Println(email)
	user := new(entity.SMEUser)
	if err := r.db.Collection(entity.CollectionSMEUser).FindOne(ctx, bson.M{
		"email": email,
	}).Decode(&user); err != nil {
		return nil, err
	}
	return user, nil
}

// FindSMEUsers :
func (r Repository) FindSMEUsers(ctx context.Context, filter FindSMEUserFilter) ([]*entity.SMEUser, string, error) {
	var users []*entity.SMEUser

	query := bson.M{}
	sortQuery := bson.M{}

	var limit int64 = 100

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
				return users, "", err
			}
			oIds = append(oIds, oid)
		}
		query["_id"] = bson.M{"$in": oIds}
	}

	if len(filter.Emails) > 0 {
		query["email"] = bson.M{"$in": filter.Emails}
	}

	if filter.CompanyID != "" {
		oid, err := primitive.ObjectIDFromHex(filter.CompanyID)
		if err != nil {
			return users, "", err
		}
		query["companyID"] = oid
	}

	if filter.FirstName != "" {
		query["firstName"] = primitive.Regex{Pattern: filter.FirstName, Options: "i"}
	}

	if filter.LastName != "" {
		query["lastName"] = primitive.Regex{Pattern: filter.LastName, Options: "i"}
	}

	if filter.Status != "" {
		query["status"] = filter.Status
	}

	nextCursor, err := r.db.Collection(entity.CollectionSMEUser).Find(
		ctx,
		query,
		options.Find().SetLimit(limit+1).SetSort(sortQuery))

	if err != nil {
		return nil, "", err
	}
	defer nextCursor.Close(ctx)

	for nextCursor.Next(ctx) {
		user := new(entity.SMEUser)
		if err := nextCursor.Decode(user); err != nil {
			return nil, "", err
		}

		users = append(users, user)
	}

	if err := nextCursor.Err(); err != nil {
		return nil, "", err
	}

	if len(users) > int(limit) {
		return users[:len(users)-1], users[len(users)-1].ID.Hex(), nil
	}
	return users, strconv.FormatInt(nextCursor.ID(), 10), nil
}

// UpsertSMEUser :
func (r Repository) UpsertSMEUser(i *entity.SMEUser) (*mongo.UpdateResult, error) {
	return r.db.Collection(entity.CollectionSMEUser).UpdateOne(
		context.Background(),
		bson.M{"_id": i.ID},
		bson.M{"$set": i},
		options.Update().SetUpsert(true))
}

// DeleteSMEUserByID : update status to deleted
func (r Repository) DeleteSMEUserByID(i *entity.SMEUser) (*mongo.UpdateResult, error) {
	i.Status = entity.UserStatusDeleted
	i.DeletedAt = time.Now().UTC()
	return r.UpsertSMEUser(i)
}
