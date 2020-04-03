package web

import (
	"github.com/dgrijalva/jwt-go"
)

//用户数据结构
type User struct {
	Id       int    `json:id`
	Username string `json:username`
	Password string `json:password`
	Level    int    `json:level`
}

//首页实验数据结构
type Challenge struct {
	Id          int    `json:id`
	Name        string `json:name`
	Img         string `json:img`
	Description string `json:descript`
	Type        int    `json:type`
	Open        int    `json:open`
	StartTime   int64  `json:startTime`
	Uid         string `json:uid`
	Url         string `json:url`
	Key         int    `json:key`
	Image       string `json:image`
	Inport      int    `json:inport`
	Username    string `json:username`
}

// JWT认证数据结构
type Claims struct {
	Id       int    `json:"id"`       // id
	Username string `json:"username"` // 用户名
	Password string `json:"password"` // 密码
	Level    int    `json:"level"`    // 密码
	jwt.StandardClaims
}

// 获取到的任务数据结构
type Tasks struct {
	Id          int
	ChallengeId int
	Start       int
	Userid      int
	Url         string
	ContainerId string
}

// 选择的项目数据结构
type Items struct {
	Id      int
	Content string
	Time    int64
	Uid     int
	Status  int
}

type Image struct {
	Id          int
	image       string
	name        string
	challengeId int
	inport      int
}
