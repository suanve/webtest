package web

import (
	"github.com/gin-gonic/gin"
)

func routing(r *gin.Engine) {

	//基础api路由
	v1 := r.Group("/")
	{
		v1.GET("/", Index)
		v1.POST("/login", Login)
	}

	//前台实验api路由
	challenge := r.Group("/api")
	challenge.Use(JWTAuth())
	{
		challenge.GET("/challenge/get", API_GetChallenge)
		challenge.POST("/challenge/getStatus", API_GetChallengeStatus)
		challenge.POST("/challenge/start", API_startChallenge)
	}

}
