package handler

import (
	"context"
	"csi-api/app/entity"
	"csi-api/app/kit/general"
	"csi-api/app/repository"
	"csi-api/app/response"
	"csi-api/app/response/errcode"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetSubscriptions :
func (h Handler) GetSubscriptions(c echo.Context) error {
	subscriptions, cursor, err := h.repository.FindSubscriptions(c.Request().Context(), repository.FindSubscriptionFilter{
		Cursor:                    c.QueryParam("cursor"),
		IDs:                       c.Request().URL.Query()["id"],
		CorporateIDs:              c.Request().URL.Query()["corporateId"],
		PaymentStatus:             c.QueryParam("paymentStatus"),
		SubscriptionStartBefore:   c.QueryParam("subscriptionStartBefore"),
		SubscriptionStartAfter:    c.QueryParam("subscriptionStartAfter"),
		SubscriptionExpiredBefore: c.QueryParam("subscriptionExpiredBefore"),
		SubscriptionExpiredAfter:  c.QueryParam("subscriptionExpiredAfter"),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, response.Items{Items: subscriptions, Cursor: cursor, Count: len(subscriptions)})
}

// DeleteSubscription :
func (h Handler) DeleteSubscription(c echo.Context) error {
	subscriptionId := c.QueryParam("id")

	subscription, httpStatus, exception := ValidateSubscription(h, c.Request().Context(), subscriptionId)
	if exception != nil {
		return c.JSON(httpStatus, exception)
	}

	if _, err := h.repository.DeleteSubscriptionByID(subscription); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{Code: errcode.ServerError, Error: err})
	}

	return c.JSON(http.StatusOK, nil)
}

// UpdateSubscription :
func (h Handler) UpdateSubscription(c echo.Context) error {
	subscriptionId := c.QueryParam("id")
	var i struct {
		CorporateID           string    `json:"corporateID" form:"corporateID" validate:"required,max=50"`
		SubscriptionPlan      string    `json:"subscriptionPlan" form:"subscriptionPlan"`
		SubscriptionPeriod    int       `json:"subscriptionPeriod" form:"subscriptionPeriod"`
		ActivationDate        time.Time `json:"activationDate" form:"activationDate"`
		PaymentStatus         string    `json:"paymentStatus" form:"paymentStatus"`
		PaymentReceivedDate   time.Time `json:"paymentReceivedDate" form:"paymentReceived"`
		ContractDate          time.Time `json:"contractDate" form:"contractDate"`
		ContractID            string    `json:"contractID" form:"contractID" validate:"max=14"`
		InvoiceNumber         string    `json:"invoiceNumber" form:"invoiceNumber" validate:"max=14"`
		SubscriptionStartDate time.Time `json:"subscriptionStartDate" form:"subscriptionStartDate"`
		SubscriptionEndDate   time.Time `json:"subscriptionEndDate" form:"subscriptionEndDate"`
		VerificationPIC       string    `json:"verificationPIC" form:"verificationPIC" validate:"max=50"`
	}

	// bind req input
	if err := c.Bind(&i); err != nil {
		return c.JSON(http.StatusBadRequest, response.Exception{Code: errcode.InvalidRequest, Error: err})
	}

	// validate
	if err := c.Validate(&i); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.Exception{Code: errcode.ValidationError, Error: err})
	}

	// check if subscription exists
	subscription, httpStatus, exception := ValidateSubscription(h, c.Request().Context(), subscriptionId)
	if exception != nil {
		return c.JSON(httpStatus, exception)
	}

	corporate, httpStatus, exception := ValidateSME(h, c.Request().Context(), i.CorporateID)
	if exception != nil {
		return c.JSON(httpStatus, exception)
	}
	subscription.CorporateID = corporate.ID

	if i.SubscriptionPlan != "" {
		subscription.SubscriptionPlan = i.SubscriptionPlan
	}

	if i.SubscriptionPeriod > 0 {
		subscription.SubscriptionPeriod = i.SubscriptionPeriod
	}

	if !i.ActivationDate.IsZero() {
		subscription.ActivationDate = i.ActivationDate
	}

	if i.PaymentStatus != "" {
		subscription.PaymentStatus = entity.Status(i.PaymentStatus)
	}

	if !i.PaymentReceivedDate.IsZero() {
		subscription.PaymentReceivedDate = i.PaymentReceivedDate
	}

	if !i.ContractDate.IsZero() {
		subscription.ContractDate = i.ContractDate
	}

	if i.ContractID != "" {
		subscription.ContractID = i.ContractID
	}

	if i.InvoiceNumber != "" {
		subscription.InvoiceNumber = i.InvoiceNumber
	}

	if !i.SubscriptionStartDate.IsZero() {
		subscription.SubscriptionStartDate = i.SubscriptionStartDate
	}

	if !i.SubscriptionEndDate.IsZero() {
		subscription.SubscriptionEndDate = i.SubscriptionEndDate
	}

	if i.VerificationPIC != "" {
		admin, httpStatus, exception := ValidateAdmin(h, c.Request().Context(), i.VerificationPIC)
		if exception != nil {
			return c.JSON(httpStatus, exception)
		}
		subscription.VerificationPIC = admin.ID
	}

	// check if receipt document is uploaded
	if receipt, _ := c.FormFile("receipt"); receipt != nil {
		// create path
		path := fmt.Sprintf("corporate/%s/subscription/%s/receipt/", corporate.ID.Hex(), subscriptionId)

		// push to s3 bucket
		// url, httpStatus, exception := aws.PushDocBucket(path, receipt)
		// if exception != nil {
		// 	return c.JSON(httpStatus, exception)
		// }
		exception := general.ReadFile(receipt, path)
		if exception != nil {
			return c.JSON(http.StatusInternalServerError, exception)
		}
		subscription.Receipt = path + receipt.Filename
	}

	// create time object
	subscription.Model.UpdatedAt = time.Now().UTC()

	if _, err := h.repository.UpsertSubscription(subscription); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, nil)
}

// CreateSubscription :
func (h Handler) CreateSubscription(c echo.Context) error {
	var i struct {
		CorporateID           string    `json:"corporateID" form:"corporateID" validate:"required,max=50"`
		SubscriptionPlan      string    `json:"subscriptionPlan" form:"subscriptionPlan" validate:"required"`
		SubscriptionPeriod    int       `json:"subscriptionPeriod" form:"subscriptionPeriod" validate:"required"`
		ActivationDate        time.Time `json:"activationDate" form:"activationDate" validate:""`
		PaymentStatus         string    `json:"paymentStatus" form:"paymentStatus" validate:""`
		PaymentReceivedDate   time.Time `json:"paymentReceivedDate" form:"paymentReceivedDate" validate:""`
		ContractDate          time.Time `json:"contractDate" form:"contractDate" validate:""`
		ContractID            string    `json:"contractID" form:"contractID" validate:"max=14"`
		InvoiceNumber         string    `json:"invoiceNumber" form:"invoiceNumber" validate:"max=14"`
		SubscriptionStartDate time.Time `json:"subscriptionStartDate" form:"subscriptionStartDate" validate:""`
		SubscriptionEndDate   time.Time `json:"subscriptionEndDate" form:"subscriptionEndDate" validate:""`
		VerificationPIC       string    `json:"verificationPIC" form:"verificationPIC" validate:"max=50"`
	}

	// bind req input
	if err := c.Bind(&i); err != nil {
		return c.JSON(http.StatusBadRequest, response.Exception{Code: errcode.InvalidRequest, Error: err})
	}

	// validate
	if err := c.Validate(&i); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.Exception{Code: errcode.ValidationError, Error: err})
	}

	// check if corporate exists
	corporate, httpStatus, exception := ValidateSME(h, c.Request().Context(), i.CorporateID)
	if exception != nil {
		return c.JSON(httpStatus, exception)
	}

	// check if admin exists
	// admin, httpStatus, exception := ValidateAdmin(h, c.Request().Context(), i.VerificationPIC)
	// if exception != nil {
	// 	return c.JSON(httpStatus, exception)
	// }

	// create time object
	timeNow := time.Now().UTC()

	// create Subscription object
	subscriptionID := primitive.NewObjectID()

	subscription := entity.Subscription{
		ID:                    &subscriptionID,
		CorporateID:           corporate.ID,
		SubscriptionPlan:      i.SubscriptionPlan,
		SubscriptionPeriod:    i.SubscriptionPeriod,
		ActivationDate:        i.ActivationDate,
		PaymentStatus:         entity.Status(i.PaymentStatus),
		PaymentReceivedDate:   i.PaymentReceivedDate,
		ContractDate:          i.ContractDate,
		ContractID:            i.ContractID,
		InvoiceNumber:         i.InvoiceNumber,
		SubscriptionStartDate: i.SubscriptionStartDate,
		SubscriptionEndDate:   i.SubscriptionEndDate,
		// VerificationPIC:       admin.ID,
		Model: entity.Model{
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		},
	}

	if _, err := h.repository.CreateSubscription(c.Request().Context(), subscription); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, response.Item{Item: subscription})
}

// ValidateSubscription :
func ValidateSubscription(h Handler, ctx context.Context, ID string) (*entity.Subscription, int, *response.Exception) {
	// check if subscription exists
	subscription, err := h.repository.FindSubscriptionByID(ctx, ID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &entity.Subscription{}, http.StatusNotFound, &response.Exception{Code: errcode.RecordNotFound, Error: fmt.Errorf("subscription %s not found", ID)}
		}
		return &entity.Subscription{}, http.StatusInternalServerError, &response.Exception{Code: errcode.ServerError, Error: err}
	}

	if !subscription.Model.DeletedAt.IsZero() {
		return &entity.Subscription{}, http.StatusBadRequest, &response.Exception{Code: errcode.StatusInActive, Error: fmt.Errorf("subscription %s status inactive", ID)}
	}
	return subscription, http.StatusOK, nil
}
