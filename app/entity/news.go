package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

// News :
type News struct {
	ID     *primitive.ObjectID `bson:"_id" json:"id"`
	Title  string              `bson:"title" json:"title"`
	Image  string              `bson:"image" json:"image"`
	Link   string              `bson:"link" json:"link"`
	Status Status              `bson:"status" json:"status"`
	Model  `bson:",inline"`
}
