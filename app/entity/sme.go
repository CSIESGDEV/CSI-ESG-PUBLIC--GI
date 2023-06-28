package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SME :
type SME struct {
	ID                   *primitive.ObjectID   `bson:"_id" json:"id"`
	CompanyName          string                `bson:"companyName" json:"companyName"`
	SSMNumber            string                `bson:"ssmNumber" json:"ssmNumber"`
	EducationType        string                `bson:"educationType" json:"educationType"`
	State                string                `bson:"state" json:"state"`
	PostCode             string                `bson:"postCode" json:"postCode"`
	MSIC                 string                `bson:"msic" json:"msic"`
	SSMDoc               string                `bson:"ssmDoc" json:"ssmDoc"`
	BusinessEntity       BusinessEntity        `bson:"businessEntity" json:"businessEntity"`
	RegisteredInEastMY   bool                  `bson:"registeredInEastMY" json:"registeredInEastMY"`
	ProfilePicture       string                `bson:"profilePicture" json:"profilePicture"`
	LinkedWiths          []LinkCorporate       `bson:"linkedWiths" json:"linkedWiths"`
	ParticipatedLearning []*primitive.ObjectID `bson:"participatedLearning" json:"participatedLearning"`
	Status               Status                `bson:"status" json:"status"`
	ApprovedBy           *primitive.ObjectID   `bson:"approvedBy" json:"approvedBy"`
	Model                `bson:",inline"`
}

// LinkCorporate :
type LinkCorporate struct {
	CorporateID *primitive.ObjectID `bson:"corporateID" json:"corporateID" form:"corporateID" validate:"max=50"`
	LinkDate    time.Time           `bson:"linkDate" json:"linkDate" form:"linkDate"`
	Status      Status              `bson:"status" json:"status" form:"status" validate:"eq=INVITED|eq=ACTIVE"`
}
