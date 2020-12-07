package asset

import (
	"github.com/gin-gonic/gin"
	"github.com/uncopied/uncopier/api/v1.0/middleware"
)

// ApplyRoutes applies router to the gin Engine
func ApplyRoutes(r *gin.RouterGroup) {
	posts := r.Group("/asset")
	{
		posts.POST("/", middleware.Authorized, create)
		posts.GET("/", middleware.Authorized, list)
		posts.GET("/:id", middleware.Authorized, read)
		//posts.DELETE("/:id", middleware.Authorized, remove)
		//posts.PATCH("/:id", middleware.Authorized, update)
	}
}
