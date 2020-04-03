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
		challenge.GET("/get", API_GetChallenges)
		challenge.POST("/get", API_GetChallenge)
		challenge.POST("/getStatus", API_GetChallengeStatus)
		challenge.POST("/start", API_startChallenge)
		challenge.POST("/stop", API_stopChallenge)
		challenge.POST("/edit", API_editChallenge)
		challenge.POST("/add", API_addChallenge)
		challenge.POST("/del", API_delChallenge)
	}
	container := r.Group("/api/container/")
	container.Use(JWTAuth())
	{
		container.GET("/get", API_getContainer)
		container.POST("/stop", API_stopContainer)
		// container.POST("/getStatus", API_GetChallengeStatus)
		// container.POST("/start", API_startChallenge)
	}

}
