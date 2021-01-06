package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/nanmu42/gzip"
	//"github.com/gin-contrib/gzip"
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
	"net/http/httputil"
	"net/url"
	"crypto/tls"
	"os"
	"strings"
	"time"
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
			//'https://www.paypal.com/xoplatform
			c.Header("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
			c.Header("Content-Security-Policy", "base-uri 'self'; default-src 'self' https://"+serverHost+"; font-src 'self' data: ; connect-src 'self' https://www.paypal.com/ ; img-src 'self' data: http://www.w3.org https://t.paypal.com/; script-src 'self' 'unsafe-inline' https://www.paypal.com/; frame-src 'self' https://www.paypal.com/; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://cdn.jsdelivr.net;")
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
	// but the default gzip middleware caused net::ERR_HTTP2_PROTOCOL_ERROR errors
	//router.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedExtensions([]string{
	//	".png", ".gif", ".jpeg", ".jpg",
	//	".pdf",
	//	})))
	var otherExtensions = []string{".ttf"}

	gzipHandler := gzip.NewHandler(gzip.Config{
		// gzip compression level to use
		CompressionLevel: 6,
		// minimum content length to trigger gzip, the unit is in byte.
		MinContentLength: 1024,
		// RequestFilter decide whether or not to compress response judging by request.
		// Filters are applied in the sequence here.
		RequestFilter: []gzip.RequestFilter{
			gzip.NewExtensionFilter(otherExtensions),
			gzip.DefaultExtensionFilter(),
		},
		// ResponseHeaderFilter decide whether or not to compress response
		// judging by response header
		ResponseHeaderFilter: []gzip.ResponseHeaderFilter{
			gzip.DefaultContentTypeFilter(),
		},
	})
	router.Use(gzipHandler.Gin)

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

	// todo : add https://prometheus.io/
	// good checklist here
	// https://github.com/fiorix/freegeoip

	// add no route to redirect to / : catch all route for React
	router.NoRoute(func(c *gin.Context) {
		// https://stackoverflow.com/questions/27928372/react-router-urls-dont-work-when-refreshing-or-writing-manually
		c.Redirect(http.StatusMovedPermanently, "/")
	})

	tlsKind := os.Getenv("SERVER_TLS")
	serverHost := os.Getenv("SERVER_HOST")
	fmt.Printf("Serving with tlsMod %s\n", tlsKind)
	if tlsKind =="http" {
		// needs redirecting 80 port to 8081
		httpPort:=os.Getenv("SERVER_HTTP_PORT")
		err = router.Run(serverHost+":"+httpPort)
		if err!=nil {
			log.Fatal(err)
		}
	} else if tlsKind =="https" {
		// needs redirecting 443 port to 8443
		httpsPort:=os.Getenv("SERVER_HTTPS_PORT")
		fullChain:=os.Getenv("SERVER_HTTPS_FULLCHAIN")
		privKey:=os.Getenv("SERVER_HTTPS_PRIVKEY")
		addr:=serverHost+":"+httpsPort
		server := &http.Server{
			Addr:              addr,
			Handler:           router,
			TLSConfig: &tls.Config{
				PreferServerCipherSuites: true,
				CurvePreferences: []tls.CurveID{
					tls.CurveP256,
					tls.X25519,
				},
				MinVersion: tls.VersionTLS12,
			},
			ReadTimeout:       30 * time.Second,
			ReadHeaderTimeout: 30 * time.Second,
			WriteTimeout:      120 * time.Second,
		}
		err := server.ListenAndServeTLS(fullChain, privKey)
		// err = http.ListenAndServeTLS(addr, fullChain, privKey, router)
		//err = router.RunTLS(addr,fullChain,privKey) // listen and serve on 0.0.0.0:8443
		if err!=nil {
			log.Fatal(err)
		}
		// to monitor files open lsof -p [PID_ID]
		// TODO: set limits ulimit -n 65535
		// https://medium.com/@muhammadtriwibowo/set-permanently-ulimit-n-open-files-in-ubuntu-4d61064429a
		httpPort:=os.Getenv("SERVER_HTTP_PORT")
		addr2:=serverHost+":"+httpPort
		// https://blog.cloudflare.com/exposing-go-on-the-internet/
		srv := &http.Server{
			Addr: addr2,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Connection", "close")
				url := "https://" + req.Host + req.URL.String()
				http.Redirect(w, req, url, http.StatusMovedPermanently)
			}),
		}
		go func() { log.Fatal(srv.ListenAndServe()) }( )
	} else if tlsKind =="autocert" {
		// needs root priviledge to run on 443
		domain1:=os.Getenv("SERVER_DOMAIN1")
		domain2:=os.Getenv("SERVER_DOMAIN2")
		err = autotls.Run(router, domain1, domain2)
		if err!=nil {
			log.Fatal(err)
		}
	}
}