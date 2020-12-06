package auth

import (
	"fmt"
	"github.com/gbrlsnchs/jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/uncopied/uncopier/database/dbmodel"
	"gorm.io/gorm"
	"time"
	//https://www.gregorygaines.com/blog/posts/2020/6/11/how-to-hash-and-salt-passwords-in-golang-using-sha512-and-why-you-shouldnt
	"golang.org/x/crypto/bcrypt"
)


func hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

func checkHash(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}


const secret = "secret"
var hs = jwt.NewHS256([]byte(secret))

type CustomPayload struct {
	jwt.Payload
	UserName string `json:"foo,omitempty"`
}

func generateToken(userName string) (string, error) {
	now := time.Now()
	pl := CustomPayload{ Payload: jwt.Payload{
			Issuer:         "uncopied",
			Subject:        "someone",
			Audience:       jwt.Audience{"https://uncopied.org", "https://uncopied.art"},
			ExpirationTime: jwt.NumericDate(now.Add(24 * 30 * 12 * time.Hour)),
			NotBefore:      jwt.NumericDate(now.Add(30 * time.Minute)),
			IssuedAt:       jwt.NumericDate(now),
			JWTID:          "uncopied",
		},
		UserName: userName,
	}


	token, err := jwt.Sign(pl, hs)
	if err != nil {
		// ...
		fmt.Println("error generating JWT token for ",userName)
	}
	return string(token),err
}

func register(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	type RequestBody struct {
		UserName    string `json:"username" binding:"required"`
		DisplayName string `json:"display_name" binding:"required"`
		EmailAddress string `json:"email" binding:"required"`
		Password    string `json:"password" binding:"required"`
	}

	var body RequestBody
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithStatus(400)
		return
	}

	// check if exists
	var exists dbmodel.User
	if err := db.Where("user_name = ?", body.UserName).First(&exists).Error; err == nil {
		c.AbortWithStatus(409)
		return
	}

	hash, hashErr := hash(body.Password)
	if hashErr != nil {
		c.AbortWithStatus(500)
		return
	}

	// create user
	user := dbmodel.User{
		UserName:     body.UserName,
		DisplayName:  body.DisplayName,
		EmailAddress: body.EmailAddress,
		PasswordHash: hash,
	}
	db.Create(&user)

	token, _ := generateToken(body.UserName)
	c.SetCookie("token", token, 60*60*24*7, "/", "", false, true)
	authToken := Token{
		UserName: body.UserName,
		Token:    token,
	}
	c.JSON(200, authToken)
}

type Token struct {
	UserName    string `json:"user" binding:"required"`
	Token string `json:"token" binding:"required"`
}

func login(c *gin.Context) {
	fmt.Println("login")
	db := c.MustGet("db").(*gorm.DB)
	type RequestBody struct {
		UserName string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	var body RequestBody
	if err := c.BindJSON(&body); err != nil {
		fmt.Println("login body err")
		c.AbortWithStatus(400)
		return
	}

	// check existancy
	var user dbmodel.User
	if err := db.Where("user_name = ?", body.UserName).First(&user).Error; err != nil {
		fmt.Println("User name not found ",body.UserName)
		c.AbortWithStatus(404) // user not found
		return
	}

	if !checkHash(body.Password, user.PasswordHash) {
		fmt.Println("User password mismatch ",body.Password)
		c.AbortWithStatus(401)
		return
	}

	token, _ := generateToken(body.UserName)

	c.SetCookie("token", token, 60*60*24*7, "/", "uncopied.org", false, false)
	authToken := Token{
		UserName: body.UserName,
		Token:    token,
	}
	c.JSON(200, authToken)
}

// check API will renew token when token life is less than 3 days, otherwise, return null for token
func check(c *gin.Context) {
	userRaw, ok := c.Get("user")
	if !ok {
		c.AbortWithStatus(401)
		return
	}

	userName := userRaw.(string)

	tokenExpire := int64(c.MustGet("token_expire").(float64))
	now := time.Now().Unix()
	diff := tokenExpire - now

	fmt.Println(diff)
	if diff < 60*60*24*3 {
		// renew token
		token, _ := generateToken(userName)
		c.SetCookie("token", token, 60*60*24*7, "/", "", false, true)
		authToken := Token{
			UserName: userName,
			Token:    token,
		}
		c.JSON(200, authToken)
		return
	}
	authToken := Token{
		UserName: userName,
		Token:    "",
	}
	c.JSON(200, authToken	)
}
