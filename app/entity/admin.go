package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

// Admin :
type Admin struct {
	ID           *primitive.ObjectID `bson:"_id" json:"id"`
	FirstName    string              `bson:"firstName" json:"firstName"`
	LastName     string              `bson:"lastName" json:"lastName"`
	IC           string              `bson:"ic" json:"ic"`
	Email        string              `bson:"email" json:"email"`
	Contact      string              `bson:"contact" json:"contact"`
	PasswordHash string              `bson:"passwordHash" json:"-"`
	PasswordSalt string              `bson:"passwordSalt" json:"-"`
	Role         UserRole            `bson:"role" json:"role"`
	CompanyName  string              `bson:"companyName" json:"companyName"`
	Position     string              `bson:"position" json:"position"`
	ApprovedBy   *primitive.ObjectID `bson:"approvedBy" json:"approvedBy"`
	Status       UserStatus          `bson:"status" json:"status"`
	Model        `bson:",inline"`
}
