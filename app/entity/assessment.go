package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Assessment :
type Assessment struct {
	ID             *primitive.ObjectID `bson:"_id" json:"id"`
	SMEID          *primitive.ObjectID `bson:"smeID" json:"smeID"`
	QuestionSetID  *primitive.ObjectID `bson:"questionSetID" json:"questionSetID"`
	SharedWiths    []LinkCorporate     `bson:"sharedWiths" json:"sharedWiths"`
	SerialNo       string              `bson:"serialNo" json:"serialNo"`
	CompletionDate time.Time           `bson:"completionDate" json:"completionDate"`
	Report         string              `bson:"report" json:"report"`
	ISODocs        []string            `bson:"isoDocs" json:"isoDocs"`
	Status         Status              `bson:"status" json:"status"`
	Model          `bson:",inline"`
}

// AssessmentEntry :
type AssessmentEntry struct {
	ID                *primitive.ObjectID `bson:"_id" json:"id"`
	AssessmentID      *primitive.ObjectID `bson:"assessmentID" json:"assessmentID"`
	SMEID             *primitive.ObjectID `bson:"smeID" json:"smeID"`
	QuestionSetID     *primitive.ObjectID `bson:"questionSetID" json:"questionSetID"`
	QuestionID        *primitive.ObjectID `bson:"questionID" json:"questionID"`
	Question          *Question           `bson:"question" json:"question"`
	QuestionType      QuestionType        `bson:"questionType" json:"questionType"`
	Answer            Answer              `bson:"answer" json:"answer"`
	SubmittedDateTime time.Time           `bson:"submittedAt" json:"submittedAt"`
	RespondStatus     RespondStatus       `bson:"respondStatus" json:"respondStatus"`
	Model             `bson:",inline"`
}
