package cert

import (
	"crypto/md5"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"github.com/gbrlsnchs/jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/uncopied/uncopier/database/dbmodel"
	"gorm.io/gorm"
	"io/ioutil"
	"log"
	"strconv"
	"time"
)
const UncopiedOrg = "uncopied.org"
const pkFile="keyPrivate.pem"

var privPEMData []byte = readPK()

func readPK() []byte {
	fmt.Println("init in certificates.go")
	privPEMData, err := ioutil.ReadFile(pkFile)
	if err != nil {
		log.Fatal(err)
	}
	return privPEMData
}


type CustomPayload struct {
	jwt.Payload
	IssuerName string
	UserRole   string
	CertificateID uint
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

func generateTokens(certificate dbmodel.Certificate, userRole string) (dbmodel.CertificateToken) {
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
	pl := CustomPayload{
		Payload: jwt.Payload{
		Issuer:         UncopiedOrg,
		Subject:        certificate.CertificateLabel,
		Audience:       jwt.Audience{"https://uncopied.org", "https://uncopied.art"},
		ExpirationTime: jwt.NumericDate(now.Add(24 * 30 * 12 * time.Hour)),
		NotBefore:      jwt.NumericDate(now.Add(30 * time.Minute)),
		IssuedAt:       jwt.NumericDate(now),
		JWTID:          string(certificate.ID)+"/"+userRole,
	},
		IssuerName: certificate.Issuer.UserName,
		UserRole:   userRole,
		CertificateID: certificate.ID,
	}
	token, err := jwt.Sign(pl, hs)
	if err != nil {
		// ...
		fmt.Println("error generating JWT token for ", certificate.Issuer.UserName," role ",userRole)
	}
	var h = md5.New()
	h.Write(token)
	tokenHash := hex.EncodeToString(h.Sum(nil))
	certificateToken := dbmodel.CertificateToken{
		CertificateID:   certificate.ID,
		Role:          userRole,
		Token:         string(token),
		TokenHash:     tokenHash,
	}
	return certificateToken
}

const uncopiedUsername = "uncopied"
func IssueCertificate(db *gorm.DB, user dbmodel.User, certificateLabel string, IsDIY bool) dbmodel.Certificate {
	// check if user
	var uncopied dbmodel.User
	if err := db.Where("user_name = ?", uncopiedUsername).First(&uncopied).Error; err != nil {
		fmt.Println("User name not found ",uncopiedUsername)
		log.Fatal("User name uncopied not found ")
		return dbmodel.Certificate{}
	}
	var printer dbmodel.User
	if IsDIY {
		printer = user
	} else {
		printer = uncopied
	}
	// create certificate
	certificate := dbmodel.Certificate{
		Issuer:                        user,
		Printer:                       printer,
		PrimaryConservator:            uncopied,
		SecondaryConservator:          uncopied,
		CertificateLabel:              certificateLabel,
	}
	db.Create(&certificate)

	issuerToken := generateTokens(certificate, "Issuer")
	ownerToken := generateTokens(certificate, "Owner")
	primaryAssetVerifierToken := generateTokens(certificate, "PrimaryAssetVerifier")
	secondaryAssetVerifierToken := generateTokens(certificate, "SecondaryAssetVerifier")
	primaryOwnerVerifierToken := generateTokens(certificate, "PrimaryOwnerVerifier")
	secondaryOwnerVerifierToken := generateTokens(certificate, "SecondaryOwnerVerifier")
	primaryIssuerVerifierToken := generateTokens(certificate, "PrimaryIssuerVerifier")
	secondaryIssuerVerifierToken := generateTokens(certificate, "SecondaryIssuerVerifier")

	db.Create(&issuerToken)
	db.Create(&ownerToken)
	db.Create(&primaryAssetVerifierToken)
	db.Create(&secondaryAssetVerifierToken)
	db.Create(&primaryOwnerVerifierToken)
	db.Create(&secondaryOwnerVerifierToken)
	db.Create(&primaryIssuerVerifierToken)
	db.Create(&secondaryIssuerVerifierToken)

	certificate.IssuerToken = issuerToken
	certificate.OwnerToken = ownerToken
	certificate.PrimaryAssetVerifierToken = primaryAssetVerifierToken
	certificate.SecondaryAssetVerifierToken = secondaryAssetVerifierToken
	certificate.PrimaryOwnerVerifierToken = primaryOwnerVerifierToken
	certificate.SecondaryOwnerVerifierToken = secondaryOwnerVerifierToken
	certificate.PrimaryIssuerVerifierToken = primaryIssuerVerifierToken
	certificate.SecondaryIssuerVerifierToken = secondaryIssuerVerifierToken

	db.Updates(&certificate)
	return certificate
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
		CertificateLabel  string `json:"label" binding:"required"`
	}
	var body RequestBody
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithStatus(400)
		return
	}
	certificate := IssueCertificate(db, user, body.CertificateLabel, true)
	type RequestResponse struct {
		Certificate dbmodel.Certificate
	}
	requestResponse := RequestResponse{
		Certificate: certificate,
	}
	c.JSON(200, requestResponse)
}

func action(c *gin.Context) {
	fmt.Println("control")
	db := c.MustGet("db").(*gorm.DB)
	certId := c.Param("certificates")
	token := c.Param("token")

	userName := c.MustGet("user")
	// check if user
	var user dbmodel.User
	if err := db.Where("user_name = ?", userName).First(&user).Error; err != nil {
		fmt.Println("User name not found ",userName)
		c.AbortWithStatus(409)
		return
	}

	var cert dbmodel.Certificate
	if err := db.Where("id = ?", certId).First(&cert).Error; err != nil {
		fmt.Println("Certificate not found ",certId)
		c.AbortWithStatus(409)
		return
	}

	var certToken dbmodel.CertificateToken
	if err := db.Where("certificate_id = ? AND token_hash = ? ", certId, token).First(&certToken).Error; err != nil {
		fmt.Printf("Certificate token not found certId = %s token = %s \n", certId, token)
		c.AbortWithStatus(409)
		return
	}

	// additional checks on token
	actionToken,err := validateToken(certToken.Token)
	if err!=nil {
		log.Fatal(err)
	}

	certIdInt, err := strconv.Atoi(certId)
	if err != nil {
		fmt.Printf("Certificate id unreadable %v \n", certId)
		c.AbortWithStatus(409)
		return
	}
	if  int(actionToken.CertificateID) != certIdInt {
		fmt.Printf("Certificate id mismatch not found %v != %v \n", certId, actionToken.CertificateID)
		c.AbortWithStatus(409)
		return
	}

	// add a cookie for the given role
	cookieName := actionToken.UserRole+"Token"
	c.SetCookie(cookieName, certToken.Token, 60*60*24*7, "/", "uncopied.org", false, false)
	type RequestResponse struct {
		Certificate dbmodel.Certificate
		UserRole string
		UserRoleToken string
	}

	response :=RequestResponse{
		Certificate :  cert,
		UserRole : actionToken.UserRole,
		UserRoleToken : certToken.Token,
	}
	c.JSON(200, response)
}
