package jwt

import (
	"sme-api/app/env"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const (
	accessTokenExpiresInHour  = 24 * 30 * 3 // every 3 months
	refreshTokenExpiresInHour = 24 * 30 * 6 // every 6 months
	tokenExpiresInSecond      = 86400       // 1 day
)

// ExtractedClaims :
type ExtractedClaims struct {
	Audience string `json:"aud"`
	Issuer   string `json:"iss"`
	Subject  string `json:"sub"`
	Exp      string `json:"exp"`    // exp extracted is a string, to be casted to time.Time
	Scopes   string `json:"scopes"` // scopes extracted is a string, delimit by ,
}

// GenerateAccessToken :
func GenerateAccessToken(secretKey string, extraClaims map[string]string) (map[string]interface{}, error) {
	// init access token and claims
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	exp := time.Now().Add(time.Hour * accessTokenExpiresInHour).Format(time.RFC3339)

	// set issuer, expiry time
	claims["iss"] = env.Config.Jwt.Issuer
	claims["exp"] = exp

	// set extra claims (admin)
	if extraClaims != nil {
		for k, v := range extraClaims {
			claims[k] = v
		}
	}

	// generate token
	accessToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"accessToken": accessToken,
		"expiresAt":   exp,
		"expiresIn":   accessTokenExpiresInHour * 3600,
	}, nil
}

// GenerateRefreshToken :
func GenerateRefreshToken(secretKey string, extraClaims map[string]string) (string, error) {
	// init access token and claims
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	// set issuer, expiry time
	claims["iss"] = env.Config.Jwt.Issuer
	claims["exp"] = time.Now().Add(time.Hour * refreshTokenExpiresInHour).Format(time.RFC3339)

	// use aud as user identifier
	claims["aud"] = extraClaims["aud"]

	// generate token
	refreshToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

// GenerateTokens : Generate both access and refresh token
func GenerateTokens(secretKey string, extraClaims map[string]string) (map[string]interface{}, error) {
	accessToken, err := GenerateAccessToken(secretKey, extraClaims)
	if err != nil {
		return nil, err
	}

	refreshToken, err := GenerateRefreshToken(secretKey, extraClaims)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"accessToken":  accessToken["accessToken"],
		"refreshToken": refreshToken,
		"expiresAt":    accessToken["expiresAt"],
		"expiresIn":    accessToken["expiresIn"],
	}, nil
}
