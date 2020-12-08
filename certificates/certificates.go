package certificates

import (
	"github.com/gin-gonic/gin"
	"github.com/uncopied/uncopier/certificates/preview"
)

// ApplyRoutes applies router to gin Router
func ApplyRoutes(r *gin.Engine) {
	cert := r.Group("/c")
	{
		preview.ApplyRoutes(cert)
	}
}
