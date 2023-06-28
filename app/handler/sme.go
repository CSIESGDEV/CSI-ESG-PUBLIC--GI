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

// GetSMEs :
func (h Handler) GetSMEs(c echo.Context) error {
	smes, cursor, err := h.repository.FindSMEs(c.Request().Context(), repository.FindSMEFilter{
		Cursor:      c.QueryParam("cursor"),
		IDs:         c.Request().URL.Query()["id"],
		LinkedWiths: c.Request().URL.Query()["linkedWith"],
		SSMNumbers:  c.Request().URL.Query()["ssmNumber"],
		CompanyName: c.QueryParam("companyName"),
		Status:      entity.Status(c.QueryParam("status")),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, response.Items{Items: smes, Cursor: cursor, Count: len(smes)})
}

// UpdateSME :
func (h Handler) UpdateSME(c echo.Context) error {
	smeId := c.QueryParam("id")
	var i struct {
		CompanyName          string   `json:"companyName" form:"companyName" validate:"max=50"`
		SSMNumber            string   `json:"ssmNumber" form:"ssmNumber" validate:"max=12"`
		BusinessEntity       string   `json:"businessEntity" form:"businessEntity"`
		RegisteredInEastMY   bool     `json:"registeredInEastMy" form:"registeredInEastMy"`
		EducationType        string   `json:"educationType" form:"educationType"`
		State                string   `json:"state" form:"state" validate:"max=40"`
		PostCode             string   `json:"postCode" form:"postCode"  validate:"max=5"`
		MSIC                 string   `json:"msic" form:"msic" validate:"msic,max=10"`
		Industry             string   `json:"industry" form:"industry"  validate:"max=40"`
		ApprovedBy           string   `json:"approvedBy" form:"approvedBy"`
		ParticipatedLearning []string `json:"participatedLearning" form:"participatedLearning"`
		LinkedWiths          []struct {
			CorporateID string        `bson:"corporateID" json:"corporateID" form:"corporateID" validate:"max=50"`
			LinkDate    time.Time     `bson:"linkDate" json:"linkDate" form:"linkDate"`
			Status      entity.Status `bson:"status" json:"status" form:"status" validate:"eq=INVITED|eq=ACTIVE|eq=INACTIVE"`
		} `json:"linkedWiths" form:"linkedWiths" validate:"dive,max=50,required"`
		Status entity.Status `json:"status" form:"status"`
	}

	// bind req input
	if err := c.Bind(&i); err != nil {
		return c.JSON(http.StatusBadRequest, response.Exception{Code: errcode.InvalidRequest, Error: err})
	}

	// cleanup
	i.PostCode = strings.TrimSpace(i.PostCode)
	i.SSMNumber = strings.TrimSpace(i.SSMNumber)

	// validate
	if err := c.Validate(&i); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.Exception{Code: errcode.ValidationError, Error: err})
	}

	// check if SME exists
	sme, err := h.repository.FindSMEByID(c.Request().Context(), smeId)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, response.Exception{Code: errcode.RecordNotFound, Error: fmt.Errorf("SME %s not found", smeId)})
		}
		return c.JSON(http.StatusInternalServerError, response.Exception{Code: errcode.ServerError, Error: err})
	}

	if i.CompanyName != "" {
		sme.CompanyName = i.CompanyName
	}

	if i.SSMNumber != "" {
		// check if SME exists
		if smes, _, err := h.repository.FindSMEs(c.Request().Context(), repository.FindSMEFilter{
			SSMNumbers: []string{i.SSMNumber},
		}); err == nil && len(smes) > 0 {
			return c.JSON(http.StatusBadRequest, response.Exception{
				Code:  errcode.SSMNumberFound,
				Error: err,
			})
		}
		sme.SSMNumber = i.SSMNumber
	}

	if i.BusinessEntity != "" {
		sme.BusinessEntity = entity.BusinessEntity(i.BusinessEntity)
	}

	if i.RegisteredInEastMY != sme.RegisteredInEastMY {
		sme.RegisteredInEastMY = i.RegisteredInEastMY
	}

	if i.EducationType != "" {
		sme.EducationType = i.EducationType
	}

	if i.State != "" {
		sme.State = i.State
	}

	if i.PostCode != "" {
		sme.PostCode = i.PostCode
	}

	if i.MSIC != "" {
		sme.MSIC = i.MSIC
	}

	if i.ApprovedBy != "" {
		// check if admin exists
		_, _, err := h.repository.FindAdmins(c.Request().Context(), repository.FindAdminFilter{
			IDs: []string{i.ApprovedBy},
		})
		if err != nil {
			return c.JSON(http.StatusBadRequest, response.Exception{
				Code:  errcode.AdminNotFound,
				Error: err,
			})
		}
		approvedBy, err := primitive.ObjectIDFromHex(i.ApprovedBy)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, response.Exception{
				Code:  errcode.ServerError,
				Error: err,
			})
		}
		sme.ApprovedBy = &approvedBy
	}

	linkedWiths := []entity.LinkCorporate{}
	if len(i.LinkedWiths) > 0 {
		for _, linkwith := range i.LinkedWiths {
			_, httpStatus, exception := ValidateSME(h, c.Request().Context(), linkwith.CorporateID)
			if exception != nil {
				return c.JSON(httpStatus, exception)
			}

			corporateID, err := primitive.ObjectIDFromHex(linkwith.CorporateID)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, response.Exception{
					Code:  errcode.ServerError,
					Error: err,
				})
			}

			linkedWiths = append(linkedWiths, entity.LinkCorporate{
				CorporateID: &corporateID,
				LinkDate:    linkwith.LinkDate,
				Status:      linkwith.Status,
			})
		}
	}
	sme.LinkedWiths = linkedWiths

	if len(i.ParticipatedLearning) > 0 {
		participatedLearning := []*primitive.ObjectID{}
		for _, id := range i.ParticipatedLearning {
			id := strings.TrimSpace(id)
			// check if resource already linked with sme
			resource, httpStatus, exception := ValidateLearningResource(h, c.Request().Context(), id)
			if exception != nil {
				return c.JSON(httpStatus, exception)
			}
			participatedLearning = append(participatedLearning, resource.ID)
		}
		sme.ParticipatedLearning = participatedLearning
	}

	if i.Status != "" {
		sme.Status = i.Status
	}

	// check if SSM document is uploaded
	if ssmDoc, _ := c.FormFile("ssmDoc"); ssmDoc != nil {
		// create path
		path := fmt.Sprintf("sme/%s/ssm/", strings.TrimSpace(smeId))

		// push to s3 bucket
		// url, httpStatus, exception := aws.PushDocBucket(path, ssmDoc)
		// if exception != nil {
		// 	return c.JSON(httpStatus, exception)
		// }
		exception := general.ReadFile(ssmDoc, path)
		if exception != nil {
			return c.JSON(http.StatusInternalServerError, exception)
		}
		sme.SSMDoc = path + ssmDoc.Filename
	}

	// check if profile picture is uploaded
	if profilePicture, _ := c.FormFile("profilePicture"); profilePicture != nil {
		// create path
		path := fmt.Sprintf("sme/%s/profile/", strings.TrimSpace(smeId))

		// push to s3 bucket
		// url, httpStatus, exception := aws.PushDocBucket(path, profilePicture)
		// if exception != nil {
		// 	return c.JSON(httpStatus, exception)
		// }
		exception := general.ReadFile(profilePicture, path)
		if exception != nil {
			return c.JSON(http.StatusInternalServerError, exception)
		}
		sme.ProfilePicture = path + profilePicture.Filename
	}

	sme.Model.UpdatedAt = time.Now().UTC()

	if _, err := h.repository.UpsertSME(sme); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, response.Item{Item: sme})
}

// DeleteSME :
func (h Handler) DeleteSME(c echo.Context) error {
	smeId := c.QueryParam("id")

	sme, err := h.repository.FindSMEByID(c.Request().Context(), smeId)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, response.Exception{Code: errcode.RecordNotFound, Error: fmt.Errorf("SME %s not found", smeId)})
		}
		return c.JSON(http.StatusInternalServerError, response.Exception{Code: errcode.ServerError, Error: err})
	}

	if _, err := h.repository.DeleteSMEByID(sme); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{Code: errcode.ServerError, Error: err})
	}

	return c.JSON(http.StatusOK, nil)
}

// CreateSME :
func (h Handler) CreateSME(c echo.Context) error {
	var i struct {
		CompanyName        string `json:"companyName" form:"companyName" validate:"required,max=50"`
		SSMNumber          string `json:"ssmNumber" form:"ssmNumber" validate:"max=24"`
		BusinessEntity     string `json:"businessEntity" form:"businessEntity"`
		RegisteredInEastMY bool   `json:"registeredInEastMy" form:"registeredInEastMy" validate:""`
		EducationType      string `json:"educationType" form:"educationType" validate:""`
		State              string `json:"state" form:"state" validate:"required,max=40"`
		PostCode           string `json:"postCode" form:"postCode" validate:"required,max=5"`
		MSIC               string `json:"msic" form:"msic" validate:"msic,max=10"`
		ApprovedBy         string `json:"approvedBy" form:"approvedBy"`
	}

	// bind req input
	if err := c.Bind(&i); err != nil {
		return c.JSON(http.StatusBadRequest, response.Exception{Code: errcode.InvalidRequest, Error: err})
	}

	// cleanup
	i.PostCode = strings.TrimSpace(i.PostCode)
	i.SSMNumber = strings.TrimSpace(i.SSMNumber)

	// validate
	if err := c.Validate(&i); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.Exception{Code: errcode.ValidationError, Error: err})
	}

	// check if SME exists
	smes, _, err := h.repository.FindSMEs(c.Request().Context(), repository.FindSMEFilter{
		SSMNumbers: []string{i.SSMNumber},
	})
	if err == nil && len(smes) > 0 {
		return c.JSON(http.StatusBadRequest, response.Exception{
			Code:  errcode.SSMNumberFound,
			Error: err,
		})
	}

	if i.ApprovedBy != "" {
		// check if admin exists
		_, _, err := h.repository.FindAdmins(c.Request().Context(), repository.FindAdminFilter{
			IDs: []string{i.ApprovedBy},
		})
		if err != nil {
			return c.JSON(http.StatusBadRequest, response.Exception{
				Code:  errcode.AdminNotFound,
				Error: err,
			})
		}
	}

	// create time object
	timeNow := time.Now().UTC()

	// create SME object
	smeID := primitive.NewObjectID()

	sme := entity.SME{
		ID:                 &smeID,
		CompanyName:        i.CompanyName,
		SSMNumber:          i.SSMNumber,
		BusinessEntity:     entity.BusinessEntity(i.BusinessEntity),
		RegisteredInEastMY: i.RegisteredInEastMY,
		EducationType:      i.EducationType,
		State:              i.State,
		PostCode:           i.PostCode,
		MSIC:               i.MSIC,
		Status:             entity.StatusActive,
		Model: entity.Model{
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		},
	}

	if i.ApprovedBy != "" {
		approvedBy, err := primitive.ObjectIDFromHex(i.ApprovedBy)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, response.Exception{
				Code:  errcode.ServerError,
				Error: err,
			})
		}
		sme.ApprovedBy = &approvedBy
	}

	// check if SSM document is uploaded
	if ssmDoc, _ := c.FormFile("ssmDoc"); ssmDoc != nil {
		// create path
		path := fmt.Sprintf("sme/%s/ssm/", strings.TrimSpace(smeID.Hex()))

		// push to s3 bucket
		// url, httpStatus, exception := aws.PushDocBucket(path, ssmDoc)
		// if exception != nil {
		// 	return c.JSON(httpStatus, exception)
		// }
		exception := general.ReadFile(ssmDoc, path)
		if exception != nil {
			return c.JSON(http.StatusInternalServerError, exception)
		}
		sme.SSMDoc = path + ssmDoc.Filename
	}

	// check if profile picture is uploaded
	if profilePicture, _ := c.FormFile("profilePicture"); profilePicture != nil {
		// create path
		path := fmt.Sprintf("sme/%s/profile/", strings.TrimSpace(smeID.Hex()))

		// push to s3 bucket
		// url, httpStatus, exception := aws.PushDocBucket(path, profilePicture)
		// if exception != nil {
		// 	return c.JSON(httpStatus, exception)
		// }
		exception := general.ReadFile(profilePicture, path)
		if exception != nil {
			return c.JSON(http.StatusInternalServerError, exception)
		}
		sme.ProfilePicture = path + profilePicture.Filename
	}

	if _, err = h.repository.CreateSME(c.Request().Context(), sme); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, response.Item{Item: sme})
}

// ValidateSME :
func ValidateSME(h Handler, ctx context.Context, companyID string) (*entity.SME, int, *response.Exception) {
	// check if SME exists
	sme, err := h.repository.FindSMEByID(ctx, companyID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &entity.SME{}, http.StatusNotFound, &response.Exception{Code: errcode.RecordNotFound, Error: fmt.Errorf("SME %s not found", companyID)}
		}
		return &entity.SME{}, http.StatusInternalServerError, &response.Exception{Code: errcode.ServerError, Error: err}
	}

	if sme.Status != entity.StatusActive {
		return &entity.SME{}, http.StatusBadRequest, &response.Exception{Code: errcode.StatusInActive, Error: fmt.Errorf("SME %s status inactive", companyID)}
	}
	return sme, http.StatusOK, nil
}
