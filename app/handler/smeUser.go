package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"sme-api/app/entity"
	"sme-api/app/env"
	"sme-api/app/kit/general"
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

// GetSMEUserByToken :
func (h Handler) GetSMEUserByToken(c echo.Context) error {
	smeAdminData := c.Get("SME_ADMIN")
	if smeAdminData != nil {
		return c.JSON(http.StatusOK, response.Item{Item: c.Get("SME_ADMIN").(entity.SMEUser)})
	} else {
		return c.JSON(http.StatusOK, response.Item{Item: c.Get("SME_USER").(entity.SMEUser)})
	}
}

// GetSMEUsers :
func (h Handler) GetSMEUsers(c echo.Context) error {
	users, cursor, err := h.repository.FindSMEUsers(c.Request().Context(), repository.FindSMEUserFilter{
		Cursor:    c.QueryParam("cursor"),
		IDs:       c.Request().URL.Query()["Id"],
		Emails:    c.Request().URL.Query()["email"],
		FirstName: c.QueryParam("firstName"),
		LastName:  c.QueryParam("lastName"),
		CompanyID: c.QueryParam("companyId"),
		Status:    entity.UserStatus(c.QueryParam("status")),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, response.Items{Items: users, Cursor: cursor, Count: len(users)})
}

// UpdateSMEUser :
func (h Handler) UpdateSMEUser(c echo.Context) error {
	userId := c.QueryParam("id")
	var i struct {
		FirstName     string            `json:"firstName" form:"firstName" validate:"max=50"`
		LastName      string            `json:"lastName" form:"lastName" validate:"max=50"`
		Title         string            `json:"title" form:"title" validate:""`
		IC            string            `json:"ic" form:"ic" validate:"max=12"`
		Position      string            `json:"position" form:"position"`
		Email         string            `json:"email" form:"email" validate:"max=100"`
		Contact       string            `json:"contact" form:"contact" validate:""`
		MobileContact string            `json:"mobileContact" form:"mobileContact" validate:""`
		ApprovedBy    string            `json:"approvedBy" form:"approvedBy"`
		Password      string            `json:"password" form:"password" validate:"max=20"`
		Status        entity.UserStatus `json:"status" form:"status"`
	}

	// bind req input
	if err := c.Bind(&i); err != nil {
		return c.JSON(http.StatusBadRequest, response.Exception{Code: errcode.InvalidRequest, Error: err})
	}

	// validate
	if err := c.Validate(&i); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.Exception{Code: errcode.ValidationError, Error: err})
	}

	// check if SME user exists
	user, err := h.repository.FindSMEUserByID(c.Request().Context(), userId)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, response.Exception{Code: errcode.RecordNotFound, Error: fmt.Errorf("SME user %s not found", userId)})
		}
		return c.JSON(http.StatusInternalServerError, response.Exception{Code: errcode.ServerError, Error: err})
	}

	if i.FirstName != "" {
		user.FirstName = i.FirstName
	}

	if i.LastName != "" {
		user.LastName = i.LastName
	}

	if i.Email != "" {
		i.Email = strings.TrimSpace(i.Email)
		user.Email = strings.ToLower(i.Email)
	}

	if i.Password != "" {
		i.Password = strings.TrimSpace(i.Password)
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
		user.PasswordHash = passwordHash
		user.PasswordSalt = salt
	}

	if i.IC != "" {
		user.IC = strings.TrimSpace(i.IC)
	}

	if i.Position != "" {
		user.Position = i.Position
	}

	if i.Contact != "" {
		user.Contact = strings.TrimSpace(i.Contact)
	}

	if i.MobileContact != "" {
		user.MobileContact = strings.TrimSpace(i.MobileContact)
	}

	if i.Status != "" {
		// check if update status to active
		if i.Status == entity.UserStatusActive {
			// check total number of users registered
			totalUsers, _, err := h.repository.FindSMEUsers(c.Request().Context(), repository.FindSMEUserFilter{
				CompanyID: user.CompanyID.Hex(),
				Status:    entity.UserStatusActive,
			})
			if err == nil && len(totalUsers) > 3 {
				return c.JSON(http.StatusBadRequest, response.Exception{
					Code:  errcode.ExceedUserLimit,
					Error: fmt.Errorf("total number of users exceeded"),
				})
			}
		}
		user.Status = i.Status
	}

	if i.Title != "" {
		user.Title = entity.UserTitle(i.Title)
	}

	// check if profile picture is uploaded
	if profilePicture, _ := c.FormFile("profilePicture"); profilePicture != nil {
		// create path
		path := fmt.Sprintf("smeUser/%s/profile/", strings.TrimSpace(user.CompanyID.Hex()))

		// push to s3 bucket
		// url, httpStatus, exception := aws.PushDocBucket(path, profilePicture)
		// if exception != nil {
		// 	return c.JSON(httpStatus, exception)
		// }
		exception := general.ReadFile(profilePicture, path)
		if exception != nil {
			return c.JSON(http.StatusInternalServerError, exception)
		}
		user.ProfilePicture = path + profilePicture.Filename
	}

	// create time object
	timeNow := time.Now().UTC()
	user.Model.UpdatedAt = timeNow

	if _, err = h.repository.UpsertSMEUser(user); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	return c.JSON(http.StatusOK, nil)
}

// Create SME user :
func (h Handler) CreateSMEUser(c echo.Context) error {
	var i struct {
		CompanyID     string          `json:"companyId" form:"companyId" validate:"required,max=50"`
		FirstName     string          `json:"firstName" form:"firstName" validate:"required,max=50"`
		LastName      string          `json:"lastName" form:"lastName" validate:"required,max=50"`
		Title         string          `json:"title" form:"title" validate:"required"`
		IC            string          `json:"ic" form:"ic" validate:"max=12"`
		Position      string          `json:"position" form:"position" validate:"required"`
		Email         string          `json:"email" form:"email" validate:"required,email,min=10,max=100"`
		Contact       string          `json:"contact" form:"contact" validate:""`
		MobileContact string          `json:"mobileContact" form:"mobileContact" validate:""`
		ApprovedBy    string          `json:"approvedBy" form:"approvedBy"`
		Password      string          `json:"password" form:"password" validate:"required,min=6,max=20"`
		Role          entity.UserRole `json:"role" form:"role" validate:"required,role"`
		Status        string          `json:"status" form:"status" validate:"required,eq=ACTIVE|eq=INACTIVE"`
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
	i.CompanyID = strings.TrimSpace(i.CompanyID)
	i.Contact = strings.TrimSpace(i.Contact)
	i.MobileContact = strings.TrimSpace(i.MobileContact)
	i.ApprovedBy = strings.TrimSpace(i.ApprovedBy)
	maxUser := 3

	// validate
	if err := c.Validate(&i); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.Exception{Code: errcode.ValidationError, Error: err})
	}

	// check if SME company exists
	if sme, err := h.repository.FindSMEByID(c.Request().Context(), i.CompanyID); err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, response.Exception{Code: errcode.RecordNotFound, Error: fmt.Errorf("SME %s not found", i.CompanyID)})
		}
		return c.JSON(http.StatusInternalServerError, response.Exception{Code: errcode.ServerError, Error: err})
	} else if sme.Status != entity.StatusActive {
		return c.JSON(http.StatusBadRequest, response.Exception{Code: errcode.StatusInActive, Error: fmt.Errorf("SME %s status inactive", i.CompanyID)})
	}
	
	// check if user exists
	exists, _ := h.repository.FindSMEUserByEmail(c.Request().Context(), i.Email)
	if exists != nil {
		return c.JSON(http.StatusBadRequest, response.Exception{
			Code: errcode.EmailFound,
		})
	}

	fmt.Println(6)
	subscription, _, err := h.repository.FindSubscriptions(c.Request().Context(), repository.FindSubscriptionFilter{
		CorporateIDs: []string{i.CompanyID},
	})
	if err == nil {
		if len(subscription) > 0 {
			if subscription[0].SubscriptionPlan != "Single Business Plan" {
                        	maxUser = 10
                	}
		}
	}

	// check total number of users registered
	totalUsers, _, err := h.repository.FindSMEUsers(c.Request().Context(), repository.FindSMEUserFilter{
		CompanyID: i.CompanyID,
		Status:    entity.UserStatusActive,
	})
	if err == nil && len(totalUsers) > maxUser {
		return c.JSON(http.StatusBadRequest, response.Exception{
			Code:  errcode.ExceedUserLimit,
			Error: fmt.Errorf("total number of users exceeded"),
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

	// create user object
	userID := primitive.NewObjectID()
	companyOId, err := primitive.ObjectIDFromHex(i.CompanyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}
	user := entity.SMEUser{
		ID:            &userID,
		FirstName:     i.FirstName,
		LastName:      i.LastName,
		Title:         entity.UserTitle(i.Title),
		Email:         i.Email,
		IC:            i.IC,
		Contact:       i.Contact,
		MobileContact: i.MobileContact,
		CompanyID:     &companyOId,
		Position:      i.Position,
		PasswordHash:  passwordHash,
		PasswordSalt:  salt,
		Role:          i.Role,
		Status:        entity.UserStatus(i.Status),
		Model: entity.Model{
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		},
	}

	// check if profile picture is uploaded
	if profilePicture, _ := c.FormFile("profilePicture"); profilePicture != nil {
		// create path
		path := fmt.Sprintf("user/%s/profile/", strings.TrimSpace(i.CompanyID))

		// push to s3 bucket
		exception := general.ReadFile(profilePicture, path)
		if exception != nil {
			return c.JSON(http.StatusInternalServerError, exception)
		}
		user.ProfilePicture = path + profilePicture.Filename
	}

	if _, err = h.repository.CreateSMEUser(c.Request().Context(), user); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Exception{
			Code:  errcode.ServerError,
			Error: err,
		})
	}

	// SME user self register
	// if i.ApprovedBy == "" {
	// 	content :=
	// 		`<p><span style='font-size:16px;line-height:normal;font-family:'Calibri Light',sans-serif;color:#0C0F33;'>Thank you for signing up, {Title} {FirstName} {LastName},</span></p>
	// 		<p style='margin:0in;margin-bottom:0in;line-height:normal;font-size:15px;font-family:'Arial',sans-serif;background:white;'><br></p>
	// 		<p style='margin:0in;margin-bottom:0in;line-height:normal;font-size:15px;font-family:'Arial',sans-serif;background:white;'><span style='font-size:16px;font-family:'Calibri Light',sans-serif;color:#0C0F33;'>You&rsquo;re now part of a growing business community who is preparing their company for sustainable changes and the race to stop the planet&rsquo;s temperature from rising in 2030.</span></p>
	// 		<p style='margin:0in;margin-bottom:0in;line-height:normal;font-size:15px;font-family:'Arial',sans-serif;background:white;'><br></p>
	// 		<p style='margin:0in;margin-bottom:0in;line-height:normal;font-size:15px;font-family:'Arial',sans-serif;background:white;'><span style='font-size:16px;font-family:'Calibri Light',sans-serif;color:#0C0F33;'>We built CSI because we wanted to create a trustworthy and localised ESG platform to help you&nbsp;</span><span style='font-size:16px;font-family:'Calibri Light',sans-serif;color:#0C0F33;'>respond to your sustainability management challenges and prepare for a transition to the new economy.</span></p>
	// 		<p style='margin:0in;margin-bottom:0in;line-height:normal;font-size:15px;font-family:'Arial',sans-serif;background:white;'><br></p>
	// 		<p style='font-size:16px;font-family:'Calibri Light',sans-serif;color:#0C0F33;'>To help you get started, we&rsquo;re here to show you how to use CSI with these 3 simple tips:</p>
	// 		<ol>
	// 		    <li>
	// 		        <div>
	// 		            <p style='font-size:16px;font-family:'Calibri Light',sans-serif;color:#0C0F33;'>Take your ESG Assessment</p>
	// 		            <p style='font-size:16px;font-family:'Calibri Light',sans-serif;color:#0C0F33;'>Answer the 40 Yes-No questionnaire and upload supporting documents (e.g. ISO certifications).</p>
	// 		        </div>
	// 		    </li>
	// 		    <li>
	// 		        <div>
	// 		            <p style='font-size:16px;font-family:'Calibri Light',sans-serif;color:#0C0F33;'>View your ESG Report</p>
	// 		            <p style='font-size:16px;font-family:'Calibri Light',sans-serif;color:#0C0F33;'>Evaluate your ESG performance in comparison with peers.&nbsp;</p>
	// 		        </div>
	// 		    </li>
	// 		    <li>
	// 		        <div>
	// 		            <p style='font-size:16px;font-family:'Calibri Light',sans-serif;color:#0C0F33;'>Share your ESG Report</p>
	// 		            <p style='font-size:16px;font-family:'Calibri Light',sans-serif;color:#0C0F33;'>Decide if you want to share your ESG results and collaborate with your preferred supply chain owners and corporates.</p>
	// 		        </div>
	// 		    </li>
	// 		</ol>
	// 		<p style='margin:0in;margin-bottom:0in;line-height:normal;font-size:15px;font-family:'Arial',sans-serif;background:white;'><br></p>
	// 		<p style='margin:0in;margin-bottom:0in;line-height:normal;font-size:15px;font-family:'Arial',sans-serif;background:white;'><span style='font-size:16px;line-height:115%;font-family:'Calibri Light',sans-serif;color:#0C0F33;'><a href='http://ec2-13-214-140-250.ap-southeast-1.compute.amazonaws.com/sme'>Embark on your ESG discovery journey now!</a></span></p>
	// 		<p style='margin:0in;margin-bottom:0in;line-height:normal;font-size:15px;font-family:'Arial',sans-serif;background:white;'><br></p>
	// 		<p style='margin:0in;margin-bottom:0in;line-height:normal;font-size:15px;font-family:'Arial',sans-serif;background:white;'><span style='font-size:16px;font-family:'Calibri Light',sans-serif;color:#0C0F33;'>Please reply to this email if you still have questions. We hope you have a productive experience with us!</span></p>
	// 		<p style='margin:0in;margin-bottom:0in;line-height:normal;font-size:15px;font-family:'Arial',sans-serif;background:white;'><br></p>
	// 		<p style='margin:0in;margin-bottom:0in;line-height:115%;font-size:15px;font-family:'Arial',sans-serif;background:white;'><span style='font-size:16px;line-height:115%;font-family:'Calibri Light',sans-serif;color:#0C0F33;'>Best regards,</span></p>
	// 		<p style='margin:0in;margin-bottom:0in;line-height:115%;font-size:15px;font-family:'Arial',sans-serif;'><span style='font-size:16px;line-height:115%;font-family:'Calibri Light',sans-serif;color:#0C0F33;'>SDM Team</span></p>
	// 		<p style='margin:0in;margin-bottom:0in;line-height:normal;font-size:15px;font-family:'Arial',sans-serif;background:white;'><br></p>`
	// 	content = strings.Replace(content, "{Title}", i.Title, 1)
	// 	content = strings.Replace(content, "{FirstName}", i.FirstName, 1)
	// 	content = strings.Replace(content, "{LastName}", i.LastName, 1)

	// 	recipient := i.Email
	// 	if httpStatus, exception := aws.SendEmails([]*string{&recipient}, "Welcome to CSI!", content, "", env.Config.AWS.Sender2); exception != nil {
	// 		return c.JSON(httpStatus, exception)
	// 	}
	// }

	return c.JSON(http.StatusOK, nil)
}

// SME user login :
func (h Handler) SMEUserLogin(c echo.Context) error {
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

	// check if user exists
	user, err := h.repository.FindSMEUserByEmail(c.Request().Context(), i.Email)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Exception{
			Code:  errcode.UserNotFound,
			Error: err,
		})
	}
	// check user status
	if user.Status != entity.UserStatusActive {
		return c.JSON(http.StatusBadRequest, response.Exception{Code: errcode.StatusInActive, Error: fmt.Errorf("SME user %s status inactive", i.Email)})
	}

	// check if SME company exists
	if sme, err := h.repository.FindSMEByID(c.Request().Context(), user.CompanyID.Hex()); err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, response.Exception{Code: errcode.RecordNotFound, Error: fmt.Errorf("SME %s not found", user.CompanyID.Hex())})
		}
		return c.JSON(http.StatusInternalServerError, response.Exception{Code: errcode.ServerError, Error: err})
	} else if sme.Status != entity.StatusActive {
		return c.JSON(http.StatusBadRequest, response.Exception{Code: errcode.StatusInActive, Error: fmt.Errorf("SME %s status inactive", user.CompanyID.Hex())})
	}

	// compare password & return error if password is not same
	salt, pepper := user.PasswordSalt, env.Config.Jwt.Secret
	if isSame := password.Compare(i.Password, salt, pepper, user.PasswordHash); !isSame {
		return c.JSON(http.StatusBadRequest, response.Exception{
			Code: errcode.AuthenticationError,
		})
	}

	// convert to string array
	scopes := []string{}
	s := fmt.Sprintf("%v", user.Role)
	scopes = append(scopes, s)

	// generate token pair
	tokens, err := jwt.GenerateTokens(env.Config.Jwt.Secret, map[string]string{
		"sub":    user.ID.Hex(),
		"aud":    user.ID.Hex(),
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

// ValidateSMEUser :
func ValidateSMEUser(h Handler, ctx context.Context, ID string) (*entity.SMEUser, int, *response.Exception) {
	// check if SME user exists
	user, err := h.repository.FindSMEUserByID(ctx, ID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &entity.SMEUser{}, http.StatusNotFound, &response.Exception{Code: errcode.RecordNotFound, Error: fmt.Errorf("SME user %s not found", ID)}
		}
		return &entity.SMEUser{}, http.StatusInternalServerError, &response.Exception{Code: errcode.ServerError, Error: err}
	}

	if user.Status != entity.UserStatusActive {
		return &entity.SMEUser{}, http.StatusBadRequest, &response.Exception{Code: errcode.StatusInActive, Error: fmt.Errorf("SME user %s status inactive", ID)}
	}
	return user, http.StatusOK, nil
}
