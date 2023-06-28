package handler

import (
	"context"
	"fmt"
	"net/http"
	"sme-api/app/entity"
	"sme-api/app/repository"
	"sme-api/app/response"
	"sme-api/app/response/errcode"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetConnections :
func (h Handler) GetConnections(c echo.Context) error {
	connections, cursor, err := h.repository.FindConnections(c.Request().Context(), repository.FindConnectionFilter{
		Cursor:            c.QueryParam("cursor"),
		IDs:               c.Request().URL.Query()["id"],
		RequestCompanyID:  c.QueryParam("requestCompanyID"),
		ReceivedCompanyID: c.QueryParam("receivedCompanyID"),
		Status:            entity.Status(c.QueryParam("status")),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, response.Items{Items: connections, Cursor: cursor, Count: len(connections)})
}

// UpdateConnection :
func (h Handler) UpdateConnection(c echo.Context) error {
	connectionId := c.QueryParam("id")
	var i struct {
		RequestCompanyID  string        `bson:"requestCompanyID" json:"requestCompanyID" form:"requestCompanyID" validate:"max=50"`
		ReceivedCompanyID string        `bson:"receivedCompanyID" json:"receivedCompanyID" form:"receivedCompanyID" validate:"max=50"`
		LinkDate          time.Time     `bson:"linkDate" json:"linkDate" form:"linkDate"`
		Status            entity.Status `bson:"status" json:"status" form:"status" validate:"eq=INVITED|eq=ACTIVE|eq=INACTIVE"`
	}

	// bind req input
	if err := c.Bind(&i); err != nil {
		return c.JSON(http.StatusBadRequest, response.Exception{Code: errcode.InvalidRequest, Error: err})
	}

	// validate
	if err := c.Validate(&i); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.Exception{Code: errcode.ValidationError, Error: err})
	}

	// check if Connection exists
	connection, err := h.repository.FindConnectionByID(c.Request().Context(), connectionId)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, response.Exception{Code: errcode.RecordNotFound, Error: fmt.Errorf("Connection %s not found", connectionId)})
		}
		return c.JSON(http.StatusInternalServerError, response.Exception{Code: errcode.ServerError, Error: err})
	}

	if i.RequestCompanyID != "" {
		reqIds, err := primitive.ObjectIDFromHex(i.RequestCompanyID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, response.Exception{
				Code:  errcode.ServerError,
				Error: err,
			})
		}
		connection.RequestCompanyID = &reqIds
	}

	if i.ReceivedCompanyID != "" {
		recIds, err := primitive.ObjectIDFromHex(i.ReceivedCompanyID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, response.Exception{
				Code:  errcode.ServerError,
				Error: err,
			})
		}
		connection.ReceivedCompanyID = &recIds
	}

	if i.Status != "" {
		connection.Status = i.Status
	}

	connection.LinkDate = time.Now().UTC()

	if _, err := h.repository.UpsertConnection(connection); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, response.Item{Item: connection})
}

// CreateConnection :
func (h Handler) CreateConnection(c echo.Context) error {
	var i struct {
		RequestCompanyID  string        `bson:"requestCompanyID" json:"requestCompanyID" form:"requestCompanyID" validate:"required,max=50"`
		ReceivedCompanyID string        `bson:"receivedCompanyID" json:"receivedCompanyID" form:"receivedCompanyID" validate:"required,max=50"`
		LinkDate          time.Time     `bson:"linkDate" json:"linkDate" form:"linkDate"`
		Status            entity.Status `bson:"status" json:"status" form:"status" validate:"eq=INVITED|eq=ACTIVE|eq=INACTIVE"`
	}

	// bind req input
	if err := c.Bind(&i); err != nil {
		return c.JSON(http.StatusBadRequest, response.Exception{Code: errcode.InvalidRequest, Error: err})
	}

	// validate
	if err := c.Validate(&i); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.Exception{Code: errcode.ValidationError, Error: err})
	}

	// create time object
	timeNow := time.Now().UTC()

	// create connection object
	connectionID := primitive.NewObjectID()

	reqIds, err := primitive.ObjectIDFromHex(i.RequestCompanyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	recIds, err := primitive.ObjectIDFromHex(i.ReceivedCompanyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	connection := entity.Connection{
		ID:                &connectionID,
		RequestCompanyID:  &reqIds,
		ReceivedCompanyID: &recIds,
		LinkDate:          timeNow,
		Status:            i.Status,
	}

	if _, err = h.repository.CreateConnection(c.Request().Context(), connection); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, response.Item{Item: connection})
}

// ValidateConnection :
func ValidateConnection(h Handler, ctx context.Context, companyID string) (*entity.Connection, int, *response.Exception) {
	// check if connection exists
	connection, err := h.repository.FindConnectionByID(ctx, companyID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &entity.Connection{}, http.StatusNotFound, &response.Exception{Code: errcode.RecordNotFound, Error: fmt.Errorf("connection %s not found", connection)}
		}
		return &entity.Connection{}, http.StatusInternalServerError, &response.Exception{Code: errcode.ServerError, Error: err}
	}

	if connection.Status != entity.StatusActive {
		return &entity.Connection{}, http.StatusBadRequest, &response.Exception{Code: errcode.StatusInActive, Error: fmt.Errorf("connection %s status inactive", connection)}
	}
	return connection, http.StatusOK, nil
}
