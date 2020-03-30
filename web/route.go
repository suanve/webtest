package web

import (
	"github.com/gin-gonic/gin"
)

func routing(r *gin.Engine) {

	//基础api路由
	v1 := r.Group("/api/")
	{
		v1.GET("/", Index)
		v1.POST("/login", Login)
	}

	//前台实验api路由
	challenge := r.Group("/api/challenge/")
	challenge.Use(JWTAuth())
	{
		challenge.GET("/get", API_GetChallenge)
		challenge.POST("/getStatus", API_GetChallengeStatus)
		challenge.POST("/start", API_startChallenge)
		challenge.POST("/stop", API_stopChallenge)
	}
	container := r.Group("/api/container/")
	container.Use(JWTAuth())
	{
		container.GET("/list", API_container_Get)
		// container.POST("/getStatus", API_GetChallengeStatus)
		// container.POST("/start", API_startChallenge)
	}

}