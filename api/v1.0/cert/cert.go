package cert

import (
	"github.com/gin-gonic/gin"
	"github.com/uncopied/uncopier/api/v1.0/middleware"
)

const mailTo = "NAMSOR SAS\nBP 40373\n78000 VERSAILLES\nCEDEX FRANCE"

func MailTo() string {
	return mailTo
}

const topHelper = "This is a chirograph by UNCOPIED with 5 parts (CENTER,TOP,BOTTOM,LEFT,RIGHT):\n-1 Print it on quality 100+ gram paper (ex. C by Clairfontaine recommended https://uncopied.art/blog/papers or other ISO-9706).\n-2 Cut it using scissors into 5 parts along the blue lines (slight imperfections are OK, make it unique and tamper-proof).\n-3 Optionally, add your own identical physical mark or signature to each of the 5 parts (without corrupting the QR Codes).\n-4 Glue its CENTER part to the back of your painting or limited edition print (ex. 3M recommended https://uncopied.art/blog/glues)."
const leftHelper = "Please, do not forget to send this LEFT part by post. We will activate the digital certificate within 3 days of receiving the physical copy in our PO Box. The LEFT part will be archived by UNCOPIED internally."
const rightHelper = "Please, do not forget to send this RIGHT part by post. We will activate the digital certificate within 3 days of receiving the physical copy in our PO Box. The RIGHT part will be archived by a third party."
const bottomHelper = "-5 Glue the other 4 parts to each 4 copies of the legal documentation (make sure labels correspond).\n-6 Retain one copy of the legal document (the one with the TOP chirograph part).\n-7 Send one copy of the legal document to the other contracting party (the one with the BOTTOM chirograph part).\n-8 Send the other two copies of the legal documents to UNCOPIED PO Box by post (the ones with LEFT and RIGHT chirograph parts)."

func TopHelper() string {
	return topHelper
}
func LeftHelper() string {
	return leftHelper
}
func RightHelper() string {
	return rightHelper
}
func BottomHelper() string {
	return bottomHelper
}

// ApplyRoutes applies router to the gin Engine
func ApplyRoutes(r *gin.RouterGroup) {
	cert := r.Group("/cert")
	{
		cert.GET("/:id", middleware.Authorized, preview)
		cert.POST("/order", middleware.Authorized, order)
	}
	order := r.Group("/order")
	{
		order.GET("/checkout/:uuid", middleware.Authorized, checkout)
		order.POST("/process", middleware.Authorized, process)
	}
}
