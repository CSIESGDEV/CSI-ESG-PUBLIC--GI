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

// AssessmentEntryInput :
type AssessmentEntryInput struct {
	AssessmentID  *primitive.ObjectID `json:"assessmentID" form:"assessmentID" validate:"required,max=50"`
	QuestionSetID *primitive.ObjectID `json:"questionSetID" form:"questionSetID" validate:"required,max=50"`
	SMEID         *primitive.ObjectID `json:"smeID" form:"smeID" validate:"required,max=50"`
}

// GetAssessmentEntries :
func (h Handler) GetAssessmentEntries(c echo.Context) error {
	assessmentEntries, cursor, err := h.repository.FindAssessmentEntries(c.Request().Context(), repository.FindAssessmentEntryFilter{
		Cursor:        c.QueryParam("cursor"),
		IDs:           c.Request().URL.Query()["id"],
		QuestionType:  c.QueryParam("questionType"),
		QuestionSetID: c.QueryParam("questionSetId"),
		AssessmentID:  c.QueryParam("assessmentId"),
		RespondStatus: entity.RespondStatus(c.QueryParam("respondStatus")),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, response.Items{Items: assessmentEntries, Cursor: cursor, Count: len(assessmentEntries)})
}

// SubmitAssessment : switch question status to SUBMITTED
func (h Handler) SubmitAssessment(c echo.Context) error {
	smeUserData := c.Get("SME_ADMIN")
	smeUser := smeUserData.(entity.SMEUser)
	assessmentID := c.QueryParam("assessmentId")

	assessmentOID, err := primitive.ObjectIDFromHex(assessmentID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{Code: errcode.ServerError, Error: err})
	}

	// check if assessment exist
	assessment, httpStatus, exception := ValidateAssessment(h, c.Request().Context(), assessmentID)
	if exception != nil {
		return c.JSON(httpStatus, exception)
	}

	if _, err := h.repository.SubmitAssessment(c.Request().Context(), repository.SubmitAssessmentFilter{
		AssessmentID: assessmentOID.Hex(),
		SMEIDs:       []string{smeUser.CompanyID.Hex()},
	}); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{Code: errcode.ServerError, Error: err})
	}

	// update assessment completion date
	assessment.CompletionDate = time.Now().UTC()
	assessment.Model.UpdatedAt = time.Now().UTC()

	if _, err := h.repository.UpsertAssessment(assessment); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{Code: errcode.ServerError, Error: err})
	}

	return c.JSON(http.StatusOK, nil)
}

// SubmitResponse :
func (h Handler) SubmitResponse(c echo.Context) error {
	smeUserData := c.Get("SME_ADMIN")
	smeUser := smeUserData.(entity.SMEUser)

	var i struct {
		Responses []struct {
			AssessmentEntryID *primitive.ObjectID `json:"assessmentEntryID" form:"assessmentEntryID" validate:"required,max=50"`
			Answer            entity.Answer       `json:"answer" validate:"required"`
		} `json:"responses" form:"responses" validate:"gt=0,dive,required"`
	}

	// bind req input
	if err := c.Bind(&i); err != nil {
		return c.JSON(http.StatusBadRequest, response.Exception{Code: errcode.InvalidRequest, Error: err})
	}

	// validate
	if err := c.Validate(&i); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.Exception{Code: errcode.ValidationError, Error: err})
	}

	timeNow := time.Now()
	entries := make([]*entity.AssessmentEntry, 0)
	for _, res := range i.Responses {
		entry, err := h.repository.FindAssessmentEntryByID(c.Request().Context(), res.AssessmentEntryID.Hex())
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return c.JSON(http.StatusNotFound, response.Exception{Code: errcode.RecordNotFound, Error: fmt.Errorf("assessment entry %s not found", res.AssessmentEntryID.Hex())})
			}
			return c.JSON(http.StatusInternalServerError, response.Exception{Code: errcode.ServerError, Error: err})
		}
		if entry.SMEID.Hex() != smeUser.CompanyID.Hex() {
			return c.JSON(http.StatusBadRequest, response.Exception{Code: errcode.MismatchSME, Error: fmt.Errorf("mismatch sme id")})
		}

		if entry.RespondStatus != entity.ResponseStatusToSubmitted {
			entry.Answer = res.Answer
			entry.RespondStatus = entity.ResponseStatusInProgress
			entry.SubmittedDateTime = timeNow
			entries = append(entries, entry)
		}
	}

	var updateResult interface{}
	var err error
	if len(entries) != 0 {
		updateResult, err = h.repository.BulkWriteAssessmentEntries(c.Request().Context(), entries)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, response.Exception{Code: errcode.ServerError, Error: err})
		}
		// return c.JSON(http.StatusOK, response.Item{Item: updateResult})
	}

	// return c.JSON(http.StatusOK, response.Item{Item: "No questions have been saved"})
	return c.JSON(http.StatusOK, response.Item{Item: updateResult})
}

// ValidateAssessmentEntry :
func ValidateAssessmentEntry(h Handler, ctx context.Context, id string) (*entity.AssessmentEntry, int, *response.Exception) {
	// check if AssessmentEntry exists
	assessmentEntry, err := h.repository.FindAssessmentEntryByID(ctx, id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &entity.AssessmentEntry{}, http.StatusNotFound, &response.Exception{Code: errcode.RecordNotFound, Error: fmt.Errorf("assessmentEntrie %s not found", id)}
		}
		return &entity.AssessmentEntry{}, http.StatusInternalServerError, &response.Exception{Code: errcode.ServerError, Error: err}
	}

	if assessmentEntry.RespondStatus != entity.ResponseStatusToStart {
		return &entity.AssessmentEntry{}, http.StatusBadRequest, &response.Exception{Code: errcode.StatusInActive, Error: fmt.Errorf("assessmentEntrie %s status inactive", id)}
	}
	return assessmentEntry, http.StatusOK, nil
}

// generateAssessmentEntries : given a assessmentID, questionSetID, return a list of entries with different questions
func generateAssessmentEntries(c context.Context, h Handler, i AssessmentEntryInput) ([]*entity.AssessmentEntry, int, *response.Exception) {
	assessment, httpStatus, exception := ValidateAssessment(h, c, i.AssessmentID.Hex())
	if exception != nil {
		return nil, httpStatus, exception
	}

	questionSet, httpStatus, exception := ValidateQuestionSet(h, c, i.QuestionSetID.Hex())
	if exception != nil {
		return nil, httpStatus, exception
	}

	sme, httpStatus, exception := ValidateSME(h, c, i.SMEID.Hex())
	if exception != nil {
		return nil, httpStatus, exception
	}

	// check if a assessment-question set pair already exist in a Assessment, if yes then skip
	existingEntry, _, err := h.repository.FindAssessmentEntries(c, repository.FindAssessmentEntryFilter{
		AssessmentID:  i.AssessmentID.Hex(),
		QuestionSetID: i.QuestionSetID.Hex(),
	})
	if err != nil {
		return nil, http.StatusInternalServerError, &response.Exception{Code: errcode.ServerError, Error: err}
	}
	if len(existingEntry) > 0 {
		return nil, http.StatusOK, nil
	}

	// get all questions in question set
	cursor := ""
	var assessmentQuestions []*entity.Question
	for {
		questions, nextCursor, err := h.repository.FindQuestions(c, repository.FindQuestionFilter{
			Cursor:        cursor,
			QuestionSetID: questionSet.ID.Hex(),
			Status:        entity.StatusActive,
		})
		if err != nil {
			return nil, http.StatusInternalServerError, &response.Exception{Code: errcode.ServerError, Error: err}
		}
		assessmentQuestions = append(assessmentQuestions, questions...)
		if nextCursor == "0" {
			break
		} else {
			cursor = nextCursor
		}
	}

	// make assessment entry
	var entires []*entity.AssessmentEntry
	for _, question := range assessmentQuestions {
		oid := primitive.NewObjectID()
		entry := &entity.AssessmentEntry{
			ID:                &oid,
			AssessmentID:      assessment.ID,
			SMEID:             sme.ID,
			QuestionSetID:     questionSet.ID,
			QuestionID:        question.ID,
			Question:          question,
			QuestionType:      question.QuestionType,
			Answer:            entity.Answer{},
			SubmittedDateTime: time.Time{},
			RespondStatus:     entity.ResponseStatusToStart,
		}
		entires = append(entires, entry)
	}

	return entires, http.StatusOK, nil
}
