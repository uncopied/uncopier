package src

import (
	"github.com/uncopied/uncopier/middleware"
	"github.com/gin-gonic/gin"
)

// ApplyRoutes applies router to the gin Engine
func ApplyRoutes(r *gin.RouterGroup) {
	posts := r.Group("/src")
	{
		posts.POST("/", middleware.Authorized, create)
		posts.GET("/", middleware.Authorized, list)
		posts.GET("/:id", middleware.Authorized, read)
		//posts.DELETE("/:id", middleware.Authorized, remove)
		//posts.PATCH("/:id", middleware.Authorized, update)
	}
}
