package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"sme-api/app/entity"
	"sme-api/app/env"
	"sme-api/app/kit/jwt"
	"sme-api/app/kit/password"
	"sme-api/app/repository"
	"sme-api/app/response"
	"sme-api/app/response/errcode"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/random"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetAdminByToken :
func (h Handler) GetAdminByToken(c echo.Context) error {
	adminData := c.Get("ADMIN")
	admin := adminData.(entity.Admin)

	return c.JSON(http.StatusOK, response.Item{Item: admin})
}

// GetAdmins :
func (h Handler) GetAdmins(c echo.Context) error {
	admins, cursor, err := h.repository.FindAdmins(c.Request().Context(), repository.FindAdminFilter{
		Cursor: c.QueryParam("cursor"),
		IDs:    c.Request().URL.Query()["id"],
		Roles:  c.Request().URL.Query()["role"],
		Status: entity.UserStatus(c.QueryParam("status")),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, response.Items{Items: admins, Cursor: cursor, Count: len(admins)})
}

// UpdateAdmin :
func (h Handler) UpdateAdmin(c echo.Context) error {
	adminId := c.QueryParam("id")
	var i struct {
		FirstName   string `json:"firstName" form:"firstName" validate:"max=50"`
		LastName    string `json:"lastName" form:"lastName" validate:"max=50"`
		IC          string `json:"ic" form:"ic" validate:"max=12"`
		Email       string `json:"email" form:"email" validate:"max=100"`
		Contact     string `json:"contact" form:"contact" validate:"max=12"`
		CompanyName string `json:"companyName" form:"companyName" validate:"max=50"`
		Position    string `json:"position" form:"position" validate:"max=50"`
		ApprovedBy  string `json:"approvedBy" form:"approvedBy" validate:"max=50"`
		Password    string `json:"password" form:"password" validate:"max=20"`
		Role        string `json:"role" form:"role" validate:"max=20"`
		Status      string `json:"status" form:"status" validate:"max=10"`
	}

	// bind req input
	if err := c.Bind(&i); err != nil {
		return c.JSON(http.StatusBadRequest, response.Exception{Code: errcode.InvalidRequest, Error: err})
	}

	// validate
	if err := c.Validate(&i); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.Exception{Code: errcode.ValidationError, Error: err})
	}

	// cleanup
	i.Email = strings.TrimSpace(i.Email)
	i.Email = strings.ToLower(i.Email)
	i.Password = strings.TrimSpace(i.Password)
	i.IC = strings.TrimSpace(i.IC)
	i.Contact = strings.TrimSpace(i.Contact)

	admin, err := h.repository.FindAdminByID(c.Request().Context(), adminId)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, response.Exception{Code: errcode.RecordNotFound, Error: fmt.Errorf("admin %s not found", adminId)})
		}
		return c.JSON(http.StatusInternalServerError, &response.Exception{Code: errcode.ServerError, Error: err})
	}

	if i.FirstName != "" {
		admin.FirstName = i.FirstName
	}

	if i.LastName != "" {
		admin.LastName = i.LastName
	}

	if i.IC != "" {
		admin.IC = i.IC
	}

	if i.Email != "" {
		_, err := h.repository.FindAdminByEmail(c.Request().Context(), i.Email)
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusBadRequest, response.Exception{Code: errcode.RecordNotFound, Error: fmt.Errorf("admin %s not found", adminId)})
		}
		admin.Email = i.Email
	}

	if i.Contact != "" {
		admin.Contact = strings.TrimSpace(i.Contact)
	}

	if i.CompanyName != "" {
		admin.CompanyName = i.CompanyName
	}

	if i.Position != "" {
		admin.Position = i.Position
	}

	if i.Password != "" {
		// create password hash and salt
		salt := random.String(10)
		pepper := env.Config.Jwt.Secret
		passwordHash, err := password.Create(i.Password, salt, pepper)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, response.Exception{
				Code:  errcode.ServerError,
				Error: err,
			})
		}
		admin.PasswordHash = passwordHash
		admin.PasswordSalt = salt
	}

	if i.Role != "" {
		admin.Role = entity.UserRole(i.Role)
	}

	if i.Status != "" {
		admin.Status = entity.UserStatus(i.Status)
	}

	if i.ApprovedBy != "" {
		// check if admin exists
		approvedBy, httpStatus, exception := ValidateAdmin(h, c.Request().Context(), i.ApprovedBy)
		if exception != nil {
			return c.JSON(httpStatus, exception)
		}
		admin.ApprovedBy = approvedBy.ID
	}

	admin.Model.UpdatedAt = time.Now().UTC()

	if _, err = h.repository.UpsertAdmin(admin); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, nil)
}

// CreateAdmin :
func (h Handler) CreateAdmin(c echo.Context) error {
	var i struct {
		FirstName   string `json:"firstName" form:"firstName" validate:"required,max=50"`
		LastName    string `json:"lastName" form:"lastName" validate:"required,max=50"`
		IC          string `json:"ic" form:"ic" validate:"max=12"`
		Email       string `json:"email" form:"email" validate:"required,email,min=10,max=100"`
		Contact     string `json:"contact" form:"contact" validate:"required,min=10,max=12"`
		CompanyName string `json:"companyName" form:"companyName" validate:"required,max=100"`
		Position    string `json:"position" form:"position" validate:"required,max=50"`
		ApprovedBy  string `json:"approvedBy" form:"approvedBy" validate:"required,max=50"`
		Password    string `json:"password" form:"password" validate:"required,min=6,max=20"`
		Role        string `json:"role" form:"role" validate:"required,eq=USER|eq=SME_VENDOR|eq=CORPORATE_VENDOR"`
	}

	// bind req input
	if err := c.Bind(&i); err != nil {
		return c.JSON(http.StatusBadRequest, response.Exception{Code: errcode.InvalidRequest, Error: err})
	}

	// cleanup
	i.Email = strings.TrimSpace(i.Email)
	i.Email = strings.ToLower(i.Email)
	i.Password = strings.TrimSpace(i.Password)
	i.IC = strings.TrimSpace(i.IC)
	i.Contact = strings.TrimSpace(i.Contact)

	// validate
	if err := c.Validate(&i); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.Exception{Code: errcode.ValidationError, Error: err})
	}
	// check if admin exists
	admins, _, err := h.repository.FindAdmins(c.Request().Context(), repository.FindAdminFilter{
		Emails: []string{i.Email},
	})
	if err == nil && len(admins) > 0 {
		return c.JSON(http.StatusBadRequest, response.Exception{
			Code:  errcode.EmailFound,
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
	// create password hash and salt
	salt := random.String(10)
	pepper := env.Config.Jwt.Secret
	passwordHash, err := password.Create(i.Password, salt, pepper)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	// create admin object
	adminId := primitive.NewObjectID()
	approvedBy, err := primitive.ObjectIDFromHex(i.ApprovedBy)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}
	admin := entity.Admin{
		ID:           &adminId,
		FirstName:    i.FirstName,
		LastName:     i.LastName,
		Email:        i.Email,
		IC:           i.IC,
		Contact:      i.Contact,
		CompanyName:  i.CompanyName,
		Position:     i.Position,
		PasswordHash: passwordHash,
		PasswordSalt: salt,
		Role:         entity.UserRole(i.Role),
		Status:       entity.UserStatusActive,
		ApprovedBy:   &approvedBy,
		Model: entity.Model{
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		},
	}

	if _, err = h.repository.CreateAdmin(c.Request().Context(), admin); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, nil)
}

// Admin login :
func (h Handler) AdminLogin(c echo.Context) error {
	var i struct {
		Email    string `json:"email" form:"email" validate:"required,email,min=10,max=100"`
		Password string `json:"password" form:"password" validate:"required,min=6,max=20"`
	}

	// bind req input
	if err := c.Bind(&i); err != nil {
		return c.JSON(http.StatusBadRequest, response.Exception{Code: errcode.InvalidRequest, Error: err})
	}

	// cleanup
	i.Email = strings.TrimSpace(i.Email)
	i.Email = strings.ToLower(i.Email)
	i.Password = strings.TrimSpace(i.Password)

	// validate
	if err := c.Validate(&i); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.Exception{Code: errcode.ValidationError, Error: err})
	}

	// check if admin exists
	admin, err := h.repository.FindAdminByEmail(c.Request().Context(), i.Email)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Exception{
			Code:  errcode.AdminNotFound,
			Error: err,
		})
	}

	if admin.Status != entity.UserStatusActive {
		return c.JSON(http.StatusBadRequest, response.Exception{Code: errcode.StatusInActive, Error: fmt.Errorf("admin %s status inactive", i.Email)})
	}
	// compare password & return error if password is not same
	salt, pepper := admin.PasswordSalt, env.Config.Jwt.Secret
	if isSame := password.Compare(i.Password, salt, pepper, admin.PasswordHash); !isSame {
		return c.JSON(http.StatusBadRequest, response.Exception{
			Code: errcode.AuthenticationError,
		})
	}

	// convert to string array
	scopes := []string{}
	s := fmt.Sprintf("%v", admin.Role)
	scopes = append(scopes, s)

	// generate token pair
	tokens, err := jwt.GenerateTokens(env.Config.Jwt.Secret, map[string]string{
		"sub":    admin.ID.Hex(),
		"aud":    admin.ID.Hex(),
		"scopes": strings.Join(scopes, ","),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, response.Item{Item: tokens})
}

// VendorLogin :
func (h Handler) VendorLogin(c echo.Context) error {
	var i struct {
		Email    string `json:"email" form:"email" validate:"required,email,min=10,max=100"`
		Password string `json:"password" form:"password" validate:"required,min=6,max=20"`
		Role     string `json:"role" form:"role" validate:"required,eq=SME_VENDOR|eq=CORPORATE_VENDOR"`
	}

	// bind req input
	if err := c.Bind(&i); err != nil {
		return c.JSON(http.StatusBadRequest, response.Exception{Code: errcode.InvalidRequest, Error: err})
	}

	// cleanup
	i.Email = strings.TrimSpace(i.Email)
	i.Email = strings.ToLower(i.Email)
	i.Password = strings.TrimSpace(i.Password)

	// validate
	if err := c.Validate(&i); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.Exception{Code: errcode.ValidationError, Error: err})
	}

	// check if admin exists
	admin, err := h.repository.FindAdminByEmail(c.Request().Context(), i.Email)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Exception{
			Code:  errcode.AdminNotFound,
			Error: err,
		})
	}

	if admin.Status != entity.UserStatusActive {
		return c.JSON(http.StatusBadRequest, response.Exception{Code: errcode.StatusInActive, Error: fmt.Errorf("admin %s status inactive", i.Email)})
	}
	// compare password & return error if password is not same
	salt, pepper := admin.PasswordSalt, env.Config.Jwt.Secret
	if isSame := password.Compare(i.Password, salt, pepper, admin.PasswordHash); !isSame {
		return c.JSON(http.StatusBadRequest, response.Exception{
			Code: errcode.AuthenticationError,
		})
	}
	if i.Role == string(entity.UserRoleSMEVendor) {
		if admin.Role != entity.UserRoleSMEVendor {
			return c.JSON(http.StatusBadRequest, response.Exception{
				Code:  errcode.InvalidAccess,
				Error: fmt.Errorf("sme vendor %s is not found", i.Email),
			})
		}
	} else if i.Role == string(entity.UserRoleCorporateVendor) {
		if admin.Role != entity.UserRoleCorporateVendor {
			return c.JSON(http.StatusBadRequest, response.Exception{
				Code:  errcode.InvalidAccess,
				Error: fmt.Errorf("corporate vendor %s is not found", i.Email),
			})
		}
	} else {
		return c.JSON(http.StatusBadRequest, response.Exception{
			Code: errcode.InvalidAccess,
		})
	}

	// convert to string array
	scopes := []string{}
	s := fmt.Sprintf("%v", admin.Role)
	scopes = append(scopes, s)

	// generate token pair
	tokens, err := jwt.GenerateTokens(env.Config.Jwt.Secret, map[string]string{
		"sub":    admin.ID.Hex(),
		"aud":    admin.ID.Hex(),
		"scopes": strings.Join(scopes, ","),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, response.Item{Item: tokens})
}

// ValidateAdmin :
func ValidateAdmin(h Handler, ctx context.Context, ID string) (*entity.Admin, int, *response.Exception) {
	// check if SME user exists
	admin, err := h.repository.FindAdminByID(ctx, ID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &entity.Admin{}, http.StatusNotFound, &response.Exception{Code: errcode.RecordNotFound, Error: fmt.Errorf("admin %s not found", ID)}
		}
		return &entity.Admin{}, http.StatusInternalServerError, &response.Exception{Code: errcode.ServerError, Error: err}
	}

	if admin.Status != entity.UserStatusActive {
		return &entity.Admin{}, http.StatusBadRequest, &response.Exception{Code: errcode.StatusInActive, Error: fmt.Errorf("admin %s status inactive", ID)}
	}
	return admin, http.StatusOK, nil
}
