package handler

import (
	"context"
	"fmt"
	"net/http"
	"sme-api/app/entity"
	"sme-api/app/kit/general"
	"sme-api/app/repository"
	"sme-api/app/response"
	"sme-api/app/response/errcode"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetAssessments :
func (h Handler) GetAssessments(c echo.Context) error {
	assessments, cursor, err := h.repository.FindAssessments(c.Request().Context(), repository.FindAssessmentFilter{
		Cursor:      c.QueryParam("cursor"),
		IDs:         c.Request().URL.Query()["id"],
		SMEID:       c.QueryParam("smeId"),
		SharedWiths: c.Request().URL.Query()["sharedWith"],
		Status:      entity.Status(c.QueryParam("status")),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, response.Items{Items: assessments, Cursor: cursor, Count: len(assessments)})
}

// DeleteAssessment :
func (h Handler) DeleteAssessment(c echo.Context) error {
	assessmentId := c.QueryParam("id")

	assessment, httpStatus, exception := ValidateAssessment(h, c.Request().Context(), assessmentId)
	if exception != nil {
		return c.JSON(httpStatus, exception)
	}

	if _, err := h.repository.DeleteAssessmentByID(assessment); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{Code: errcode.ServerError, Error: err})
	}

	// delete entries under assessment
	if _, err := h.repository.BulkDeleteAssessmentEntries(c.Request().Context(), []string{assessmentId}); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{Code: errcode.ServerError, Error: err})
	}

	return c.JSON(http.StatusOK, nil)
}

// UpdateAssessment :
func (h Handler) UpdateAssessment(c echo.Context) error {
	assessmentId := c.QueryParam("id")
	var i struct {
		SerialNo    string `json:"serialNo" form:"serialNo"`
		SharedWiths []struct {
			CorporateID string        `json:"corporateID" form:"corporateID" validate:"max=50"`
			LinkDate    time.Time     `json:"linkDate" form:"linkDate"`
			Status      entity.Status `json:"status" form:"status" validate:"eq=INVITED|eq=ACTIVE"`
		} `json:"sharedWiths" form:"sharedWiths" validate:"dive,max=50"`
		CompletionDate time.Time     `json:"completionDate" form:"completionDate"`
		Status         entity.Status `json:"status" form:"status"`
	}

	// bind req input
	if err := c.Bind(&i); err != nil {
		return c.JSON(http.StatusBadRequest, response.Exception{Code: errcode.InvalidRequest, Error: err})
	}

	// validate
	if err := c.Validate(&i); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.Exception{Code: errcode.ValidationError, Error: err})
	}

	// check if assessment exists
	assessment, httpStatus, exception := ValidateAssessment(h, c.Request().Context(), assessmentId)
	if exception != nil {
		return c.JSON(httpStatus, exception)
	}

	if i.SerialNo != "" {
		assessment.SerialNo = strings.TrimSpace(i.SerialNo)
	}

	if i.Status != "" {
		assessment.Status = i.Status
	}

	if !i.CompletionDate.IsZero() {
		assessment.CompletionDate = i.CompletionDate
	}

	sharedWiths := []entity.LinkCorporate{}
	if len(i.SharedWiths) > 0 {
		for _, sharedWith := range i.SharedWiths {
			_, httpStatus, exception := ValidateSME(h, c.Request().Context(), sharedWith.CorporateID)
			if exception != nil {
				return c.JSON(httpStatus, exception)
			}

			corporateID, err := primitive.ObjectIDFromHex(sharedWith.CorporateID)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, response.Exception{
					Code:  errcode.ServerError,
					Error: err,
				})
			}
			sharedWiths = append(sharedWiths, entity.LinkCorporate{
				CorporateID: &corporateID,
				LinkDate:    sharedWith.LinkDate,
				Status:      sharedWith.Status,
			})
		}
	}
	assessment.SharedWiths = sharedWiths

	// check if report document is uploaded
	if report, _ := c.FormFile("report"); report != nil {
		// create path
		path := fmt.Sprintf("esg/%s/report/", strings.TrimSpace(assessment.SMEID.Hex()))

		// push to s3 bucket
		// url, httpStatus, exception := aws.PushDocBucket(path, report)
		// if exception != nil {
		// 	return c.JSON(httpStatus, exception)
		// }
		exception := general.ReadFile(report, path)
		if exception != nil {
			return c.JSON(http.StatusInternalServerError, exception)
		}
		assessment.Report = path + report.Filename
	}

	// check if ISO document is uploaded
	form, err := c.MultipartForm()
	if err == nil {
		isoDocs := form.File["isoDocs"]

		for _, doc := range isoDocs {
			// create path
			path := fmt.Sprintf("esg/%s/iso/", strings.TrimSpace(assessment.SMEID.Hex()))

			// push to s3 bucket
			// url, httpStatus, exception := aws.PushDocBucket(path, doc)
			// if exception != nil {
			// 	return c.JSON(httpStatus, exception)
			// }
			exception := general.ReadFile(doc, path)
			if exception != nil {
				return c.JSON(http.StatusInternalServerError, exception)
			}
			assessment.ISODocs = append(assessment.ISODocs, path+doc.Filename)
		}
	}

	// create time object
	assessment.Model.UpdatedAt = time.Now().UTC()

	if _, err = h.repository.UpsertAssessment(assessment); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, nil)
}

// CreateAssessment :
func (h Handler) CreateAssessment(c echo.Context) error {
	var i struct {
		SMEID         string `json:"smeID" form:"smeID" validate:"required,max=50"`
		QuestionSetID string `json:"questionSetID" form:"questionSetID" validate:"required,max=50"`
		SerialNo      string `json:"serialNo" form:"serialNo" validate:""`
	}

	// bind req input
	if err := c.Bind(&i); err != nil {
		return c.JSON(http.StatusBadRequest, response.Exception{Code: errcode.InvalidRequest, Error: err})
	}

	// cleanup
	i.SMEID = strings.TrimSpace(i.SMEID)
	i.SerialNo = strings.TrimSpace(i.SerialNo)
	i.QuestionSetID = strings.TrimSpace(i.QuestionSetID)

	// validate
	if err := c.Validate(&i); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.Exception{Code: errcode.ValidationError, Error: err})
	}

	// check if sme exists
	sme, httpStatus, exception := ValidateSME(h, c.Request().Context(), i.SMEID)
	if exception != nil {
		return c.JSON(httpStatus, exception)
	}

	// check if question set exists
	questionSet, httpStatus, exception := ValidateQuestionSet(h, c.Request().Context(), i.QuestionSetID)
	if exception != nil {
		return c.JSON(httpStatus, exception)
	}

	// create time object
	timeNow := time.Now().UTC()

	// create assessment object
	assessmentID := primitive.NewObjectID()
	smeID, err := primitive.ObjectIDFromHex(i.SMEID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	questionSetID, err := primitive.ObjectIDFromHex(i.QuestionSetID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	assessment := entity.Assessment{
		ID:             &assessmentID,
		SMEID:          &smeID,
		QuestionSetID:  &questionSetID,
		SerialNo:       i.SerialNo,
		CompletionDate: time.Time{},
		Status:         entity.StatusActive,
		Model: entity.Model{
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		},
	}

	if _, err = h.repository.CreateAssessment(c.Request().Context(), assessment); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	assessmentEntries, httpStatus, exception := generateAssessmentEntries(c.Request().Context(), h, AssessmentEntryInput{
		AssessmentID:  assessment.ID,
		QuestionSetID: questionSet.ID,
		SMEID:         sme.ID,
	})
	if exception != nil {
		return c.JSON(httpStatus, exception)
	}

	// bulk insert entries
	_, err = h.repository.BulkInsertAssessmentEntries(c.Request().Context(), assessmentEntries)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}
	return c.JSON(http.StatusOK, response.Item{Item: assessment.ID})
}

// ValidateAssessment :
func ValidateAssessment(h Handler, ctx context.Context, ID string) (*entity.Assessment, int, *response.Exception) {
	// check if assessment exists
	assessment, err := h.repository.FindAssessmentByID(ctx, ID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &entity.Assessment{}, http.StatusNotFound, &response.Exception{Code: errcode.RecordNotFound, Error: fmt.Errorf("assessment %s not found", ID)}
		}
		return &entity.Assessment{}, http.StatusInternalServerError, &response.Exception{Code: errcode.ServerError, Error: err}
	}

	if assessment.Status != entity.StatusActive {
		return &entity.Assessment{}, http.StatusBadRequest, &response.Exception{Code: errcode.StatusInActive, Error: fmt.Errorf("assessment %s status inactive", ID)}
	}
	return assessment, http.StatusOK, nil
}
