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
	level := getLevel(user.Username)
	if password == user.Password {

		var claims Claims
		claims.Username = user.Username
		claims.Id = id
		claims.Level = level

		t, _ := CreateToken(&claims)
		c.JSON(http.StatusOK, gin.H{"code": 200, "isAdmin": claims.Level, "token": t})
	} else {
		c.JSON(http.StatusOK, gin.H{"code": 401})
	}
}

// Register 用于实现注册用户
func Register(c *gin.Context) {
	data, _ := ioutil.ReadAll(c.Request.Body)
	var user User

	if err := json.Unmarshal(data, &user); err == nil {
		fmt.Println(user.Username)
		fmt.Println(user.Password)
	}
	if user.Username == "" || user.Password == "" {
		c.JSON(http.StatusOK, gin.H{"code": 500})
	}
	code := 50
	user.Level = 0
	// 先判断用户是否存在
	resUser := getUserFromUsername(user.Username)
	if len(resUser) != 0 {
		c.JSON(http.StatusOK, gin.H{"message": "用户重复!", "code": code})
		return
	}
	res := addUser(user)
	if res == 1 {
		code = 200
	} else if res == 2 {
		code = 999
	}
	c.JSON(http.StatusOK, gin.H{"message": "success!", "code": code})
}

// 获取所有实验的信息
func API_getChallenges(c *gin.Context) {
	Challenges := getChallenges()
	c.JSON(http.StatusOK, gin.H{"message": "success", "length": len(Challenges), "data": Challenges})
}

// 获取指定实验的信息
func API_getChallenge(c *gin.Context) {
	// 获取token
	token := c.Request.Header.Get("token")
	// 获取用户名
	_, err := ValidateToken(token)
	if !err {
		c.JSON(http.StatusOK, gin.H{
			"status": -1,
			"msg":    "token faild",
		})
		c.Abort()
		return
	}

	data, _ := ioutil.ReadAll(c.Request.Body)
	//临时接受变量的结构体
	var challenge Challenge
	fmt.Println("data", string(data))
	if err := json.Unmarshal(data, &challenge); err == nil {
		fmt.Println("key:", challenge.Id)
	}
	Rescode := 401
	Challenges := getChallenge(challenge.Id)
	if len(Challenges) > 0 {
		Rescode = 200
	}

	c.JSON(http.StatusOK, gin.H{"message": "success", "code": Rescode, "length": len(Challenges), "data": Challenges})
}

// 根据token 获取用户身份，查找其启动的任务
func API_getChallengeStatus(c *gin.Context) {
	// 获取token
	token := c.Request.Header.Get("token")
	// 获取用户名
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
	res := startChallenge(challenge.Id, userInfo.Id, userInfo.Username)
	if res == 1 {
		code = 200
	} else if res == 2 {
		code = 999
	}

	c.JSON(http.StatusOK, gin.H{"message": "success!", "code": code})
}

//获取用户请求的实验id，停止对应的容器
func API_stopChallenge(c *gin.Context) {
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
	res := stopChallenge(challenge.Id, userInfo.Id, userInfo.Username)
	if res == 1 {
		code = 200
	} else if res == 2 {
		code = 999
	}

	c.JSON(http.StatusOK, gin.H{"message": "success!", "code": code})
}

//获取用户输入的信息,添加对应的实验
func API_addChallenge(c *gin.Context) {
	code := 403

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
	if userInfo.Level < 1 {
		c.JSON(http.StatusOK, gin.H{
			"status": -1,
			"msg":    "access faild",
		})
		c.Abort()
		return
	}
	var challenge Challenge

	data, _ := ioutil.ReadAll(c.Request.Body)
	if err := json.Unmarshal(data, &challenge); err == nil {

	}
	//传入用户id与实验id

	if challenge.Img == "" {
		challenge.Img = "https://gw.alipayobjects.com/zos/rmsportal/JiqGstEfoWAOHiTxclqi.png"
	}
	if challenge.Description == "" {
		challenge.Description = "nothing"
	}
	res := addChallenge(challenge)
	if res == 1 {
		code = 200
	} else if res == 2 {
		code = 999
	}

	c.JSON(http.StatusOK, gin.H{"message": "success!", "code": code})
}

func API_delChallenge(c *gin.Context) {
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
	if userInfo.Level < 1 {
		c.JSON(http.StatusOK, gin.H{
			"status": -1,
			"msg":    "access faild",
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
	res := delChallenge(challenge.Id)
	if res == 1 {
		code = 200
	} else if res == 2 {
		code = 999
	}

	c.JSON(http.StatusOK, gin.H{"message": "success!", "code": code})
}

//获取用户输入的信息,修改对应的实验信息
func API_editChallenge(c *gin.Context) {
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
	if userInfo.Level < 1 {
		c.JSON(http.StatusOK, gin.H{
			"status": -1,
			"msg":    "access faild",
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
	res := updateChallenge(challenge)
	if res == 1 {
		code = 200
	} else if res == 2 {
		code = 999
	}

	c.JSON(http.StatusOK, gin.H{"message": "success!", "code": code})
}

// 获取用户们开启的容器
func API_getContainer(c *gin.Context) {
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
	if userInfo.Level < 1 {
		c.JSON(http.StatusOK, gin.H{
			"status": -1,
			"msg":    "access faild",
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
	code := 200
	res := getContainers()
	length := len(res)
	c.JSON(http.StatusOK, gin.H{"message": "success!", "code": code, "length": length, "data": res})
}

// 后台停止用户开启的容器
func API_stopContainer(c *gin.Context) {
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
	if userInfo.Level < 1 {
		c.JSON(http.StatusOK, gin.H{
			"status": -1,
			"msg":    "access faild",
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
	code := 200
	res := stopContainer(challenge)

	c.JSON(http.StatusOK, gin.H{"message": "success!", "code": code, "data": res})
}

//获取当前服务器所有的用户
func API_getUsers(c *gin.Context) {
	res := getUsers()
	c.JSON(http.StatusOK, gin.H{"message": "success!", "length": len(res), "data": res})
}

//获取指定用户的信息
func API_getUser(c *gin.Context) {

	// 获取token
	token := c.Request.Header.Get("token")
	// 获取用户名
	userInfo, err := ValidateToken(token)
	if !err {
		c.JSON(http.StatusOK, gin.H{
			"status": -1,
			"msg":    "token faild",
		})
		c.Abort()
		return
	}

	if userInfo.Level < 1 {
		c.JSON(http.StatusOK, gin.H{
			"status": -1,
			"msg":    "access faild",
		})
		c.Abort()
		return
	}

	data, _ := ioutil.ReadAll(c.Request.Body)
	//临时接受变量的结构体
	var user User
	fmt.Println("data", string(data))
	if err := json.Unmarshal(data, &user); err == nil {
		fmt.Println("key:", user.Id)
	}
	Rescode := 401
	Challenges := getUser(user.Id)
	if len(Challenges) > 0 {
		Rescode = 200
	}

	c.JSON(http.StatusOK, gin.H{"message": "success", "code": Rescode, "length": len(Challenges), "data": Challenges})
}

// 添加用户
func API_addUser(c *gin.Context) {
	code := 403

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
	if userInfo.Level < 1 {
		c.JSON(http.StatusOK, gin.H{
			"status": -1,
			"msg":    "access faild",
		})
		c.Abort()
		return
	}
	var user User

	data, _ := ioutil.ReadAll(c.Request.Body)
	if err := json.Unmarshal(data, &user); err == nil {

	}

	res := addUser(user)
	if res == 1 {
		code = 200
	} else if res == 2 {
		code = 999
	}

	c.JSON(http.StatusOK, gin.H{"message": "success!", "code": code})
}

// 删除用户
func API_delUser(c *gin.Context) {
	code := 403
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
	if userInfo.Level < 1 {
		c.JSON(http.StatusOK, gin.H{
			"status": -1,
			"msg":    "access faild",
		})
		c.Abort()
		return
	}
	var user User

	data, _ := ioutil.ReadAll(c.Request.Body)
	if err := json.Unmarshal(data, &user); err == nil {

	}

	res := delUser(user.Id)
	if res == 1 {
		code = 200
	} else if res == 2 {
		code = 999
	}

	c.JSON(http.StatusOK, gin.H{"message": "success!", "code": code})
}

// 获取用户输入的信息,修改对应的用户
func API_editUser(c *gin.Context) {
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
	if userInfo.Level < 1 {
		c.JSON(http.StatusOK, gin.H{
			"status": -1,
			"msg":    "access faild",
		})
		c.Abort()
		return
	}
	var user User

	data, _ := ioutil.ReadAll(c.Request.Body)
	if err := json.Unmarshal(data, &user); err == nil {
		fmt.Println(user.Id)
	}
	fmt.Println(userInfo)

	//传入用户id与实验id
	code := 403
	res := updateUser(user)
	if res == 1 {
		code = 200
	} else if res == 2 {
		code = 999
	}

	c.JSON(http.StatusOK, gin.H{"message": "success!", "code": code})
}

// 修改用户密码
func API_UpdatePass(c *gin.Context) {
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
	type TmpUser struct {
		OldPass   string `json:"old_Password"`
		NewPass   string `json:"new_Password"`
		CheckPass string `json:"check_Password"`
	}

	var user TmpUser
	data, _ := ioutil.ReadAll(c.Request.Body)
	if err := json.Unmarshal(data, &user); err == nil {
	}

	// 确认旧密码是否相等
	fmt.Println(userInfo)
	fmt.Println(user)
	resUser := getUserFromUsername(userInfo.Username)
	if len(resUser) != 1 {
		c.JSON(http.StatusOK, gin.H{"message": "未知错误!", "code": 500})
		return
	}

	if resUser[0].Password != user.OldPass {
		c.JSON(http.StatusOK, gin.H{"message": "旧密码不正确!", "code": 500})
		return
	}
	if user.NewPass != user.CheckPass {
		c.JSON(http.StatusOK, gin.H{"message": "新密码输入不相同!", "code": 500})
		return
	}
	if user.NewPass == "" || user.CheckPass == "" {
		c.JSON(http.StatusOK, gin.H{"message": "密码输入错误!", "code": 500})
		return
	}

	code := 403
	res := updateUser(User{
		Id:       resUser[0].Id,
		Username: resUser[0].Username,
		Password: user.NewPass,
		Level:    resUser[0].Level,
	})
	if res == 1 {
		code = 200
	} else if res == 2 {
		code = 999
	}
	c.JSON(http.StatusOK, gin.H{"message": "success!", "code": code})
}
