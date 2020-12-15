package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/uncopied/uncopier/api"
	"github.com/uncopied/uncopier/api/v1.0/middleware"
	"github.com/uncopied/uncopier/certificates"
	"github.com/uncopied/uncopier/database"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
)
// direct call to IPFS, http://127.0.0.1:8080/ipfs/QmYZ8w9v86HHUcxM8Yi1sBKNPgqNoL3jcitNUEtyWp5muP
// proxied call to IPDS http://127.0.0.1:8081/ipfs/QmYZ8w9v86HHUcxM8Yi1sBKNPgqNoL3jcitNUEtyWp5muP
// IPFS node should be configured with disabling of subdomains
// "PublicGateways": {
//			"localhost": {
//				"Paths": ["/ipfs"],
//				"UseSubdomains": false
//			},
//			"127.0.0.1": {
//				"Paths": ["/ipfs"],
//				"UseSubdomains": false
//			}
//		}
func ReverseProxyIPFS() gin.HandlerFunc {
	localIPFSHost:=os.Getenv("LOCAL_IPFS_HOST")
	localIPFSPort:=os.Getenv("LOCAL_IPFS_PORT")
	target := localIPFSHost+":"+localIPFSPort
	targetUrl,err := url.Parse("http://"+target)
	if err!=nil {
		log.Fatal(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(targetUrl)
	return func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func main() {

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	// initializes database
	db, _ := database.Initialize()
	db.Debug()
	router := gin.Default()
	// CORS for https://foo.com and https://github.com origins, allowing:
	// - PUT and PATCH methods
	// - Origin header
	// - Credentials share
	// - Preflight requests cached for 12 hours
	// TODO : tune this for prod
	router.Use(cors.Default())


	router.Use(static.ServeRoot("/", "./public")) // static files have higher priority over dynamic routes
	router.LoadHTMLGlob("templates/*")
	//router.LoadHTMLFiles("templates/template1.html", "templates/template2.html")
	router.GET("/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Main website",
			"artistName": "Mary Stone",
		})
	})
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "this call was relayed by the reverse proxy")
	}))
	defer backendServer.Close()

	// proxy for IPFS requests
	router.GET("/ipfs/:hash", ReverseProxyIPFS())

	router.Use(database.Inject(db))
	router.Use(middleware.JWTMiddleware())
	// apply api router
	api.ApplyRoutes(router)
	// apply certificates router
	certificates.ApplyRoutes(router)

	tls := os.Getenv("SERVER_TLS")
	serverHost := os.Getenv("SERVER_HOST")
	fmt.Printf("Serving with tlsMod %s\n", tls)
	if tls=="http" {
		// needs redirecting 80 port to 8081
		httpPort:=os.Getenv("SERVER_HTTP_PORT")
		router.Run(serverHost+":"+httpPort)
	} else if tls=="https" {
		// needs redirecting 443 port to 8443
		httpsPort:=os.Getenv("SERVER_HTTPS_PORT")
		fullChain:=os.Getenv("SERVER_HTTPS_FULLCHAIN")
		privKey:=os.Getenv("SERVER_HTTPS_PRIVKEY")
		router.RunTLS((serverHost+":"+httpsPort),fullChain,privKey) // listen and serve on 0.0.0.0:8443
	} else if tls=="autocert" {
		// needs root priviledge to run on 443
		domain1:=os.Getenv("SERVER_DOMAIN1")
		domain2:=os.Getenv("SERVER_DOMAIN2")
		log.Fatal(autotls.Run(router, domain1, domain2))
	}
}