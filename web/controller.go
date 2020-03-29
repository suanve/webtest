package web

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"webtest/config"

	//
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Set("content-type", "application/json")
		}
		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}

func init() {
	db, _ = sql.Open("mysql", config.DB)

	// mysql image starts need time.
	for {
		err := db.Ping()
		if err == nil {
			break
		}
		fmt.Println(err)
		time.Sleep(2 * time.Second)
	}
	// https://github.com/go-sql-driver/mysql/issues/674
	db.SetMaxIdleConns(0)
}

func Index(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "It's work!"})
}

//获取用户的输入并格式化为json
func Login(c *gin.Context) {
	data, _ := ioutil.ReadAll(c.Request.Body)
	var user User

	if err := json.Unmarshal(data, &user); err == nil {
		fmt.Println(user.Username)
		fmt.Println(user.Password)
	}
	if user.Username == "" || user.Password == "" {
		c.JSON(http.StatusOK, gin.H{"code": 401})
	}

	password, id := getPassword(user.Username)
	if password == user.Password {

		var claims Claims
		claims.Username = user.Username
		claims.Id = id

		t, _ := CreateToken(&claims)
		c.JSON(http.StatusOK, gin.H{"code": 200, "token": t})
	} else {
		c.JSON(http.StatusOK, gin.H{"code": 401})
	}
}

func API_GetChallenge(c *gin.Context) {
	Challenges := getChallenges()
	c.JSON(http.StatusOK, gin.H{"message": "success", "length": len(Challenges), "data": Challenges})
}

//根据token 获取用户身份，查找其启动的任务
func API_GetChallengeStatus(c *gin.Context) {
	//获取token
	token := c.Request.Header.Get("token")
	//获取用户名
	userInfo, err := ValidateToken(token)
	if !err {
		c.JSON(http.StatusOK, gin.H{
			"status": -1,
			"msg":    "token faild",
		})
		c.Abort()
		return
	}
	fmt.Println(userInfo)
	Challenges := getChallengesStatus(userInfo.Id)

	c.JSON(http.StatusOK, gin.H{"message": "success!", "length": len(Challenges), "data": Challenges})
}

//获取用户请求的实验id，启动对应的容器
func API_startChallenge(c *gin.Context) {
	//获取token
	token := c.Request.Header.Get("token")
	//获取用户名
	userInfo, err := ValidateToken(token)
	if !err {
		c.JSON(http.StatusOK, gin.H{
			"status": -1,
			"msg":    "token faild",
		})
		c.Abort()
		return
	}

	var challenge Challenge

	data, _ := ioutil.ReadAll(c.Request.Body)
	if err := json.Unmarshal(data, &challenge); err == nil {
		fmt.Println(challenge.Id)
	}
	fmt.Println(userInfo)

	//传入用户id与实验id
	code := 403
	if startChallenges(challenge.Id, userInfo.Id) {
		code = 200
	}
	c.JSON(http.StatusOK, gin.H{"message": "success!", "code": code})
}

//用户类型
type User struct {
	Username string `json:username`
	Password string `json:password`
}

//首页实验数据类型
type Challenge struct {
	Id          int    `jsong:id`
	Name        string `jsong:name`
	Img         string `json:img`
	Description string `json:descript`
	Type        int    `json:type`
	Open        int    `json:open`
	StartTime   int64  `json:startTime`
	Uid         string `json:uid`
	Url         string `json:url`
}

type Claims struct {
	Id       int    `json:"id"`       // id
	Username string `json:"username"` // 用户名
	Password string `json:"password"` // 密码
	jwt.StandardClaims
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
