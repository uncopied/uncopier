package apiv1

import (
	"github.com/gin-gonic/gin"
	"github.com/uncopied/uncopier/auth"
	"github.com/uncopied/uncopier/src"
	"github.com/uncopied/uncopier/cert"
)



func ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

// ApplyRoutes applies router to the gin Engine
func ApplyRoutes(r *gin.RouterGroup) {
	v1 := r.Group("/v1.0")
	{
		v1.GET("/ping", ping)
		auth.ApplyRoutes(v1)
		src.ApplyRoutes(v1)
		cert.ApplyRoutes(v1)
	}
}
