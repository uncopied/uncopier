package cert

import (
	"bytes"
	"crypto/md5"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"github.com/gbrlsnchs/jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/uncopied/tallystick"
	"github.com/uncopied/uncopier/certificates/view"
	"github.com/uncopied/uncopier/database/dbmodel"
	"gorm.io/gorm"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"
)

func PermanentCertificateURL(certificateIssuanceID uint) string {
	serverHost := os.Getenv("SERVER_HOST")
	return "https://"+serverHost+"/c/y/" + strconv.Itoa(int(certificateIssuanceID))
}

func PermanentCertificateTokenURL(token string) string {
	serverHost := os.Getenv("SERVER_HOST")
	return "https://"+serverHost+"/c/t/" + token
}


var privPEMData []byte = readPK()
const pkFile="keyPrivate.pem"
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
	block, _ := pem.Decode(privPEMData)
	if block == nil || block.Type != "PRIVATE KEY" {
		log.Fatal("failed to decode PEM block containing public key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		log.Fatal(err)
	}
	var privateKeyRsa = privateKey.(*rsa.PrivateKey)
	//var hs = jwt.NewHS256([]byte(secret))
	var hs = jwt.NewRS256(jwt.RSAPublicKey(&privateKeyRsa.PublicKey))

	var pl CustomPayload
	now := time.Now()
	uncopiedDomainPrimary := os.Getenv("UNCOPIED_DOMAIN_PRIMARY")
	aud := jwt.Audience{uncopiedDomainPrimary}

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
	block, _ := pem.Decode(privPEMData)
	if block == nil || block.Type != "PRIVATE KEY" {
		log.Fatal("failed to decode PEM block containing public key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		log.Fatal(err)
	}
	var privateKeyRsa = privateKey.(*rsa.PrivateKey)

	//var hs = jwt.NewHS256([]byte(secret))
	var hs = jwt.NewRS256(jwt.RSAPublicKey(&privateKeyRsa.PublicKey), jwt.RSAPrivateKey(privateKeyRsa))
	now := time.Now()
	uncopiedOrg := os.Getenv("UNCOPIED_USERNAME")
	uncopiedDomainPrimary := os.Getenv("UNCOPIED_DOMAIN_PRIMARY")
	uncopiedDomainSecondary := os.Getenv("UNCOPIED_DOMAIN_SECONDARY")
	pl := CustomPayload{
		Payload: jwt.Payload{
		Issuer:         uncopiedOrg,
		Subject:        certificate.CertificateLabel,
		Audience:       jwt.Audience{uncopiedDomainPrimary, uncopiedDomainSecondary},
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

func IssueCertificate(db *gorm.DB, user dbmodel.User, certificateLabel string, IsDIY bool) dbmodel.Certificate {
	uncopiedUsername := os.Getenv("UNCOPIED_USERNAME")
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

	issuerToken := generateTokens(certificate, "Issuer")
	ownerToken := generateTokens(certificate, "Owner")
	primaryAssetVerifierToken := generateTokens(certificate, "PrimaryAssetVerifier")
	secondaryAssetVerifierToken := generateTokens(certificate, "SecondaryAssetVerifier")
	primaryOwnerVerifierToken := generateTokens(certificate, "PrimaryOwnerVerifier")
	secondaryOwnerVerifierToken := generateTokens(certificate, "SecondaryOwnerVerifier")
	primaryIssuerVerifierToken := generateTokens(certificate, "PrimaryIssuerVerifier")
	secondaryIssuerVerifierToken := generateTokens(certificate, "SecondaryIssuerVerifier")

	certificate.IssuerToken = issuerToken
	certificate.OwnerToken = ownerToken
	certificate.PrimaryAssetVerifierToken = primaryAssetVerifierToken
	certificate.SecondaryAssetVerifierToken = secondaryAssetVerifierToken
	certificate.PrimaryOwnerVerifierToken = primaryOwnerVerifierToken
	certificate.SecondaryOwnerVerifierToken = secondaryOwnerVerifierToken
	certificate.PrimaryIssuerVerifierToken = primaryIssuerVerifierToken
	certificate.SecondaryIssuerVerifierToken = secondaryIssuerVerifierToken
	db.Create(&certificate)
	return certificate
}


func order(c *gin.Context) {
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
	//{AssetTemplateID: 17, OrderUUID: "006c363d-c3d6-44d6-9cff-ba20a6269409"}
	type RequestBody struct {
		AssetTemplateID  int `json:"AssetTemplateID" binding:"required"`
		OrderUUID  string `json:"OrderUUID" binding:"required"`
	}
	var body RequestBody
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithStatus(400)
		return
	}
	var assetTemplate dbmodel.AssetTemplate
	if err := db.Preload("Source.Issuer").Preload("Source").Preload("Assets").Where("id = ?", body.AssetTemplateID).First(&assetTemplate).Error; err != nil {
		c.AbortWithStatus(404)
		return
	}
	if assetTemplate.Source.IssuerID != user.ID {
		fmt.Println("User doesnt own assetTemplate ", userName)
		c.AbortWithStatus(409)
		return
	}
	if assetTemplate.Source.Stamp == "" {
		fmt.Println("Asset is not stamped ")
		c.AbortWithStatus(409)
		return
	}
	if assetTemplate.ObjectUUID != body.OrderUUID {
		fmt.Println("Order and asset UUID mismatch ")
		c.AbortWithStatus(409)
		return
	}

	order := dbmodel.Order{
		OrderUUID:      body.OrderUUID,
		AssetTemplate:  assetTemplate,
		PaymentStatus:  "",
		DeliveryStatus: "",
	}

	// create the order for tracking
	db.Create(&order)

	c.JSON(200, order)
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

type CertPreview struct {
	DocPreviewURL string
	TaillyPreviewSVG string
}

func TallyStickPreview (first dbmodel.Asset ) string {
	// create a tally stick for checkout
	t := tallystick.Tallystick{
		CertificateLabel:                first.CertificateLabel,
		PrimaryLinkURL:                  PermanentCertificateURL(1),
		SecondaryLinkURL:                PermanentCertificateURL(1),
		IssuerTokenURL:                  PermanentCertificateTokenURL("certificate.IssuerToken.TokenHash"),
		OwnerTokenURL:                   PermanentCertificateTokenURL("certificate.OwnerToken.TokenHash"),
		PrimaryAssetVerifierTokenURL:    PermanentCertificateTokenURL("certificate.PrimaryAssetVerifier.TokenHash"),
		SecondaryAssetVerifierTokenURL:  PermanentCertificateTokenURL("certificate.SecondaryAssetVerifierToken.TokenHash"),
		PrimaryOwnerVerifierTokenURL:    PermanentCertificateTokenURL("certificate.PrimaryOwnerVerifierToken.TokenHash"),
		SecondaryOwnerVerifierTokenURL:  PermanentCertificateTokenURL("certificate.SecondaryOwnerVerifierToken.TokenHash"),
		PrimaryIssuerVerifierTokenURL:   PermanentCertificateTokenURL("certificate.PrimaryIssuerVerifierToken.TokenHash"),
		SecondaryIssuerVerifierTokenURL: PermanentCertificateTokenURL("certificate.SecondaryIssuerVerifierToken.TokenHash"),
		MailToContentLeft:               MailTo(),
		MailToContentRight:              MailTo(),
	}
	var buf bytes.Buffer
	err := tallystick.DrawSVG(&t, &buf)
	if err != nil {
		fmt.Println("Tallystick.DrawSVG failed ")
		return ""
	}
	return string(buf.Bytes())
}

func preview(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userName := c.MustGet("user")
	// check if user
	var user dbmodel.User
	if err := db.Where("user_name = ?", userName).First(&user).Error; err != nil {
		fmt.Println("User name not found ",userName)
		c.AbortWithStatus(409)
		return
	}
	id := c.Param("id")
	var asset dbmodel.AssetTemplate
	if err := db.Preload("Source").Preload("Assets").Where("id = ?", id).First(&asset).Error; err != nil {
		c.AbortWithStatus(404)
		return
	}
	if asset.Source.IssuerID != user.ID {
		fmt.Println("User doesnt own asset ",userName)
		c.AbortWithStatus(409)
		return
	}
	var first = asset.Assets[0]
	tallystickPreviewSVG := TallyStickPreview(first)
	baseURL := view.ServerBaseURL()
	preview := CertPreview{
		DocPreviewURL:    baseURL.ServerBaseURLExternal+"/c/preview/"+asset.ObjectUUID,
		TaillyPreviewSVG: tallystickPreviewSVG,
	}
	c.JSON(200, preview)
}

type Pricing struct {
	PriceDiy  int
	Price     int
	CcySymbol string
	Ccy       string
}

func checkout(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userName := c.MustGet("user")
	// check if user
	var user dbmodel.User
	if err := db.Where("user_name = ?", userName).First(&user).Error; err != nil {
		fmt.Println("User name not found ",userName)
		c.AbortWithStatus(409)
		return
	}
	id := c.Param("uuid")
	var order dbmodel.Order
	if err := db.Preload("AssetTemplate").Preload("AssetTemplate.Source").Preload("AssetTemplate.Assets").Where("order_uuid = ?", id).First(&order).Error; err != nil {
		c.AbortWithStatus(404)
		return
	}
	var asset = order.AssetTemplate
	if asset.Source.IssuerID != user.ID {
		fmt.Println("User doesnt own asset ",userName)
		c.AbortWithStatus(409)
		return
	}
	var first = asset.Assets[0]
	tallystickPreviewSVG := TallyStickPreview(first)
	baseURL := view.ServerBaseURL()
	preview := CertPreview {
		DocPreviewURL:    baseURL.ServerBaseURLExternal+"/c/preview/"+asset.ObjectUUID,
		TaillyPreviewSVG: tallystickPreviewSVG,
	}
	pricing := Pricing{
		PriceDiy:  5 + asset.EditionTotal * 1,
		Price:     25 + asset.EditionTotal * 5,
		CcySymbol: "€",
		Ccy:       "EUR",
	}
	type RequestResponse struct {
		Order dbmodel.Order
		Pricing Pricing
		CertPreview CertPreview
	}
	response :=RequestResponse{
		Order : order,
		Pricing :  pricing,
		CertPreview :preview,
	}
	c.JSON(200, response)
}


func process(c *gin.Context) {

	db := c.MustGet("db").(*gorm.DB)
	userName := c.MustGet("user")
	// check if user
	var user dbmodel.User
	if err := db.Where("user_name = ?", userName).First(&user).Error; err != nil {
		fmt.Println("User name not found ",userName)
		c.AbortWithStatus(409)
		return
	}

	type RequestBody struct {
		OrderUUID  string `json:"OrderUUID" binding:"required"`
		IsDIY bool `json:"IsDIY" binding:"required"`
		IsPaid bool `json:"IsPaid" binding:"required"`
		PaypalDetails string `json:"PaypalDetails" binding:"required"`
	}

	var body RequestBody
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithStatus(400)
		return
	}
	var order dbmodel.Order
	if err := db.Preload("AssetTemplate").Preload("AssetTemplate.Source").Preload("AssetTemplate.Assets").Where("order_uuid = ?", body.OrderUUID).First(&order).Error; err != nil {
		c.AbortWithStatus(404)
		return
	}

	if( body.IsPaid ) {
		// success
		order.PaymentStatus="PAID"
		order.DeliveryStatus="PENDING_DELIVERY"
		order.IsDIY=body.IsDIY
		order.PaypalDetails = body.PaypalDetails
	} else {
		// failure
		order.PaymentStatus="FAILED"
		// let's deliver for now and collect payment later
		order.DeliveryStatus="PENDING_DELIVERY"
		order.IsDIY=body.IsDIY
		order.PaypalDetails = body.PaypalDetails
	}
	db.Updates(&order)

	type RequestResponse struct {
		PaymentStatus string
		DeliveryStatus string
	}
	response :=RequestResponse{
		PaymentStatus:  order.PaymentStatus,
		DeliveryStatus: order.DeliveryStatus,
	}
	c.JSON(200, response)
}
