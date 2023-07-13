package handler

import (
	"csi-api/app/entity"
	"csi-api/app/env"
	"csi-api/app/kit/jwt"
	"csi-api/app/response"
	"csi-api/app/response/errcode"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// AuthenticatedUser:
func (h Handler) AuthenticatedUser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// default role
		role := "public"
		// get access token
		tokenArr := strings.Split(strings.TrimSpace(c.Request().Header.Get("Authorization")), " ")
		// if token is found then check user role
		if len(tokenArr) > 1 {
			_, reqClaims, httpStatus, exception := VerifyToken(tokenArr)
			if exception != nil {
				return c.JSON(httpStatus, exception)
			}
			admin, _, _ := ValidateAdmin(h, c.Request().Context(), reqClaims.Audience)

			smeUser, _, _ := ValidateSMEUser(h, c.Request().Context(), reqClaims.Audience)

			// corporateUser, _, _ := ValidateCorporateUser(h, c.Request().Context(), reqClaims.Audience)

			if admin.ID != nil {
				if admin.Role == entity.UserRoleAdmin {
					role = "superAdmin"
				} else if admin.Role == entity.UserRoleSMEVendor {
					role = "smeVendor"
				} else if admin.Role == entity.UserRoleCorporateVendor {
					role = "corporateVendor"
				} else {
					role = "admin"
				}
				c.Set("ADMIN", *admin)
			} else if smeUser.ID != nil {
				role = "smeAdmin"
				c.Set("SME_ADMIN", *smeUser)
				// if smeUser.Role == entity.UserRoleAdmin {
				// 	role = "smeAdmin"
				// 	c.Set("SME_ADMIN", *smeUser)
				// } else {
				// 	role = "smeUser"
				// 	c.Set("SME_USER", *smeUser)
				// }
			}
			// else if corporateUser.ID != nil {
			// 	if corporateUser.Role == entity.UserRoleAdmin {
			// 		role = "corporateAdmin"
			// 		c.Set("CORPORATE_ADMIN", *corporateUser)
			// 	} else {
			// 		role = "corporateUser"
			// 		c.Set("CORPORATE_USER", *corporateUser)
			// 	}
			// }
		}
		c.Set("role", role)
		return next(c)
	}
}

// VerifyToken :
func VerifyToken(tokenArr []string) (string, *jwt.ExtractedClaims, int, *response.Exception) {
	if len(tokenArr) != 2 {
		return "", nil, http.StatusUnauthorized, &response.Exception{Code: errcode.InvalidToken}
	}

	switch strings.TrimSpace(tokenArr[0]) {
	case "Bearer":
	default:
		return "", nil, http.StatusUnauthorized, &response.Exception{Code: errcode.InvalidToken}
	}

	token := strings.TrimSpace(tokenArr[1])
	claims, err := jwt.Validate(token)

	if err != nil {
		return "", nil, http.StatusUnauthorized, &response.Exception{Code: errcode.InvalidToken, Error: err}
	}

	if ok := claims.VerifyIssuer(env.Config.Jwt.Issuer, true); !ok {
		return "", nil, http.StatusUnauthorized, &response.Exception{Code: errcode.InvalidToken}
	}

	reqClaims, err := jwt.ExtractClaims(claims)
	if err != nil {
		return "", nil, http.StatusUnauthorized, &response.Exception{Code: errcode.InvalidToken, Error: err}
	}

	isExpired, err := jwt.IsTokenExpired(reqClaims.Exp)
	if err != nil {
		return "", nil, http.StatusUnauthorized, &response.Exception{Code: errcode.InvalidToken, Error: err}
	}

	if isExpired {
		return "", nil, http.StatusUnauthorized, &response.Exception{Code: errcode.InvalidToken, Error: err}
	}

	return token, reqClaims, http.StatusOK, nil
}
