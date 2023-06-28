package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

// LearningResource :
type LearningResource struct {
	ID        *primitive.ObjectID `bson:"_id" json:"id"`
	Indicator string              `bson:"indicator" json:"indicator"`
	Name      string              `bson:"name" json:"name"`
	Title     string              `bson:"title" json:"title"`
	Link      string              `bson:"link" json:"link"`
	Type      string              `bson:"type" json:"type"`
	Source    string              `bson:"source" json:"source"`
	Status    Status              `bson:"status" json:"status"`
	Model     `bson:",inline"`
}
