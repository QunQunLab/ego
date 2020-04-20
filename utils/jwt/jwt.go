// Package jwt provides JSON web token related functions.
package jwt

import (
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

var secretKey []byte

func Init(key string) {
	secretKey = []byte(key)
	//secretKey = []byte("18ddf30d665538d3ab90b8e0bf6c96879be4fa6d")
}

// JWTInfo represents JSON web token payload info.
type JWTInfo struct {
	UID     int
	IP      string
	Roles   string
	Expires int64
}

// NewTokenHMAC generates a HMAC token in HS256 signing method with secretKey
func NewTokenHMAC(jwtInfo JWTInfo, exp int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"UID":   jwtInfo.UID,
		"IP":    jwtInfo.IP,
		"Roles": jwtInfo.Roles,
		"Exp":   time.Now().Unix() + exp,
	})
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// TokenValidate parse and validate token, returns JWT payload info.
func TokenValidate(tokenString string) (*JWTInfo, error) {
	token, err := jwt.Parse(tokenString,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v",
					token.Header["alg"])
			}
			// returns secretKey for the anonymous function(keyFunc)
			return secretKey, nil
		})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		info := JWTInfo{
			UID:     int(claims["UID"].(float64)),
			IP:      claims["IP"].(string),
			Roles:   claims["Roles"].(string),
			Expires: int64(claims["Exp"].(float64)),
		}
		return &info, nil
	}
	return nil, fmt.Errorf("token:%s valid failed", tokenString)
}
