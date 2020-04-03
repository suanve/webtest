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
		challenge.GET("/get", API_getChallenges)
		challenge.POST("/get", API_getChallenge)
		challenge.POST("/getStatus", API_getChallengeStatus)
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
	user := r.Group("/api/user/")
	user.Use(JWTAuth())
	{
		user.GET("/get", API_getUsers)
		user.POST("/get", API_getUser)
		user.POST("/add", API_addUser)
		user.POST("/del", API_delUser)
		user.POST("/edit", API_editUser)
		// container.POST("/getStatus", API_GetChallengeStatus)
		// container.POST("/start", API_startChallenge)
	}
}
