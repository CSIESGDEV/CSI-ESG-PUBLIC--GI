package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Connection :
type Connection struct {
	ID                *primitive.ObjectID `bson:"_id" json:"id"`
	RequestCompanyID  *primitive.ObjectID `bson:"requestCompanyID" json:"requestCompanyID"`
	ReceivedCompanyID *primitive.ObjectID `bson:"receivedCompanyID" json:"receivedCompanyID"`
	LinkDate          time.Time           `bson:"linkDate" json:"linkDate"`
	Status            Status              `bson:"status" json:"status" validate:"eq=INVITED|eq=ACTIVE"`
}
