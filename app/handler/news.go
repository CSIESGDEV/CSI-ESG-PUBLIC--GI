package handler

import (
	"context"
	"fmt"
	"net/http"
	"sme-api/app/entity"
	"sme-api/app/kit/aws"
	"sme-api/app/repository"
	"sme-api/app/response"
	"sme-api/app/response/errcode"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetNewss :
func (h Handler) GetNews(c echo.Context) error {
	news, cursor, err := h.repository.FindNews(c.Request().Context(), repository.FindNewsFilter{
		Cursor: c.QueryParam("cursor"),
		IDs:    c.Request().URL.Query()["id"],
		Link:   c.QueryParam("link"),
		Status: entity.Status(c.QueryParam("status")),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, response.Items{Items: news, Cursor: cursor, Count: len(news)})
}

// UpdateNews :
func (h Handler) UpdateNews(c echo.Context) error {
	newsId := c.QueryParam("id")
	var i struct {
		Title  string        `json:"title" form:"title"`
		Link   string        `json:"link" form:"link"`
		Status entity.Status `json:"status" form:"status"`
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

	// check if News exists
	news, err := h.repository.FindNewsByID(c.Request().Context(), newsId)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, response.Exception{Code: errcode.RecordNotFound, Error: fmt.Errorf("news %s not found", newsId)})
		}
		return c.JSON(http.StatusInternalServerError, response.Exception{Code: errcode.ServerError, Error: err})
	}

	if i.Link != "" {
		news.Link = i.Link
	}

	if i.Title != "" {
		news.Title = i.Title
	}

	if i.Status != "" {
		news.Status = i.Status
	}

	// check if image is uploaded
	if image, _ := c.FormFile("image"); image != nil {
		// create path
		path := fmt.Sprintf("news/%s/", strings.TrimSpace(newsId))

		// push to s3 bucket
		url, httpStatus, exception := aws.PushDocBucket(path, image)
		if exception != nil {
			return c.JSON(httpStatus, exception)
		}
		news.Image = url
	}

	news.Model.UpdatedAt = time.Now().UTC()

	if _, err := h.repository.UpsertNews(news); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, response.Item{Item: news})
}

// DeleteNews :
func (h Handler) DeleteNews(c echo.Context) error {
	newsId := c.QueryParam("id")

	news, err := h.repository.FindNewsByID(c.Request().Context(), newsId)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, response.Exception{Code: errcode.RecordNotFound, Error: fmt.Errorf("news %s not found", newsId)})
		}
		return c.JSON(http.StatusInternalServerError, response.Exception{Code: errcode.ServerError, Error: err})
	}

	if _, err := h.repository.DeleteNewsByID(news); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{Code: errcode.ServerError, Error: err})
	}

	return c.JSON(http.StatusOK, nil)
}

// CreateNews :
func (h Handler) CreateNews(c echo.Context) error {
	var i struct {
		Title string `json:"title" form:"title" validate:"required"`
		Link  string `json:"link" form:"link" validate:"required"`
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

	// create News object
	newsID := primitive.NewObjectID()

	news := entity.News{
		ID:     &newsID,
		Title:  i.Title,
		Link:   i.Link,
		Status: entity.StatusActive,
		Model: entity.Model{
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		},
	}

	// check if image is uploaded
	if image, _ := c.FormFile("image"); image != nil {
		// create path
		path := fmt.Sprintf("news/%s/", strings.TrimSpace(newsID.Hex()))

		// push to s3 bucket
		url, httpStatus, exception := aws.PushDocBucket(path, image)
		if exception != nil {
			return c.JSON(httpStatus, exception)
		}
		news.Image = url
	}

	if _, err := h.repository.CreateNews(c.Request().Context(), news); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, response.Item{Item: news})
}

// ValidateNews :
func ValidateNews(h Handler, ctx context.Context, id string) (*entity.News, int, *response.Exception) {
	// check if News exists
	News, err := h.repository.FindNewsByID(ctx, id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &entity.News{}, http.StatusNotFound, &response.Exception{Code: errcode.RecordNotFound, Error: fmt.Errorf("news %s not found", id)}
		}
		return &entity.News{}, http.StatusInternalServerError, &response.Exception{Code: errcode.ServerError, Error: err}
	}

	if News.Status != entity.StatusActive {
		return &entity.News{}, http.StatusBadRequest, &response.Exception{Code: errcode.StatusInActive, Error: fmt.Errorf("news %s status inactive", id)}
	}
	return News, http.StatusOK, nil
}
