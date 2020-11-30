package cert

import (
	"../../../database/dbmodel"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/gbrlsnchs/jwt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"time"
)

const secret = `
-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCjwdu1Mh8d3I08
oBuEOJVePgXj1wKyzu3qxjyVmCgikSB9XRegy/DPpX/n4uRqxVB4iPZXflGu7sch
FhCMdfxb4byBI9JF7p4rK3xHlab9a4EkUdvOEr28zTNgtwkmme6CDFbhtxXFQSjd
fPQ+BKYA0t2x8IHqAoI9dBCHLTj9BN7HPIagNFLv8gMM9SXEzA1FGHxp7OhoOuHZ
ltfAOKI1Suwvd51/oHSUojRcX0LoTeLJovvX/lJl1Mu72eW1RBKdBv7Sxk4sYmkb
1UCitMnCo+ZCqdb/8qGbjg+S7qelRb6jNj/H+brcHIUw3uMbz2goYP/NbCGU/4uG
XWYvXX11AgMBAAECggEBAISPnYdkd4P40exNv3idRWzw0FvL5cdRc48lwk1myraQ
vLg+762e6eVtl8jjBvzXlXi9hoz1GLJ/YHsMHYFW0V6fsbTohoNN0oQnw4c/QdrL
d9Mq4MBEs4tuoTSddq7k1Qo5aut1Bg6T3LzPNfguUyM/j29HviLsvPl6RxbmKMfI
KfBmGs5Enxvyp3DJkP1CvQDmEuB2kW8fGXYbVy+x3plACOO/OSBvbrTWp4fGrCo4
mNWMpAK/MU208MlCUNMH4PQXJgV7uXm0aRiTDxDWyYtSH056muIqFwfkrGzZJMkt
93MzhnU36bXOX1mQsAPYMmjxpJfaScXeL2wX7tHb8oECgYEA0OLqz1Sz9kA9vZ/y
x8Bo/PejKibJzY9JQ0n2Jqs9Fy6dcIA7e5KiwU4SEbXzWISHFsXV9uRZSWR+Ci65
XEOApFxgrAh/oYLru7zR1pYrjr4ez+cc6m685rv7ryocsssCq9LoxxECsCA5Io+n
cant0jbU83P+H9VWvauPak+izEkCgYEAyLEwQB+rdNSGBujw/K8JICavdaFXoqCo
OyafN/LLQha5XNxN2O2MTmctMUTl7DMmzwjtkvBZEB+OAVBZmjEeQ92rY7vtKFxE
OmKuEDgDnKLVwZm2YLjLD/co98RXf3KeBUcIdaRinbpwsDQ9A4S3/6xjbgdTPyWh
j0lZZ/mKr80CgYBjrgVzTu5Z8qoD1VIbtFvla573PG9MorXJYIAQT+LlLx9+UhMQ
kxcLu9+vh+5KLWPxoBLMsIdTGJt07HsT5jp7NIIFVkDhqAIqIp7YEe1TPrKhb55C
2PlX+hjOq//p6iqqKAlhBWMM/TOGpJq5COguSnAwhQed1UaBWF8l0j7T0QKBgGVc
eI4qcKJVJEwhInW8wdMnNr8meeh9U/psC0ZqrhX2/C/WZMsHTzHaEo0ryyR8wUEX
tUXddl4aUdKADoE+BZcpQgLhS2pzD1KdvGQcplZaN7PMOrynGIg7wMlCtR59eSoZ
MkCYgeY/3+Jev+IjCftrydwsfvMJwotn9Gv7MPyRAoGAdhk++VzYWrz65FL36Btj
NbrEVeXVA3z+ByoUXR+CKIInDSMEf+FDp8kE5EIhVd3UT0pEgLFjB94ZTbytoShD
qx+LF1b6Ndg156XG1xEThZB58zdZJdEz96NqGFkXHEBUCI9E7/j2OWLUeJ6eOUwo
wGS3ha8R0NfDrjnPv0Vrfgw=
-----END PRIVATE KEY-----
`

var privPEMData = []byte(secret)

type CustomPayload struct {
	jwt.Payload
	IssuerName string `json:"foo,omitempty"`
	UserRole   string `json:"foo,omitempty"`
	CertificateID uint `json:"foo,omitempty"`
}


func validateToken(token string) (CustomPayload, error) {
	block, rest := pem.Decode(privPEMData)
	if block == nil || block.Type != "PRIVATE KEY" {
		log.Fatal("failed to decode PEM block containing public key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Got a %T, with remaining data: %q", privateKey, rest)
	var privateKeyRsa = privateKey.(*rsa.PrivateKey)
	//var hs = jwt.NewHS256([]byte(secret))
	var hs = jwt.NewRS256(jwt.RSAPublicKey(&privateKeyRsa.PublicKey))

	var pl CustomPayload
	now := time.Now()
	aud := jwt.Audience{"https://uncopied.org"}

	// Validate claims "iat", "exp" and "aud".
	iatValidator := jwt.IssuedAtValidator(now)
	expValidator := jwt.ExpirationTimeValidator(now)
	audValidator := jwt.AudienceValidator(aud)

	// Use jwt.ValidatePayload to build a jwt.VerifyOption.
	// Validators are run in the order informed.
	validatePayload := jwt.ValidatePayload(&pl.Payload, iatValidator, expValidator, audValidator)
	_, err2 := jwt.Verify([]byte(token), hs, &pl, validatePayload)
	if err2 != nil {
		// ...
		log.Fatal(err2)
		return pl, err2
	}
	return pl, nil
}


func generateTokens(userName string, userRole string, certificateId uint) (string) {
	block, rest := pem.Decode(privPEMData)
	if block == nil || block.Type != "PRIVATE KEY" {
		log.Fatal("failed to decode PEM block containing public key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Got a %T, with remaining data: %q", privateKey, rest)
	var privateKeyRsa = privateKey.(*rsa.PrivateKey)

	//var hs = jwt.NewHS256([]byte(secret))
	var hs = jwt.NewRS256(jwt.RSAPublicKey(&privateKeyRsa.PublicKey), jwt.RSAPrivateKey(privateKeyRsa))
	now := time.Now()
	pl := CustomPayload{Payload: jwt.Payload{
		Issuer:         "uncopied",
		Subject:        "someone",
		Audience:       jwt.Audience{"https://uncopied.org", "https://uncopied.art"},
		ExpirationTime: jwt.NumericDate(now.Add(24 * 30 * 12 * time.Hour)),
		NotBefore:      jwt.NumericDate(now.Add(30 * time.Minute)),
		IssuedAt:       jwt.NumericDate(now),
		JWTID:          "uncopied",
	},
		IssuerName: userName,
		UserRole:   userRole,
		CertificateID: certificateId,
	}
	token, err := jwt.Sign(pl, hs)
	if err != nil {
		// ...
		fmt.Println("error generating JWT token for ", userName," role ",userRole)
	}
	return string(token)
}

func issue(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	userName := c.MustGet("user")
	// check if user
	var user dbmodel.User
	if err := db.Where("user_name = ?", userName).First(&user).Error; err != nil {
		fmt.Println("User name not found ",userName)
		c.AbortWithStatus(409)
		return
	}
	// additional checks on user?

	type RequestBody struct {
		Documentation  string `json:"documentation" binding:"required"`
	}

	var body RequestBody
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithStatus(400)
		return
	}

	// create user
	certificate := dbmodel.Certificate{
		Documentation: body.Documentation,
		Issuer:        user,
	}
	db.Create(&certificate)

	type RequestResponse struct {
		Certificate dbmodel.Certificate
		IssuerToken string
		OwnerToken string
		IssuerVerifierTokenLeft string
		AssetVerifierTokenLeft string
		OwnerVerifierTokenLeft string
		IssuerVerifierTokenRight string
		AssetVerifierTokenRight string
		OwnerVerifierTokenRight string
	}

	requestResponse := RequestResponse{
		Certificate:              certificate,
		IssuerToken:              generateTokens(user.UserName, "Issuer", certificate.ID),
		OwnerToken:               generateTokens(user.UserName, "Owner", certificate.ID),
		IssuerVerifierTokenLeft:  generateTokens(user.UserName, "IssuerVerifierLeft", certificate.ID),
		AssetVerifierTokenLeft:   generateTokens(user.UserName, "AssetVerifierLeft", certificate.ID),
		OwnerVerifierTokenLeft:   generateTokens(user.UserName, "AssetVerifierLeft", certificate.ID),
		IssuerVerifierTokenRight: generateTokens(user.UserName, "IssuerVerifierRight", certificate.ID),
		AssetVerifierTokenRight:  generateTokens(user.UserName, "AssetVerifierRight", certificate.ID),
		OwnerVerifierTokenRight:  generateTokens(user.UserName, "OwnerVerifierRight", certificate.ID),
	}

	c.JSON(200, requestResponse)
}



func action(c *gin.Context) {
	fmt.Println("control")
	db := c.MustGet("db").(*gorm.DB)
	token := c.Param("token")
	action := c.Param("action")

	userName := c.MustGet("user")
	// check if user
	var user dbmodel.User
	if err := db.Where("user_name = ?", userName).First(&user).Error; err != nil {
		fmt.Println("User name not found ",userName)
		c.AbortWithStatus(409)
		return
	}
	// additional checks on user?

	actionToken,err := validateToken(token)
	if err!=nil {
		log.Fatal(err)
	}

	var cert dbmodel.Certificate
	if err := db.Where("certificate_id = ?", actionToken.CertificateID).First(&cert).Error; err != nil {
		fmt.Println("Certificate not found ",actionToken.CertificateID)
		c.AbortWithStatus(409)
		return
	}

	var changeOwner = false
	var certifyAsset = false
	if actionToken.UserRole == "Owner" && action == "ChangeOwner" {
		// unlock change owner feature
		changeOwner = true
	}
	if actionToken.UserRole == "Issuer" && action == "CertifyAsset" {
		// unlock change owner feature
		certifyAsset = true
	}

	c.SetCookie("token", token, 60*60*24*7, "/", "uncopied.org", false, false)
	type RequestResponse struct {
		Certificate dbmodel.Certificate
		ChangeOwner bool
		CertifyAsset bool
	}

	response :=RequestResponse{
		Certificate:  cert,
		ChangeOwner:  changeOwner,
		CertifyAsset: certifyAsset,
	}
	c.JSON(200, response)
}
