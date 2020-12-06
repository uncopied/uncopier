package middleware

import (
	"github.com/gbrlsnchs/jwt"
	"github.com/gin-gonic/gin"
	"github.com/uncopied/uncopier/api/v1.0/auth"
	"log"
	"strings"
	"time"
)

const secret = "secret"
var hs = jwt.NewHS256([]byte(secret))

func validateToken(token string) (string, error) {
	var pl auth.CustomPayload
	now := time.Now()
	aud := jwt.Audience{"https://uncopied.org"}

	// Validate claims "iat", "exp" and "aud".
	iatValidator := jwt.IssuedAtValidator(now)
	expValidator := jwt.ExpirationTimeValidator(now)
	audValidator := jwt.AudienceValidator(aud)

	// Use jwt.ValidatePayload to build a jwt.VerifyOption.
	// Validators are run in the order informed.
	validatePayload := jwt.ValidatePayload(&pl.Payload, iatValidator, expValidator, audValidator)
	_, err := jwt.Verify([]byte(token), hs, &pl, validatePayload)
	if err != nil {
		// ...
		log.Fatal(err)
		return "", err
	}
	return pl.UserName, nil
}

// JWTMiddleware parses JWT token from cookie and stores data and expires date to the context
// JWT Token can be passed as cookie, or Authorization header
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("token")
		// failed to read cookie
		if err != nil {
			// try reading HTTP Header
			authorization := c.Request.Header.Get("Authorization")
			if authorization == "" {
				c.Next()
				return
			}
			sp := strings.Split(authorization, "Bearer ")
			// invalid token
			if len(sp) < 1 {
				c.Next()
				return
			}
			tokenString = sp[1]
		}
		tokenData, err := validateToken(tokenString)
		if err != nil {
			c.Next()
			return
		}

		c.Set("user", tokenData)
		c.Next()
	}
}
