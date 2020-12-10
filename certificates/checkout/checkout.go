package checkout

import (
	"github.com/gin-gonic/gin"
	"github.com/uncopied/uncopier/api/v1.0/middleware"
)

// ApplyRoutes applies router to the gin Engine
func ApplyRoutes(r *gin.RouterGroup) {
	posts := r.Group("/checkout")
	{
		posts.GET("/:id", middleware.Authorized, checkout)
		posts.GET("/:id/success", middleware.Authorized, success)
		posts.GET("/:id/cancel", middleware.Authorized, cancel)
	}
}