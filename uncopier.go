package main

import (
	"flag"
	"fmt"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/uncopied/uncopier/api"
	"github.com/uncopied/uncopier/api/v1.0/middleware"
	"github.com/uncopied/uncopier/certificates"
	"github.com/uncopied/uncopier/database"
	"github.com/gin-gonic/autotls"
	"log"
	"net/http"
)




func main() {
	tls := flag.String("tls", "http", "TLS mod : http, https or autocert")

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	// initializes database
	db, _ := database.Initialize()
	db.Debug()

	router := gin.Default()
	router.Use(static.ServeRoot("/public", "./public")) // static files have higher priority over dynamic routes
	router.Static("/assets", "./assets")
	router.Static("/img", "./img")
	router.LoadHTMLGlob("templates/*")

	//router.LoadHTMLFiles("templates/template1.html", "templates/template2.html")
	router.GET("/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Main website",
			"artistName": "Mary Stone",
		})
	})
	router.Use(database.Inject(db))
	router.Use(middleware.JWTMiddleware())
	// apply api router
	api.ApplyRoutes(router)
	// apply certificates router
	certificates.ApplyRoutes(router)
	fmt.Printf("Serving with tlsMod %s\n", *tls)
	if *tls=="http" {
		router.Run(":8081")
	} else if *tls=="https" {
		router.RunTLS((":8443"),"/etc/letsencrypt/live/uncopied.art/fullchain.pem","/etc/letsencrypt/live/uncopied.art/privkey.pem") // listen and serve on 0.0.0.0:8443
	} else if *tls=="autocert" {
		log.Fatal(autotls.Run(router, "uncopied.org", "uncopied.art"))
	}
}