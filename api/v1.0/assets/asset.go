package assets

import (
	"../middleware"
	"github.com/gin-gonic/gin"
)

// ApplyRoutes applies router to the gin Engine
func ApplyRoutes(r *gin.RouterGroup) {
	posts := r.Group("/assets")
	{
		posts.POST("/", middleware.Authorized, create)
		posts.GET("/", list)
		posts.GET("/:id", read)
		posts.DELETE("/:id", middleware.Authorized, remove)
		posts.PATCH("/:id", middleware.Authorized, update)
	}
}
