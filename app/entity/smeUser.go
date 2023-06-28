package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

// SMEUser :
type SMEUser struct {
	ID             *primitive.ObjectID `bson:"_id" json:"id"`
	CompanyID      *primitive.ObjectID `bson:"companyID" json:"companyID"`
	FirstName      string              `bson:"firstName" json:"firstName"`
	LastName       string              `bson:"lastName" json:"lastName"`
	Title          UserTitle           `bson:"title" json:"title"`
	IC             string              `bson:"ic" json:"ic"`
	Email          string              `bson:"email" json:"email"`
	Contact        string              `bson:"contact" json:"contact"`
	MobileContact  string              `bson:"mobileContact" json:"mobileContact"`
	Position       string              `bson:"position" json:"position"`
	ProfilePicture string              `bson:"profilePicture" json:"profilePicture"`
	PasswordHash   string              `bson:"passwordHash" json:"-"`
	PasswordSalt   string              `bson:"passwordSalt" json:"-"`
	Status         UserStatus          `bson:"status" json:"status"`
	Role           UserRole            `bson:"role" json:"role"`
	Model          `bson:",inline"`
}
