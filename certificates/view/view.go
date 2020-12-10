package view

import (
	"github.com/gin-gonic/gin"
	"github.com/uncopied/uncopier/api/v1.0/middleware"
)

// ApplyRoutes applies router to the gin Engine
func ApplyRoutes(r *gin.RouterGroup) {
	posts := r.Group("/preview")
	{
		// preview by ID
		posts.GET("/:id", middleware.Authorized, preview)
	}
	views := r.Group("/v")
	{
		// preview by token (this is PUBLIC)
		views.GET("/:token", view)
	}
}