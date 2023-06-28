package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Subscription struct {
	ID                    *primitive.ObjectID `bson:"_id" json:"id"`
	CorporateID           *primitive.ObjectID `bson:"corporateID" json:"corporateID"`
	SubscriptionPlan      string              `bson:"subscriptionPlan" json:"subscriptionPlan"`
	SubscriptionPeriod    int                 `bson:"subscriptionPeriod" json:"subscriptionPeriod"`
	ActivationDate        time.Time           `bson:"activationDate" json:"activationDate"`
	PaymentStatus         Status              `bson:"paymentStatus" json:"paymentStatus"`
	PaymentReceivedDate   time.Time           `bson:"paymentReceivedDate" json:"paymentReceivedDate"`
	ContractDate          time.Time           `bson:"contractDate" json:"contractDate"`
	ContractID            string              `bson:"contractID" json:"contractID"`
	InvoiceNumber         string              `bson:"invoiceNumber" json:"invoiceNumber"`
	SubscriptionStartDate time.Time           `bson:"subscriptionStartDate" json:"subscriptionStartDate"`
	SubscriptionEndDate   time.Time           `bson:"subscriptionEndDate" json:"subscriptionEndDate"`
	VerificationPIC       *primitive.ObjectID `bson:"verificationPIC" json:"verificationPIC"`
	Receipt               string              `bson:"receipt" json:"receipt"`
	Model                 `bson:",inline"`
}
