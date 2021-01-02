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
	"github.com/uncopied/uncopier/upload"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
	"github.com/gin-contrib/gzip"
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

// https://github.com/gin-gonic/gin/issues/1543
func headersByRequestURI() gin.HandlerFunc {
	return func(c *gin.Context) {
		tls := os.Getenv("SERVER_TLS")
		if tls == "https" {
			serverHost := os.Getenv("SERVER_HOST")
			// https://blog.bracebin.com/achieving-perfect-ssl-labs-score-with-go
			c.Header("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
			c.Header("Content-Security-Policy", "default-src 'self' https://"+serverHost+"; font-src 'self'; img-src 'self' http://www.w3.org; script-src 'self'; frame-src 'self'; style-src 'self' https://fonts.googleapis.com https://cdn.jsdelivr.net;")
			c.Header("X-Frame-Options","SAMEORIGIN")
			// INLINE_RUNTIME_CHUNK=false
		}
		if strings.HasPrefix(c.Request.RequestURI, "/static/") ||
			strings.HasPrefix(c.Request.RequestURI, "/assets/") ||
			strings.HasPrefix(c.Request.RequestURI, "/docs/") {
			// 10 days
			c.Header("Cache-Control", "max-age=864000")
			//c.Header("Content-Description", "File Transfer")
			//c.Header("Content-Type", "application/octet-stream")
			//c.Header("Content-Transfer-Encoding", "binary")
		} else if !strings.HasPrefix(c.Request.RequestURI, "/api/") {
			// default : one day
			c.Header("Cache-Control", "max-age=86400")
		}
	}
}

func main() {

	fmt.Println("odotenv.Load()")
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	// initializes database
	db, _ := database.Initialize()
	db.Debug()
	router := gin.Default()
	// gzip as per https://www.webpagetest.org/ reco
	router.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedExtensions([]string{
		".png", ".gif", ".jpeg", ".jpg",
		".pdf",
	})))

	// max-age or expires
	// https://github.com/gin-gonic/gin/issues/1543
	router.Use(headersByRequestURI())

	// CORS for https://foo.com and https://github.com origins, allowing:
	// - PUT and PATCH methods
	// - Origin header
	// - Credentials share
	// - Preflight requests cached for 12 hours
	// TODO : tune this for prod
	// Also beware there is a bug when calling http://localhost:8081/api/v1.0/src (redirect 307 not working with CORS)
	// so need to call http://localhost:8081/api/v1.0/src/ to avoid redirects
	router.Use(cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
		AllowFiles: true,
		AllowAllOrigins: true,
	}))

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
	// ipfs upload
	upload.ApplyRoutes(router)

	// todo : configure some read/write timeout?
	//https://github.com/gin-contrib/timeout

	// todo : add https://prometheus.io/
	// good checklist here
	// https://github.com/fiorix/freegeoip

	// todo : Boost Ubuntu Network Performance by Enabling TCP BBR
	// https://www.linuxbabe.com/ubuntu/enable-google-tcp-bbr-ubuntu

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