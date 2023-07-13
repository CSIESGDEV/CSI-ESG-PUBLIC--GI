package repository

import (
	"context"
	"csi-api/app/entity"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FindConnectionFilter :
type FindConnectionFilter struct {
	Cursor            string
	IDs               []string
	RequestCompanyID  string
	ReceivedCompanyID string
	Status            entity.Status
}

// CreateConnection :
func (r Repository) CreateConnection(ctx context.Context, i entity.Connection) (*mongo.InsertOneResult, error) {
	return r.Create(entity.CollectionConnection, i)
}

// FindConnectionByID :
func (r Repository) FindConnectionByID(ctx context.Context, id string) (*entity.Connection, error) {
	connection := new(entity.Connection)
	err := r.FindByObjectID(entity.CollectionConnection, id, &connection)
	if err != nil {
		return nil, err
	}
	return connection, nil
}

// FindConnections :
func (r Repository) FindConnections(ctx context.Context, filter FindConnectionFilter) ([]*entity.Connection, string, error) {
	var Connections []*entity.Connection

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
				return Connections, "", err
			}
			oIds = append(oIds, oid)
		}
		query["_id"] = bson.M{"$in": oIds}
	}

	if filter.RequestCompanyID != "" {
		oid, err := primitive.ObjectIDFromHex(filter.RequestCompanyID)
		if err != nil {
			return Connections, "", err
		}
		query["requestCompanyID"] = oid
	}

	if filter.ReceivedCompanyID != "" {
		oid, err := primitive.ObjectIDFromHex(filter.ReceivedCompanyID)
		if err != nil {
			return Connections, "", err
		}
		query["receivedCompanyID"] = oid
	}

	if filter.Status != "" {
		query["status"] = filter.Status
	}

	nextCursor, err := r.db.Collection(entity.CollectionConnection).Find(
		ctx,
		query,
		options.Find().SetLimit(limit+1).SetSort(sortQuery))

	if err != nil {
		return nil, "", err
	}
	defer nextCursor.Close(ctx)

	for nextCursor.Next(ctx) {
		Connection := new(entity.Connection)
		if err := nextCursor.Decode(Connection); err != nil {
			return nil, "", err
		}

		Connections = append(Connections, Connection)
	}

	if err := nextCursor.Err(); err != nil {
		return nil, "", err
	}

	if len(Connections) > int(limit) {
		return Connections[:len(Connections)-1], Connections[len(Connections)-1].ID.Hex(), nil
	}
	return Connections, strconv.FormatInt(nextCursor.ID(), 10), nil
}

// UpsertConnection :
func (r Repository) UpsertConnection(i *entity.Connection) (*mongo.UpdateResult, error) {
	return r.db.Collection(entity.CollectionConnection).UpdateOne(
		context.Background(),
		bson.M{"_id": i.ID},
		bson.M{"$set": i},
		options.Update().SetUpsert(true))
}
