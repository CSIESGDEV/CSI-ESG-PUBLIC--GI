package repository

import (
	"context"
	"sme-api/app/entity"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FindAdminFilter :
type FindAdminFilter struct {
	Cursor    string
	FirstName string
	LastName  string
	IDs       []string
	Emails    []string
	Roles     []string
	Status    entity.UserStatus
}

// CreateAdmin :
func (r Repository) CreateAdmin(ctx context.Context, i entity.Admin) (*mongo.InsertOneResult, error) {
	return r.Create(entity.CollectionAdmin, i)
}

// FindAdminByID :
func (r Repository) FindAdminByID(ctx context.Context, id string) (*entity.Admin, error) {
	admin := new(entity.Admin)
	err := r.FindByObjectID(entity.CollectionAdmin, id, &admin)
	if err != nil {
		return nil, err
	}
	return admin, nil
}

// FindAdminByEmail :
func (r Repository) FindAdminByEmail(ctx context.Context, email string) (*entity.Admin, error) {
	admin := new(entity.Admin)
	if err := r.db.Collection(entity.CollectionAdmin).FindOne(ctx, bson.M{
		"email": email,
	}).Decode(&admin); err != nil {
		return nil, err
	}
	return admin, nil
}

// FindAdmins :
func (r Repository) FindAdmins(ctx context.Context, filter FindAdminFilter) ([]*entity.Admin, string, error) {
	var admins []*entity.Admin

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
				return admins, "", err
			}
			oIds = append(oIds, oid)
		}
		query["_id"] = bson.M{"$in": oIds}
	}

	if len(filter.Roles) > 0 {
		query["role"] = bson.M{"$in": filter.Roles}
	}

	if len(filter.Emails) > 0 {
		query["email"] = bson.M{"$in": filter.Emails}
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

	nextCursor, err := r.db.Collection(entity.CollectionAdmin).Find(
		ctx,
		query,
		options.Find().SetLimit(limit+1).SetSort(sortQuery))

	if err != nil {
		return nil, "", err
	}
	defer nextCursor.Close(ctx)

	for nextCursor.Next(ctx) {
		admin := new(entity.Admin)
		if err := nextCursor.Decode(admin); err != nil {
			return nil, "", err
		}

		admins = append(admins, admin)
	}

	if err := nextCursor.Err(); err != nil {
		return nil, "", err
	}

	if len(admins) > int(limit) {
		return admins[:len(admins)-1], admins[len(admins)-1].ID.Hex(), nil
	}
	return admins, strconv.FormatInt(nextCursor.ID(), 10), nil
}

// UpsertAdmin :
func (r Repository) UpsertAdmin(i *entity.Admin) (*mongo.UpdateResult, error) {
	return r.db.Collection(entity.CollectionAdmin).UpdateOne(
		context.Background(),
		bson.M{"_id": i.ID},
		bson.M{"$set": i},
		options.Update().SetUpsert(true))
}

// DeleteAdminByID : update status to deleted
func (r Repository) DeleteAdminByID(i *entity.Admin) (*mongo.UpdateResult, error) {
	i.Status = entity.UserStatusDeleted
	i.DeletedAt = time.Now().UTC()
	return r.UpsertAdmin(i)
}
