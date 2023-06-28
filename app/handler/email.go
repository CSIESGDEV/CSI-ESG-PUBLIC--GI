package handler

import (
	"net/http"
	"sme-api/app/env"
	kit "sme-api/app/kit/aws"
	"sme-api/app/response"
	"sme-api/app/response/errcode"
	"strings"

	"github.com/labstack/echo/v4"
)

// SendContactUsEmail :
func (h Handler) SendContactUsEmail(c echo.Context) error {
	var i struct {
		Title   string `json:"title" form:"title" validate:"required,min=10,max=150"`
		Content string `json:"content" form:"content" validate:"required,max=300"`
		Email   string `json:"email" form:"email" validate:"required,email,min=10,max=100"`
	}

	// bind req input
	if err := c.Bind(&i); err != nil {
		return c.JSON(http.StatusBadRequest, response.Exception{Code: errcode.InvalidRequest, Error: err})
	}

	// cleanup
	i.Email = strings.TrimSpace(i.Email)

	// validate
	if err := c.Validate(&i); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.Exception{Code: errcode.ValidationError, Error: err})
	}

	content := "<p>Name: {Name}</p><p>Email: {Email}</p><p>Message: {Message}</p>"
	content = strings.Replace(content, "{Name}", i.Title, 1)
	content = strings.Replace(content, "{Email}", i.Email, 1)
	content = strings.Replace(content, "{Message}", i.Content, 1)

	recipient := "enquiry@sdmsb.com"
	if httpStatus, exception := kit.SendEmails([]*string{&recipient}, i.Title, content, "", env.Config.AWS.Sender); exception != nil {
		return c.JSON(httpStatus, exception)
	}

	return c.JSON(http.StatusOK, nil)
}
