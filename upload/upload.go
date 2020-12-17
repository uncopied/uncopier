package upload

import (
	"github.com/gin-gonic/gin"
)

// ApplyRoutes applies router to gin Router
func ApplyRoutes(r *gin.Engine) {
	posts := r.Group("/upload")
	{
		// preview by ID
		posts.POST("/", // middleware.Authorized,
			upload)
	}
}

