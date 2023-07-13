package jwt

import (
	"csi-api/app/env"
	"encoding/json"
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Validate :
func Validate(tokenString string) (*jwt.MapClaims, error) {

	// verify token string and return a hmac secret
	token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Unexpected signing method: ")
		}
		return []byte(env.Config.Jwt.Secret), nil
	})

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		return &claims, nil
	}

	return nil, errors.New("Unable to validate token")
}

// ExtractClaims :
func ExtractClaims(claims *jwt.MapClaims) (*ExtractedClaims, error) {
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return nil, errors.New("Fail to marshal claims. Please contact admin for assistance")
	}

	reqClaims := ExtractedClaims{}
	if err := json.Unmarshal(claimsJSON, &reqClaims); err != nil {
		return nil, errors.New("Fail to unmarshal claims. Please contact admin for assistance")
	}
	return &reqClaims, nil
}

// IsTokenExpired : check if token is expired
func IsTokenExpired(timeString string) (bool, error) {
	rfcLayout := "2006-01-02T15:04:05Z07:00"
	expTime, err := time.Parse(rfcLayout, timeString)
	if err != nil {
		return true, errors.New("Fail to parse expiry time string")
	}
	if expTime.Unix() <= time.Now().Unix() {
		return true, nil
	}
	return false, nil
}
