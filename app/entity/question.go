package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

// QuestionType :
type QuestionType string

var (
	QuestionTypePsychographic QuestionType = "PSYCHOGRAPHIC"
	QuestionTypeDemographic   QuestionType = "DEMOGRAPHIC"
	QuestionTypeUpload        QuestionType = "UPLOAD"
)

// Answer :
type Answer struct {
	Bool bool     `bson:"scalar" json:"scalar" form:"scalar"`
	Text []string `bson:"text" json:"text" form:"text"`
}

// Question :
type Question struct {
	ID            *primitive.ObjectID `bson:"_id" json:"id"`
	QuestionSetID *primitive.ObjectID `bson:"questionSetID" json:"questionSetID"`
	Dimension     string              `bson:"dimension" json:"dimension"`
	// SubCategory   string              `bson:"subCategory" json:"subCategory"`
	// Indicator     string              `bson:"indicator" json:"indicator"`
	QuestionLabel string       `bson:"questionLabel" json:"questionLabel"`
	QuestionType  QuestionType `bson:"questionType" json:"questionType"`
	Options       []string     `bson:"options" json:"options"`
	Status        Status       `bson:"status" json:"status"`
	Model         `bson:",inline"`
}

// QuestionSet :
type QuestionSet struct {
	ID     *primitive.ObjectID `bson:"_id" json:"id"`
	Label  string              `bson:"label" json:"label"`
	Status Status              `bson:"status" json:"status"`
	Model  `bson:",inline"`
}
