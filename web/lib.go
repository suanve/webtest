package web

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
	"webtest/config"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func RandInt64(min, max int) int {
	rand.Seed(time.Now().Unix())
	if min >= max || min == 0 || max == 0 {
		return max
	}
	return rand.Intn(max-min) + min
}

// CreateToken create token
func CreateToken(claims *Claims) (signedToken string, success bool) {
	claims.ExpiresAt = time.Now().Add(time.Minute * 30).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(config.SecretKey))
	if err != nil {
		return
	}
	success = true
	return
}

func ValidateToken(signedToken string) (claims *Claims, success bool) {
	token, err := jwt.ParseWithClaims(signedToken, &Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected login method %v", token.Header["alg"])
			}
			return []byte(config.SecretKey), nil
		})

	if err != nil {
		return
	}

	claims, ok := token.Claims.(*Claims)
	if ok && token.Valid {
		success = true
		return
	}

	return
}

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("token")
		if token == "" {
			c.JSON(http.StatusOK, gin.H{
				"status": -1,
				"msg":    "请求未携带token，无权限访问",
			})
			c.Abort()
			return
		}

		_, err := ValidateToken(token)
		if !err {
			c.JSON(http.StatusOK, gin.H{
				"status": -1,
				"msg":    "token faild",
			})
			c.Abort()
			return
		}
	}
}
