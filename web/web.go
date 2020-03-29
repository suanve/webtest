package web

import "github.com/gin-gonic/gin"

func Run() {
	r := gin.Default()
	r.Use(CorsMiddleware())
	routing(r)
	// r.Static("/static", "web/static")
	r.Run(":8080") // listen and serve on 0.0.0.0:8080
}
