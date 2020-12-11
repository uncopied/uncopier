package verify

import (
	"github.com/gin-gonic/gin"
)

// ApplyRoutes applies router to the gin Engine
func ApplyRoutes(r *gin.RouterGroup) {
	posts := r.Group("/y")
	{
		// preview by ID
		posts.GET("/:id", verify)
	}
}