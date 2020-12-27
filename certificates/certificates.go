package certificates

import (
	"github.com/gin-gonic/gin"
	"github.com/uncopied/uncopier/certificates/checkout"
	"github.com/uncopied/uncopier/certificates/verify"
	"github.com/uncopied/uncopier/certificates/view"
)


// ApplyRoutes applies router tow gin Router
func ApplyRoutes(r *gin.Engine) {
	cert := r.Group("/c")
	{
		view.ApplyRoutes(cert)
		checkout.ApplyRoutes(cert)
		verify.ApplyRoutes(cert)
	}
}

