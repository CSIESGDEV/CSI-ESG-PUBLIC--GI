package handler

import (
	"context"
	"fmt"
	"net/http"
	"sme-api/app/entity"
	"sme-api/app/repository"
	"sme-api/app/response"
	"sme-api/app/response/errcode"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetQuestions :
func (h Handler) GetQuestions(c echo.Context) error {
	questions, cursor, err := h.repository.FindQuestions(c.Request().Context(), repository.FindQuestionFilter{
		Cursor:        c.QueryParam("cursor"),
		IDs:           c.Request().URL.Query()["id"],
		QuestionSetID: c.QueryParam("questionSetId"),
		Dimension:     c.QueryParam("dimension"),
		// SubCategory:   c.QueryParam("subCategory"),
		// Indicator:     c.QueryParam("indicator"),
		QuestionType: c.QueryParam("questionType"),
		Status:       entity.Status(c.QueryParam("status")),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, response.Items{Items: questions, Cursor: cursor, Count: len(questions)})
}

// UpdateQuestion :
func (h Handler) UpdateQuestion(c echo.Context) error {
	questionId := c.QueryParam("id")
	var i struct {
		QuestionSetID string `json:"questionSetId" form:"questionSetId" validate:"max=50"`
		Dimension     string `json:"dimension" form:"dimension"`
		// SubCategory   string              `json:"subCategory" form:"subCategory"`
		// Indicator     string              `json:"indicator" form:"indicator"`
		QuestionLabel string              `json:"questionLabel" form:"questionLabel"`
		QuestionType  entity.QuestionType `json:"questionType" form:"questionType"`
		Options       []string            `json:"options" form:"options"`
		Status        entity.Status       `json:"status" form:"status"`
	}

	// bind req input
	if err := c.Bind(&i); err != nil {
		return c.JSON(http.StatusBadRequest, response.Exception{Code: errcode.InvalidRequest, Error: err})
	}

	// clean up
	i.QuestionSetID = strings.TrimSpace(i.QuestionSetID)

	// validate
	if err := c.Validate(&i); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.Exception{Code: errcode.ValidationError, Error: err})
	}

	// check if Question exists
	question, err := h.repository.FindQuestionByID(c.Request().Context(), questionId)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, response.Exception{Code: errcode.RecordNotFound, Error: fmt.Errorf("question %s not found", questionId)})
		}
		return c.JSON(http.StatusInternalServerError, response.Exception{Code: errcode.ServerError, Error: err})
	}

	if i.QuestionSetID != "" {
		if _, err := h.repository.FindQuestionSetByID(c.Request().Context(), i.QuestionSetID); err != nil {
			if err == mongo.ErrNoDocuments {
				return c.JSON(http.StatusNotFound, response.Exception{Code: errcode.RecordNotFound, Error: fmt.Errorf("question set %s not found", i.QuestionSetID)})
			}
			return c.JSON(http.StatusInternalServerError, response.Exception{Code: errcode.ServerError, Error: err})
		}
		questionSetID, err := primitive.ObjectIDFromHex(i.QuestionSetID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, response.Exception{
				Code:  errcode.ServerError,
				Error: err,
			})
		}
		question.QuestionSetID = &questionSetID
	}

	if i.Dimension != "" {
		question.Dimension = i.Dimension
	}

	// if i.SubCategory != "" {
	// 	question.SubCategory = i.SubCategory
	// }

	// if i.Indicator != "" {
	// 	question.Indicator = i.Indicator
	// }

	if i.QuestionLabel != "" {
		question.QuestionLabel = i.QuestionLabel
	}

	if i.QuestionType != "" {
		question.QuestionType = i.QuestionType
	}

	if len(i.Options) > 0. {
		question.Options = i.Options
	}

	if i.Status != "" {
		question.Status = i.Status
	}

	question.Model.UpdatedAt = time.Now().UTC()

	if _, err := h.repository.UpsertQuestion(question); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, response.Item{Item: question})
}

// DeleteQuestion :
func (h Handler) DeleteQuestion(c echo.Context) error {
	questionId := c.QueryParam("id")

	question, err := h.repository.FindQuestionByID(c.Request().Context(), questionId)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, response.Exception{Code: errcode.RecordNotFound, Error: fmt.Errorf("question %s not found", questionId)})
		}
		return c.JSON(http.StatusInternalServerError, response.Exception{Code: errcode.ServerError, Error: err})
	}

	if _, err := h.repository.DeleteQuestionByID(question); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{Code: errcode.ServerError, Error: err})
	}

	return c.JSON(http.StatusOK, nil)
}

// CreateQuestion :
func (h Handler) CreateQuestion(c echo.Context) error {
	questions := make([]*entity.Question, 0)
	type question struct {
		QuestionSetID string `json:"questionSetId" form:"questionSetId" validate:"required,max=50"`
		Dimension     string `json:"dimension" form:"dimension" validate:"required"`
		// SubCategory   string              `json:"subCategory" form:"subCategory" validate:"required"`
		// Indicator     string              `json:"indicator" form:"indicator" validate:"required"`
		QuestionLabel string              `json:"questionLabel" form:"questionLabel" validate:"required"`
		QuestionType  entity.QuestionType `json:"questionType" form:"questionType" validate:"required"`
		Options       []string            `json:"options" form:"options"`
	}

	var i struct {
		Responses []*question `json:"responses" form:"responses" validate:"gt=0,dive,required"`
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

	for _, res := range i.Responses {
		// clean up
		res.QuestionSetID = strings.TrimSpace(res.QuestionSetID)
		// check if question set exists
		if _, err := h.repository.FindQuestionSetByID(c.Request().Context(), res.QuestionSetID); err != nil {
			if err == mongo.ErrNoDocuments {
				return c.JSON(http.StatusNotFound, response.Exception{Code: errcode.RecordNotFound, Error: fmt.Errorf("question set %s not found", res.QuestionSetID)})
			}
			return c.JSON(http.StatusInternalServerError, response.Exception{Code: errcode.ServerError, Error: err})
		}

		// create Question object
		questionID := primitive.NewObjectID()
		questionSetID, err := primitive.ObjectIDFromHex(res.QuestionSetID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, response.Exception{
				Code:  errcode.ServerError,
				Error: err,
			})
		}

		question := entity.Question{
			ID:            &questionID,
			QuestionSetID: &questionSetID,
			Dimension:     res.Dimension,
			// SubCategory:   res.SubCategory,
			// Indicator:     res.Indicator,
			QuestionLabel: res.QuestionLabel,
			QuestionType:  res.QuestionType,
			Options:       res.Options,
			Status:        entity.StatusActive,
			Model: entity.Model{
				CreatedAt: timeNow,
				UpdatedAt: timeNow,
			},
		}

		// if res.QuestionType == entity.QuestionTypeDemographic {
		// 	question.Options = res.Options
		// }

		questions = append(questions, &question)
	}

	result, err := h.repository.BulkInsertQuestions(c.Request().Context(), questions)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, response.Item{Item: result})
}

// ValidateQuestion :
func ValidateQuestion(h Handler, ctx context.Context, id string) (*entity.Question, int, *response.Exception) {
	// check if Question exists
	questions, err := h.repository.FindQuestionByID(ctx, id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &entity.Question{}, http.StatusNotFound, &response.Exception{Code: errcode.RecordNotFound, Error: fmt.Errorf("question %s not found", id)}
		}
		return &entity.Question{}, http.StatusInternalServerError, &response.Exception{Code: errcode.ServerError, Error: err}
	}

	if questions.Status != entity.StatusActive {
		return &entity.Question{}, http.StatusBadRequest, &response.Exception{Code: errcode.StatusInActive, Error: fmt.Errorf("question %s status inactive", id)}
	}
	return questions, http.StatusOK, nil
}
