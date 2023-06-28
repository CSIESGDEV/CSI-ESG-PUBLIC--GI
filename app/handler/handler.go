package handler

import (
	"context"
	"net/http"
	"strings"

	"sme-api/app/kit/aws"
	"sme-api/app/repository"
	"sme-api/app/response"
	"sme-api/app/response/errcode"

	"sme-api/app/bootstrap"

	"github.com/labstack/echo/v4"
)

// Handler :
type Handler struct {
	repository *repository.Repository
}

// New :
func New(bs *bootstrap.Bootstrap) *Handler {
	return &Handler{
		repository: bs.Repository,
	}
}

// APIHealthCheckVersion1 :
func (h Handler) APIHealthCheckVersion1(c echo.Context) error {
	if err := h.repository.HealthCheck(context.Background()); err != nil {
		return err
	}
	return c.Render(http.StatusOK, "info", "SME API IS FINE")
}

// Create signed url for S3 bucket :
func (h Handler) GetSignedUrl(c echo.Context) error {
	var i struct {
		Path string `json:"path" form:"path" validate:"required"`
	}

	// bind req input
	if err := c.Bind(&i); err != nil {
		return c.JSON(http.StatusBadRequest, response.Exception{Code: errcode.InvalidRequest, Error: err})
	}

	// validate
	if err := c.Validate(&i); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.Exception{Code: errcode.ValidationError, Error: err})
	}

	// clear up
	i.Path = strings.TrimSpace(i.Path)

	// create presigned url
	signedUrl, err := aws.SignedURL(i.Path)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Exception{
			Code:  errcode.ResourceNotFound,
			Error: err,
		})
	}
	return c.JSON(http.StatusOK, response.Item{Item: signedUrl})
}
