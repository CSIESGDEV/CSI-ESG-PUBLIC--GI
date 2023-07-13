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

// FindNewsFilter :
type FindNewsFilter struct {
	Cursor string
	Link   string
	IDs    []string
	Status entity.Status
}

// CreateNews :
func (r Repository) CreateNews(ctx context.Context, i entity.News) (*mongo.InsertOneResult, error) {
	return r.Create(entity.CollectionNews, i)
}

// FindNewsByID :
func (r Repository) FindNewsByID(ctx context.Context, id string) (*entity.News, error) {
	news := new(entity.News)
	err := r.FindByObjectID(entity.CollectionNews, id, &news)
	if err != nil {
		return nil, err
	}
	return news, nil
}

// FindNews :
func (r Repository) FindNews(ctx context.Context, filter FindNewsFilter) ([]*entity.News, string, error) {
	var news []*entity.News

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
				return news, "", err
			}
			oIds = append(oIds, oid)
		}
		query["_id"] = bson.M{"$in": oIds}
	}

	if filter.Link != "" {
		query["link"] = filter.Link
	}

	if filter.Status != "" {
		query["status"] = filter.Status
	}

	nextCursor, err := r.db.Collection(entity.CollectionNews).Find(
		ctx,
		query,
		options.Find().SetLimit(limit+1).SetSort(sortQuery))

	if err != nil {
		return nil, "", err
	}
	defer nextCursor.Close(ctx)

	for nextCursor.Next(ctx) {
		new := new(entity.News)
		if err := nextCursor.Decode(new); err != nil {
			return nil, "", err
		}

		news = append(news, new)
	}

	if err := nextCursor.Err(); err != nil {
		return nil, "", err
	}

	if len(news) > int(limit) {
		return news[:len(news)-1], news[len(news)-1].ID.Hex(), nil
	}
	return news, strconv.FormatInt(nextCursor.ID(), 10), nil
}

// UpsertNews :
func (r Repository) UpsertNews(i *entity.News) (*mongo.UpdateResult, error) {
	return r.db.Collection(entity.CollectionNews).UpdateOne(
		context.Background(),
		bson.M{"_id": i.ID},
		bson.M{"$set": i},
		options.Update().SetUpsert(true))
}

// DeleteNewsByID : update status to deleted
func (r Repository) DeleteNewsByID(i *entity.News) (*mongo.UpdateResult, error) {
	i.Status = entity.StatusDeleted
	i.DeletedAt = time.Now().UTC()
	return r.UpsertNews(i)
}
