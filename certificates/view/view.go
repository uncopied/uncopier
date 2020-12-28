package view

import (
	"github.com/gin-gonic/gin"
	"github.com/uncopied/uncopier/api/v1.0/middleware"
	"os"
)


type BaseURL struct {
	ServerBaseURLInternal string
	ServerBaseURLExternal string
}


func ServerBaseURL() BaseURL {
	// URL for preview
	tls := os.Getenv("SERVER_TLS")
	serverHost := os.Getenv("SERVER_HOST")
	serverBaseURLExternal := ""
	serverBaseURLInternal := ""
	if tls=="http" {
		// needs redirecting 80 port to 8081
		httpPort:=os.Getenv("SERVER_HTTP_PORT")
		serverBaseURLInternal="http://"+serverHost+":"+httpPort
		serverBaseURLExternal=serverBaseURLInternal
	} else if tls=="https" {
		// needs redirecting 443 port to 8443
		httpsPort:=os.Getenv("SERVER_HTTPS_PORT")
		serverBaseURLInternal="https://"+serverHost+":"+httpsPort
		serverBaseURLExternal="https://"+serverHost
	} else if tls=="autocert" {
		// needs root priviledge to run on 443
		serverBaseURLInternal="https://"+serverHost
		serverBaseURLExternal=serverBaseURLInternal
	}
	baseURL := BaseURL{
		ServerBaseURLInternal: serverBaseURLInternal,
		ServerBaseURLExternal: serverBaseURLExternal,
	}
	return baseURL
}

// ApplyRoutes applies router to the gin Engine
func ApplyRoutes(r *gin.RouterGroup) {
	posts := r.Group("/preview")
	{
		// preview by ID
		posts.GET("/:uuid", middleware.Authorized, preview)
	}
	views := r.Group("/v")
	{
		// preview by token (this is PUBLIC)
		views.GET("/:token", view)
	}
}