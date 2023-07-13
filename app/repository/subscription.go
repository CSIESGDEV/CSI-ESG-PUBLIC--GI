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

// FindSubscriptionFilter :
type FindSubscriptionFilter struct {
	Cursor                    string
	IDs                       []string
	CorporateIDs              []string
	PaymentStatus             string
	SubscriptionStartBefore   string // find all subscriptions that yet to start
	SubscriptionStartAfter    string // find all subscriptions that already start
	SubscriptionExpiredBefore string // find all subscriptions that not yet expired
	SubscriptionExpiredAfter  string // find all subscriptions that expired after this date
}

// CreateSubscription :
func (r Repository) CreateSubscription(ctx context.Context, i entity.Subscription) (*mongo.InsertOneResult, error) {
	return r.Create(entity.CollectionSubscription, i)
}

// FindSubscriptionByID :
func (r Repository) FindSubscriptionByID(ctx context.Context, id string) (*entity.Subscription, error) {
	subscription := new(entity.Subscription)
	err := r.FindByObjectID(entity.CollectionSubscription, id, &subscription)
	if err != nil {
		return nil, err
	}
	return subscription, nil
}

// FindSubscriptions :
func (r Repository) FindSubscriptions(ctx context.Context, filter FindSubscriptionFilter) ([]*entity.Subscription, string, error) {
	var subscriptions []*entity.Subscription

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
				return subscriptions, "", err
			}
			oIds = append(oIds, oid)
		}
		query["_id"] = bson.M{"$in": oIds}
	}

	if len(filter.CorporateIDs) > 0 {
		oIds := make([]primitive.ObjectID, 0)
		for i := 0; i < len(filter.CorporateIDs); i++ {
			oid, err := primitive.ObjectIDFromHex(filter.CorporateIDs[i])
			if err != nil {
				return subscriptions, "", err
			}
			oIds = append(oIds, oid)
		}
		query["corporateID"] = bson.M{"$in": oIds}
	}

	if filter.SubscriptionStartBefore != "" {
		startAfter, _ := time.Parse("2006-01-02T15:04:05Z", filter.SubscriptionStartBefore)
		query["subscriptionStartDate"] = bson.M{"$gt": primitive.NewDateTimeFromTime(startAfter)}
	}

	if filter.SubscriptionStartBefore != "" {
		startBefore, _ := time.Parse("2006-01-02T15:04:05Z", filter.SubscriptionStartBefore)
		query["subscriptionStartDate"] = bson.M{"$lte": primitive.NewDateTimeFromTime(startBefore)}
	}

	if filter.SubscriptionExpiredBefore != "" {
		expiredBefore, _ := time.Parse("2006-01-02T15:04:05Z", filter.SubscriptionExpiredBefore)
		query["subscriptionEndDate"] = bson.M{"$lte": primitive.NewDateTimeFromTime(expiredBefore)}
	}

	if filter.SubscriptionExpiredAfter != "" {
		expiredAfter, _ := time.Parse("2006-01-02T15:04:05Z", filter.SubscriptionExpiredAfter)
		query["subscriptionEndDate"] = bson.M{"$gt": primitive.NewDateTimeFromTime(expiredAfter)}
	}

	nextCursor, err := r.db.Collection(entity.CollectionSubscription).Find(
		ctx,
		query,
		options.Find().SetLimit(limit+1).SetSort(sortQuery))

	if err != nil {
		return nil, "", err
	}
	defer nextCursor.Close(ctx)

	for nextCursor.Next(ctx) {
		subscription := new(entity.Subscription)
		if err := nextCursor.Decode(subscription); err != nil {
			return nil, "", err
		}

		subscriptions = append(subscriptions, subscription)
	}

	if err := nextCursor.Err(); err != nil {
		return nil, "", err
	}

	if len(subscriptions) > int(limit) {
		return subscriptions[:len(subscriptions)-1], subscriptions[len(subscriptions)-1].ID.Hex(), nil
	}
	return subscriptions, strconv.FormatInt(nextCursor.ID(), 10), nil
}

// UpsertSubscription :
func (r Repository) UpsertSubscription(i *entity.Subscription) (*mongo.UpdateResult, error) {
	return r.db.Collection(entity.CollectionSubscription).UpdateOne(
		context.Background(),
		bson.M{"_id": i.ID},
		bson.M{"$set": i},
		options.Update().SetUpsert(true))
}

// DeleteSubscriptionByID : update status to deleted
func (r Repository) DeleteSubscriptionByID(i *entity.Subscription) (*mongo.UpdateResult, error) {
	i.DeletedAt = time.Now().UTC()
	return r.UpsertSubscription(i)
}
