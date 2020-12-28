	package cert

import (
	"github.com/gin-gonic/gin"
	"github.com/uncopied/uncopier/api/v1.0/middleware"
)

const mailTo = "NAMSOR SAS\nBP 40373\n78000 VERSAILLES\nCEDEX FRANCE"

func MailTo( ) string {
	return mailTo
}

// ApplyRoutes applies router to the gin Engine
func ApplyRoutes(r *gin.RouterGroup) {
	cert:= r.Group("/cert")
	{
		cert.GET("/:id", middleware.Authorized, preview)
		cert.POST("/order", middleware.Authorized, order)
	}
	order:= r.Group("/order")
	{
		order.GET("/checkout/:uuid", middleware.Authorized, checkout)
		order.POST("/process", middleware.Authorized, process)
	}
}

