package cert

import (
	"archive/zip"
	"bytes"
	"crypto/md5"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/gbrlsnchs/jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/uncopied/chirograph"
	"github.com/uncopied/uncopier/blockchain"
	"github.com/uncopied/uncopier/certificates/view"
	"github.com/uncopied/uncopier/database/dbmodel"
	"gorm.io/gorm"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
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

func chirographPreview (first dbmodel.Asset ) string {
	// create a tally stick for checkout
	t := chirograph.Chirograph{
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
		TopHelper: TopHelper(),
		BottomHelper: BottomHelper(),
		LeftHelper:LeftHelper(),
		RightHelper:RightHelper(),
	}
	var buf bytes.Buffer
	err := chirograph.DrawSVG(&t, &buf)
	if err != nil {
		fmt.Println("chirograph.DrawSVG failed ")
		return ""
	}
	svg := string(buf.Bytes())
	// the svg should not have a fixed width/height
	svg = strings.Replace(svg,"width=\"297mm\" height=\"210mm\"","",1)
	return svg
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
	chirographPreviewSVG := chirographPreview(first)
	baseURL := view.ServerBaseURL()
	preview := CertPreview{
		DocPreviewURL:    baseURL.ServerBaseURLExternal+"/c/preview/"+asset.ObjectUUID,
		TaillyPreviewSVG: chirographPreviewSVG,
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
	chirographPreviewSVG := chirographPreview(first)
	baseURL := view.ServerBaseURL()
	preview := CertPreview {
		DocPreviewURL:    baseURL.ServerBaseURLExternal+"/c/preview/"+asset.ObjectUUID,
		TaillyPreviewSVG: chirographPreviewSVG,
	}
	pricing := Pricing{
		PriceDiy:  asset.EditionTotal * 1, // for now
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

	// in any case, mark the order for delivery and prepare order
	if( body.IsPaid ) {
		// success
		order.PaymentStatus="PAID"
		order.DeliveryStatus="PENDING_DELIVERY"
		order.IsDIY=body.IsDIY
		order.PaypalDetails = body.PaypalDetails
	} else {
		// failure
		order.PaymentStatus="FAILED"
		// let's prepare for now and collect payment later
		order.DeliveryStatus="PENDING_DELIVERY"
		order.IsDIY=body.IsDIY
		order.PaypalDetails = body.PaypalDetails
	}
	db.Updates(&order)

	// in any case, prepare the order
	go prepare(db, order, user)

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

type chirographDoc struct {
	CertificateLabel string
	OrderUUID        string
	Filename         string
}

func prepareFail(db *gorm.DB, order dbmodel.Order, user dbmodel.User, errorMessage string) {
	order.ProductionStatus = "FAIL"
	order.ProductionMessage = errorMessage
	db.Updates(&order)
}

func prepareStep(db *gorm.DB, order dbmodel.Order, user dbmodel.User, productionStatus string, productionMessage string) {
	fmt.Println("delivery step "+productionStatus+" "+productionMessage)
	order.ProductionStatus = productionStatus
	order.ProductionMessage = productionMessage
	db.Updates(&order)
}

func prepare(db *gorm.DB, order dbmodel.Order, user dbmodel.User) {
	// do stuff
	uuid := order.OrderUUID
	assetTemplate := order.AssetTemplate
	mailTo := MailTo()
	// PDF document with all the certificates
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	// NB/ occasionally has https://github.com/wkhtmltopdf/wkhtmltopdf/issues/3933
	// I should try : Issue got resolved after adding --javascript-delay 1000
	if err != nil {
		fmt.Println("wkhtmltopdf.NewPDFGenerator() failed ")
		prepareFail(db, order, user, "wkhtmltopdf.NewPDFGenerator() failed ")
		return
	}
	prepareStep(db, order, user, "PDF_GENERATOR",uuid)
	// Set global options
	pdfg.Dpi.Set(300)
	pdfg.Orientation.Set(wkhtmltopdf.OrientationPortrait)
	pdfg.Grayscale.Set(false)


	chirographDocs := make([]chirographDoc, 0)
	filePath := "./public/doc/" + uuid + "/"
	err = os.MkdirAll(filePath, os.ModePerm)
	if err != nil {
		fmt.Println("os.MkdirAll() error "+err.Error())
		prepareFail(db, order, user, "os.MkdirAll() error "+err.Error())
		return
	}
	serverBaseURL := view.ServerBaseURL()
	//# IPFSNode = "localhost:5001"
	//LOCAL_IPFS_NODE_HOST=127.0.0.1
	//LOCAL_IPFS_NODE_PORT=5001

	// checkout the first
	for _, asset := range assetTemplate.Assets {
		assetViewURLExternal :=serverBaseURL.ServerBaseURLExternal+"/c/v/" + strconv.Itoa(int(asset.ID))
		assetViewURLInternal :=serverBaseURL.ServerBaseURLInternal+"/c/v/" + strconv.Itoa(int(asset.ID))
		prepareStep(db, order, user, "PROCESS_ASSET", assetViewURLExternal)

		// create hash with metadata
		prepareStep(db, order, user, "READ_METADATA", assetViewURLExternal)
		metadata := asset.Metadata
		if metadata == "" {
			// in case we don't have metadata, just use our asset object
			fallback, err := json.Marshal(asset)
			if err != nil {
				fmt.Println("json.Marshal(asset) "+err.Error())
				prepareFail(db, order, user, "json.Marshal(asset) "+err.Error())
				return
			}
			metadata=string(fallback)
		}

		transactionId, err := blockchain.AlgorandCreateNFT(&asset, assetViewURLExternal)
		if err != nil {
			fmt.Println("blockchain.AlgorandCreateNFT "+err.Error())
			prepareFail(db, order, user, "blockchain.AlgorandCreateNFT "+err.Error())
			return
		}
		prepareStep(db, order, user, "ISSUE_CERTIFICATE", assetViewURLExternal)
		certificate := IssueCertificate(db, user, asset.CertificateLabel, order.IsDIY)
		certificateIssuance := dbmodel.CertificateIssuance{
			Asset:                 asset,
			Order:                 order,
			Certificate:           certificate,
			AlgorandTransactionID: transactionId,
		}
		db.Create(&certificateIssuance)

		prepareStep(db, order, user, "DRAW_chirograph", assetViewURLExternal)
		// create a chirograph for checkout
		t := chirograph.Chirograph{
			CertificateLabel:                asset.CertificateLabel,
			PrimaryLinkURL:                  PermanentCertificateURL(certificateIssuance.ID),
			// TODO change this
			SecondaryLinkURL:                "algorand://tx/" + transactionId,
			IssuerTokenURL:                  PermanentCertificateTokenURL(certificate.IssuerToken.TokenHash),
			OwnerTokenURL:                   PermanentCertificateTokenURL(certificate.OwnerToken.TokenHash),
			PrimaryAssetVerifierTokenURL:    PermanentCertificateTokenURL(certificate.PrimaryAssetVerifierToken.TokenHash),
			SecondaryAssetVerifierTokenURL:  PermanentCertificateTokenURL(certificate.SecondaryAssetVerifierToken.TokenHash),
			PrimaryOwnerVerifierTokenURL:    PermanentCertificateTokenURL(certificate.PrimaryOwnerVerifierToken.TokenHash),
			SecondaryOwnerVerifierTokenURL:  PermanentCertificateTokenURL(certificate.SecondaryOwnerVerifierToken.TokenHash),
			PrimaryIssuerVerifierTokenURL:   PermanentCertificateTokenURL(certificate.PrimaryIssuerVerifierToken.TokenHash),
			SecondaryIssuerVerifierTokenURL: PermanentCertificateTokenURL(certificate.SecondaryIssuerVerifierToken.TokenHash),
			MailToContentLeft:               mailTo,
			MailToContentRight:              mailTo,
			TopHelper: TopHelper(),
			BottomHelper: BottomHelper(),
			LeftHelper:LeftHelper(),
			RightHelper:RightHelper(),
		}
		var buf bytes.Buffer
		err = chirograph.DrawPDF(&t, &buf)
		if err != nil {
			fmt.Println("chirograph.DrawPDF(&t, &buf) "+err.Error())
			prepareFail(db, order, user, "chirograph.DrawPDF(&t, &buf)"+err.Error())
			return
		}
		chirographDoc := chirographDoc{
			OrderUUID:        uuid,
			CertificateLabel: asset.CertificateLabel,
			Filename:         "chirograph_" + uuid + "_" + strconv.Itoa(int(asset.ID)) + ".pdf",
		}
		prepareStep(db, order, user, "SAVE_chirograph", assetViewURLExternal)
		out, err := os.Create(filePath + chirographDoc.Filename)
		if err != nil {
			fmt.Println("os.Create(filePath + chirographDoc.Filename) "+err.Error())
			prepareFail(db, order, user, "os.Create(filePath + chirographDoc.Filename) "+err.Error())
			return
		}
		defer out.Close()
		// Write the body to file
		_, err = out.Write(buf.Bytes())
		if err != nil {
			fmt.Println("out.Write(buf.Bytes()) "+err.Error())
			prepareFail(db, order, user, "out.Write(buf.Bytes()) "+err.Error())
			return
		}
		chirographDocs = append(chirographDocs, chirographDoc)
		prepareStep(db, order, user, "APPEND_CERTIFICATE", assetViewURLExternal)
		// make 4 copies
		for i := 0; i < 4; i++ {
			// issuer copy
			page := wkhtmltopdf.NewPage(assetViewURLInternal)
			// https://github.com/wkhtmltopdf/wkhtmltopdf/issues/3933
			// Issue got resolved after adding --javascript-delay 1000
			page.JavascriptDelay.Set(1000)

			// Set options for this page
			page.FooterRight.Set("[page]")
			page.FooterFontSize.Set(10)
			page.Zoom.Set(0.95)
			// Add to document
			pdfg.AddPage(page)
		}
	}
	prepareStep(db, order, user, "PDF_CREATE",uuid)
	// Create PDF document in internal buffer

	err = pdfg.Create()
	if err != nil {
		fmt.Println("pdfg.Create() "+err.Error())
		prepareFail(db, order, user, "pdfg.Create() "+err.Error())
		return
	}
	// Write buffer contents to file on disk
	prepareStep(db, order, user, "PDF_WRITE",uuid)
	certificateDoc := "certificate_" + uuid + ".pdf"
	err = pdfg.WriteFile(filePath + certificateDoc)
	if err != nil {
		fmt.Println("pdfg.WriteFile(filePath + certificateDoc) "+err.Error())
		prepareFail(db, order, user, "pdfg.WriteFile(filePath + certificateDoc) "+err.Error())
		return
	}
	prepareStep(db, order, user, "ZIP_BUNDLE",uuid)
	zipBundle := "uncopied_"+uuid+".zip"
	out, err := os.Create(filePath + zipBundle)
	if err != nil {
		fmt.Println("os.Create(filePath + zipBundle) "+err.Error())
		prepareFail(db, order, user, "os.Create(filePath + zipBundle) "+err.Error())
		return
	}
	defer out.Close()
	// Create a new zip archive.
	w := zip.NewWriter(out)
	// Add files to the archive.
	{
		prepareStep(db, order, user, "ZIP_BUNDLE_ADD_CERT",uuid)
		f, err := w.Create(certificateDoc)
		if err != nil {
			fmt.Println("w.Create(certificateDoc) "+err.Error())
			prepareFail(db, order, user, "w.Create(certificateDoc) "+err.Error())
			return
		}
		b, err := ioutil.ReadFile(filePath + certificateDoc) // just pass the file name
		if err != nil {
			fmt.Println("ioutil.ReadFile(filePath + certificateDoc) "+err.Error())
			prepareFail(db, order, user, "ioutil.ReadFile(filePath + certificateDoc) "+err.Error())
			return
		}
		_, err = f.Write(b)
		if err != nil {
			fmt.Println("f.Write(b) "+err.Error())
			prepareFail(db, order, user, "f.Write(b) "+err.Error())
			return
		}
	}
	for _, file := range chirographDocs {
		prepareStep(db, order, user, "ZIP_BUNDLE_ADD_CHIROGRAPH",uuid+"/"+file.Filename)
		f, err := w.Create(file.Filename)
		if err != nil {
			fmt.Println("for _, file := range chirographDocs; w.Create(file.Filename) "+err.Error())
			prepareFail(db, order, user, "for _, file := range chirographDocs; w.Create(file.Filename)"+err.Error())
			return
		}
		b, err := ioutil.ReadFile(filePath +file.Filename) // just pass the file name
		if err != nil {
			fmt.Println("for _, file := range chirographDocs; ioutil.ReadFile(filePath +file.Filename) "+err.Error())
			prepareFail(db, order, user, "for _, file := range chirographDocs; ioutil.ReadFile(filePath +file.Filename) "+err.Error())
			return
		}
		_, err = f.Write(b)
		if err != nil {
			fmt.Println("for _, file := range chirographDocs; _, err = f.Write(b) "+err.Error())
			prepareFail(db, order, user, "for _, file := range chirographDocs; _, err = f.Write(b) "+err.Error())
			return
		}
	}
	prepareStep(db, order, user, "ZIP_BUNDLE_CLOSE",uuid)
	// Make sure to check the error on Close.
	err = w.Close()
	if err != nil {
		fmt.Println("err = w.Close() "+err.Error())
		prepareFail(db, order, user, "err = w.Close() "+err.Error())
		return
	}
	prepareStep(db, order, user, "READY_TO_DELIVER", "")
	order.ZipBundle = zipBundle
	db.Updates(&order)
}


