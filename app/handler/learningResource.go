package handler

import (
	"context"
	"csi-api/app/entity"
	"csi-api/app/repository"
	"csi-api/app/response"
	"csi-api/app/response/errcode"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetLearningResources :
func (h Handler) GetLearningResources(c echo.Context) error {
	learningResources, cursor, err := h.repository.FindLearningResources(c.Request().Context(), repository.FindLearningResourceFilter{
		Cursor:    c.QueryParam("cursor"),
		IDs:       c.Request().URL.Query()["id"],
		Indicator: c.QueryParam("indicator"),
		Type:      c.QueryParam("type"),
		Status:    entity.Status(c.QueryParam("status")),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, response.Items{Items: learningResources, Cursor: cursor, Count: len(learningResources)})
}

// UpdateLearningResource :
func (h Handler) UpdateLearningResource(c echo.Context) error {
	learningResourceId := c.QueryParam("id")
	var i struct {
		Indicator string        `json:"indicator" form:"indicator"`
		Name      string        `json:"name" form:"name"`
		Title     string        `json:"title" form:"title"`
		Link      string        `json:"link" form:"link"`
		Type      string        `json:"type" form:"type"`
		Source    string        `json:"source" form:"source" `
		Status    entity.Status `json:"status" form:"status"`
	}

	// bind req input
	if err := c.Bind(&i); err != nil {
		return c.JSON(http.StatusBadRequest, response.Exception{Code: errcode.InvalidRequest, Error: err})
	}

	// cleanup
	i.Link = strings.TrimSpace(i.Link)

	// validate
	if err := c.Validate(&i); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.Exception{Code: errcode.ValidationError, Error: err})
	}

	// check if LearningResource exists
	learningResource, err := h.repository.FindLearningResourceByID(c.Request().Context(), learningResourceId)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, response.Exception{Code: errcode.RecordNotFound, Error: fmt.Errorf("learningResource %s not found", learningResourceId)})
		}
		return c.JSON(http.StatusInternalServerError, response.Exception{Code: errcode.ServerError, Error: err})
	}

	if i.Name != "" {
		learningResource.Name = i.Name
	}

	if i.Indicator != "" {
		learningResource.Indicator = i.Indicator
	}

	if i.Link != "" {
		learningResource.Link = i.Link
	}

	if i.Title != "" {
		learningResource.Title = i.Title
	}

	if i.Type != "" {
		learningResource.Type = i.Type
	}

	if i.Source != "" {
		learningResource.Source = i.Source
	}

	if i.Status != "" {
		learningResource.Status = i.Status
	}

	learningResource.Model.UpdatedAt = time.Now().UTC()

	if _, err := h.repository.UpsertLearningResource(learningResource); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, response.Item{Item: learningResource})
}

// DeleteLearningResource :
func (h Handler) DeleteLearningResource(c echo.Context) error {
	learningResourceId := c.QueryParam("id")

	LearningResource, err := h.repository.FindLearningResourceByID(c.Request().Context(), learningResourceId)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, response.Exception{Code: errcode.RecordNotFound, Error: fmt.Errorf("learningResource %s not found", learningResourceId)})
		}
		return c.JSON(http.StatusInternalServerError, response.Exception{Code: errcode.ServerError, Error: err})
	}

	if _, err := h.repository.DeleteLearningResourceByID(LearningResource); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{Code: errcode.ServerError, Error: err})
	}

	return c.JSON(http.StatusOK, nil)
}

// CreateLearningResource :
func (h Handler) CreateLearningResource(c echo.Context) error {
	var i struct {
		Indicator string `json:"indicator" form:"indicator" validate:"required"`
		Name      string `json:"name" form:"name" validate:"required"`
		Title     string `json:"title" form:"title" validate:"required"`
		Link      string `json:"link" form:"link" validate:"required"`
		Type      string `json:"type" form:"type" validate:"required"`
		Source    string `json:"source" form:"source" validate:"required"`
	}

	// bind req input
	if err := c.Bind(&i); err != nil {
		return c.JSON(http.StatusBadRequest, response.Exception{Code: errcode.InvalidRequest, Error: err})
	}

	// cleanup
	i.Link = strings.TrimSpace(i.Link)

	// validate
	if err := c.Validate(&i); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.Exception{Code: errcode.ValidationError, Error: err})
	}

	// create time object
	timeNow := time.Now().UTC()

	// create LearningResource object
	learningResourceID := primitive.NewObjectID()

	learningResource := entity.LearningResource{
		ID:        &learningResourceID,
		Indicator: i.Indicator,
		Name:      i.Name,
		Title:     i.Title,
		Link:      i.Link,
		Type:      i.Type,
		Source:    i.Source,
		Status:    entity.StatusActive,
		Model: entity.Model{
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		},
	}

	if _, err := h.repository.CreateLearningResource(c.Request().Context(), learningResource); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, response.Item{Item: learningResource})
}

// ValidateLearningResource :
func ValidateLearningResource(h Handler, ctx context.Context, id string) (*entity.LearningResource, int, *response.Exception) {
	// check if LearningResource exists
	learningResource, err := h.repository.FindLearningResourceByID(ctx, id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &entity.LearningResource{}, http.StatusNotFound, &response.Exception{Code: errcode.RecordNotFound, Error: fmt.Errorf("learningResource %s not found", id)}
		}
		return &entity.LearningResource{}, http.StatusInternalServerError, &response.Exception{Code: errcode.ServerError, Error: err}
	}

	if learningResource.Status != entity.StatusActive {
		return &entity.LearningResource{}, http.StatusBadRequest, &response.Exception{Code: errcode.StatusInActive, Error: fmt.Errorf("learningResource %s status inactive", id)}
	}
	return learningResource, http.StatusOK, nil
}
