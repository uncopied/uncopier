	package cert

import (
	"github.com/gin-gonic/gin"
)

const mailTo = "NAMSOR SAS\nBP 40373\n78000 VERSAILLES\nCEDEX FRANCE"

func MailTo( ) string {
	return mailTo
}

// ApplyRoutes applies router to the gin Engine
func ApplyRoutes(r *gin.RouterGroup) {
	cert:= r.Group("/certificates")
	{
		cert.POST("/issue", issue)
		cert.GET("/:certificates/:token", action)
	}
}
