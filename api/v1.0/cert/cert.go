package cert

import (
	"github.com/gin-gonic/gin"
)

// ApplyRoutes applies router to the gin Engine
func ApplyRoutes(r *gin.RouterGroup) {
	cert:= r.Group("/cert")
	{
		cert.POST("/issue", issue)
		cert.GET("/:cert/:token", action)
	}
}
