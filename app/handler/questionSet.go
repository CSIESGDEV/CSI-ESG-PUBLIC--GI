package handler

import (
	"context"
	"csi-api/app/entity"
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

// GetQuestionSets :
func (h Handler) GetQuestionSets(c echo.Context) error {
	questionSets, cursor, err := h.repository.FindQuestionSets(c.Request().Context(), repository.FindQuestionSetFilter{
		Cursor: c.QueryParam("cursor"),
		IDs:    c.Request().URL.Query()["Id"],
		Label:  c.QueryParam("label"),
		Status: entity.Status(c.QueryParam("status")),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, response.Items{Items: questionSets, Cursor: cursor, Count: len(questionSets)})
}

// UpdateQuestionSet :
func (h Handler) UpdateQuestionSet(c echo.Context) error {
	questionSetId := c.QueryParam("id")
	var i struct {
		Label  string        `json:"label" form:"label" validate:"max=40"`
		Status entity.Status `json:"status" form:"status"`
	}

	// bind req input
	if err := c.Bind(&i); err != nil {
		return c.JSON(http.StatusBadRequest, response.Exception{Code: errcode.InvalidRequest, Error: err})
	}

	// validate
	if err := c.Validate(&i); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.Exception{Code: errcode.ValidationError, Error: err})
	}

	// check if QuestionSet exists
	questionSet, err := h.repository.FindQuestionSetByID(c.Request().Context(), questionSetId)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, response.Exception{Code: errcode.RecordNotFound, Error: fmt.Errorf("questionSet %s not found", questionSetId)})
		}
		return c.JSON(http.StatusInternalServerError, response.Exception{Code: errcode.ServerError, Error: err})
	}

	if i.Label != "" {
		// check if QuestionSet exists
		if questionSets, _, err := h.repository.FindQuestionSets(c.Request().Context(), repository.FindQuestionSetFilter{
			Label: i.Label,
		}); err == nil && len(questionSets) > 0 {
			return c.JSON(http.StatusBadRequest, response.Exception{
				Code:  errcode.RecordFound,
				Error: err,
			})
		}
		questionSet.Label = i.Label
	}

	if i.Status != "" {
		questionSet.Status = i.Status
	}

	questionSet.Model.UpdatedAt = time.Now().UTC()

	if _, err := h.repository.UpsertQuestionSet(questionSet); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, response.Item{Item: questionSet})
}

// DeleteQuestionSet :
func (h Handler) DeleteQuestionSet(c echo.Context) error {
	questionSetId := c.QueryParam("id")

	questionSet, err := h.repository.FindQuestionSetByID(c.Request().Context(), questionSetId)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, response.Exception{Code: errcode.RecordNotFound, Error: fmt.Errorf("questionSet %s not found", questionSetId)})
		}
		return c.JSON(http.StatusInternalServerError, response.Exception{Code: errcode.ServerError, Error: err})
	}

	if _, err := h.repository.DeleteQuestionSetByID(questionSet); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{Code: errcode.ServerError, Error: err})
	}

	return c.JSON(http.StatusOK, nil)
}

// CreateQuestionSet :
func (h Handler) CreateQuestionSet(c echo.Context) error {
	var i struct {
		Label string `json:"label" form:"label" validate:"required,max=40"`
	}

	// bind req input
	if err := c.Bind(&i); err != nil {
		return c.JSON(http.StatusBadRequest, response.Exception{Code: errcode.InvalidRequest, Error: err})
	}

	// validate
	if err := c.Validate(&i); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.Exception{Code: errcode.ValidationError, Error: err})
	}

	// check if question set exists
	questionSets, _, err := h.repository.FindQuestionSets(c.Request().Context(), repository.FindQuestionSetFilter{
		Label: i.Label,
	})
	if err == nil && len(questionSets) > 0 {
		return c.JSON(http.StatusBadRequest, response.Exception{
			Code:  errcode.RecordFound,
			Error: err,
		})
	}

	// create time object
	timeNow := time.Now().UTC()

	// create QuestionSet object
	questionSetID := primitive.NewObjectID()

	questionSet := entity.QuestionSet{
		ID:     &questionSetID,
		Label:  i.Label,
		Status: entity.StatusActive,
		Model: entity.Model{
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		},
	}

	if _, err = h.repository.CreateQuestionSet(c.Request().Context(), questionSet); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, response.Item{Item: questionSet})
}

// ValidateQuestionSet :
func ValidateQuestionSet(h Handler, ctx context.Context, id string) (*entity.QuestionSet, int, *response.Exception) {
	// check if QuestionSet exists
	questionSets, err := h.repository.FindQuestionSetByID(ctx, id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &entity.QuestionSet{}, http.StatusNotFound, &response.Exception{Code: errcode.RecordNotFound, Error: fmt.Errorf("questionSet %s not found", id)}
		}
		return &entity.QuestionSet{}, http.StatusInternalServerError, &response.Exception{Code: errcode.ServerError, Error: err}
	}

	if questionSets.Status != entity.StatusActive {
		return &entity.QuestionSet{}, http.StatusBadRequest, &response.Exception{Code: errcode.StatusInActive, Error: fmt.Errorf("questionSet %s status inactive", id)}
	}
	return questionSets, http.StatusOK, nil
}
