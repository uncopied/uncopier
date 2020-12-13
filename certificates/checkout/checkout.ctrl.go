package checkout

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/uncopied/tallystick"
	"github.com/uncopied/uncopier/api/v1.0/cert"
	"github.com/uncopied/uncopier/blockchain"
	"github.com/uncopied/uncopier/database/dbmodel"
	"gorm.io/gorm"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

const uncopied_root = "http://localhost:8081"
const IPFSNode = "localhost:5001"

func PermanentCertificateURL(certificateIssuanceID uint) string {
	return uncopied_root + "/c/y/" + strconv.Itoa(int(certificateIssuanceID))
}

func PermanentCertificateTokenURL(token string) string {
	return uncopied_root + "/c/t/" + token
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
		fmt.Println("User name not found ", userName)
		c.AbortWithStatus(409)
		return
	}

	id := c.Param("id")
	var asset dbmodel.AssetTemplate
	if err := db.Preload("Source.Issuer").Preload("Source").Preload("Assets").Where("id = ?", id).First(&asset).Error; err != nil {
		c.AbortWithStatus(404)
		return
	}

	if asset.Source.IssuerID != user.ID {
		fmt.Println("User doesnt own asset ", userName)
		c.AbortWithStatus(409)
		return
	}

	if asset.Source.Stamp == "" {
		fmt.Println("Asset is not stamped ")
		c.AbortWithStatus(409)
		return
	}

	// checkout the first
	var first = asset.Assets[0]

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
		MailToContentLeft:               cert.MailTo(),
		MailToContentRight:              cert.MailTo(),
	}
	var buf bytes.Buffer
	err := tallystick.DrawSVG(&t, &buf)
	if err != nil {
		fmt.Println("Tallystick.DrawSVG failed ")
		c.AbortWithStatus(500)
		return
	}
	tallystickTxt := string(buf.Bytes())
	tallystickHtml := template.HTML(tallystickTxt)
	//prev := CertPreview {
	//
	//	certificateSVG: string(buf.Bytes()),
	//}
	pricing := Pricing{
		PriceDiy:  asset.EditionTotal,
		Price:     20 + asset.EditionTotal*5,
		CcySymbol: "€",
		Ccy:       "EUR",
	}
	c.HTML(http.StatusOK, "checkout.tmpl", gin.H{
		"asset":      first,
		"source":     asset.Source,
		"tallystick": tallystickHtml,
		"pricing":    pricing,
	})
}

type TallystickDoc struct {
	CertificateLabel string
	OrderUUID        string
	Filename         string
}

func success(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	// renderer:=c.MustGet("render").(*render.HTMLRender)

	userName := c.MustGet("user")
	// check if user
	var user dbmodel.User
	if err := db.Where("user_name = ?", userName).First(&user).Error; err != nil {
		fmt.Println("User name not found ", userName)
		c.AbortWithStatus(409)
		return
	}
	id := c.Param("id")
	var assetTemplate dbmodel.AssetTemplate
	if err := db.Preload("Source.Issuer").Preload("Source").Preload("Assets").Where("id = ?", id).First(&assetTemplate).Error; err != nil {
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

	uuid := uuid.New().String()
	order := dbmodel.Order{
		OrderUUID:      uuid,
		AssetTemplate:  assetTemplate,
		PaymentStatus:  "FREE",
		DeliveryStatus: "NEW",
	}

	// create the order for tracking
	db.Create(&order)

	// PDF document with all the certificates
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		fmt.Println("wkhtmltopdf.NewPDFGenerator() failed ")
		c.AbortWithStatus(500)
		return
	}
	// Set global options
	pdfg.Dpi.Set(300)
	pdfg.Orientation.Set(wkhtmltopdf.OrientationPortrait)
	pdfg.Grayscale.Set(false)
	tallystickDocs := make([]TallystickDoc, 0)
	filePath := "./public/doc/" + uuid + "/"
	err = os.MkdirAll(filePath, os.ModePerm)
	if err != nil {
		log.Fatal(err)
		return
	}
	// URL for preview
	tls := os.Getenv("SERVER_TLS")
	serverHost := os.Getenv("SERVER_HOST")
	serverBaseURLExternal := ""
	serverBaseURLInternal := ""
	if tls=="http" {
		// needs redirecting 80 port to 8081
		httpPort:=os.Getenv("SERVER_HTTP_PORT")
		serverBaseURLInternal="http://"+serverHost+":"+httpPort
		serverBaseURLExternal=serverBaseURLInternal
	} else if tls=="https" {
		// needs redirecting 443 port to 8443
		httpsPort:=os.Getenv("SERVER_HTTPS_PORT")
		serverBaseURLInternal="https://"+serverHost+":"+httpsPort
		serverBaseURLExternal="https://"+serverHost
	} else if tls=="autocert" {
		// needs root priviledge to run on 443
		serverBaseURLInternal="https://"+serverHost
		serverBaseURLExternal=serverBaseURLInternal
	}

	// checkout the first
	for _, asset := range assetTemplate.Assets {
		assetViewURLExternal :=serverBaseURLExternal+"/c/v/" + strconv.Itoa(int(asset.ID))
		assetViewURLInternal :=serverBaseURLInternal+"/c/v/" + strconv.Itoa(int(asset.ID))
		// create hash with metadata
		metadata, err := json.Marshal(asset)
		if err != nil {
			log.Fatal(err)
			return
		}
		sh := shell.NewShell(IPFSNode)
		metadataHash, err := sh.Add(bytes.NewReader(metadata))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s", err)
			os.Exit(1)
		}
		transactionId, err := blockchain.AlgorandCreateNFT(&asset, assetViewURLExternal, metadataHash)
		if err != nil {
			fmt.Println("blockchain.AlgorandCreateNFT(asset) failed ")
			c.AbortWithStatus(500)
			return
		}
		certificate := cert.IssueCertificate(db, user, asset.CertificateLabel, order.IsDIY)
		certificateIssuance := dbmodel.CertificateIssuance{
			Asset:                 asset,
			Order:                 order,
			Certificate:           certificate,
			AlgorandTransactionID: transactionId,
			MetadataHash : metadataHash,
		}
		db.Create(&certificateIssuance)
		// create a tally stick for checkout
		t := tallystick.Tallystick{
			CertificateLabel:                asset.CertificateLabel,
			PrimaryLinkURL:                  PermanentCertificateURL(certificateIssuance.ID),
			SecondaryLinkURL:                "algorand://tx/" + transactionId,
			IssuerTokenURL:                  PermanentCertificateTokenURL(certificate.IssuerToken.TokenHash),
			OwnerTokenURL:                   PermanentCertificateTokenURL(certificate.OwnerToken.TokenHash),
			PrimaryAssetVerifierTokenURL:    PermanentCertificateTokenURL(certificate.PrimaryAssetVerifierToken.TokenHash),
			SecondaryAssetVerifierTokenURL:  PermanentCertificateTokenURL(certificate.SecondaryAssetVerifierToken.TokenHash),
			PrimaryOwnerVerifierTokenURL:    PermanentCertificateTokenURL(certificate.PrimaryOwnerVerifierToken.TokenHash),
			SecondaryOwnerVerifierTokenURL:  PermanentCertificateTokenURL(certificate.SecondaryOwnerVerifierToken.TokenHash),
			PrimaryIssuerVerifierTokenURL:   PermanentCertificateTokenURL(certificate.PrimaryIssuerVerifierToken.TokenHash),
			SecondaryIssuerVerifierTokenURL: PermanentCertificateTokenURL(certificate.SecondaryIssuerVerifierToken.TokenHash),
			MailToContentLeft:               cert.MailTo(),
			MailToContentRight:              cert.MailTo(),
		}
		var buf bytes.Buffer
		err = tallystick.DrawPDF(&t, &buf)
		if err != nil {
			fmt.Println("Tallystick.DrawPDF failed ")
			c.AbortWithStatus(500)
			return
		}
		tallystickDoc := TallystickDoc{
			OrderUUID:        uuid,
			CertificateLabel: asset.CertificateLabel,
			Filename:         "tallystick_" + uuid + "_" + strconv.Itoa(int(asset.ID)) + ".pdf",
		}
		out, err := os.Create(filePath + tallystickDoc.Filename)
		if err != nil {
			log.Fatal(err)
			return
		}
		defer out.Close()
		// Write the body to file
		_, err = out.Write(buf.Bytes())
		if err != nil {
			log.Fatal(err)
			return
		}
		tallystickDocs = append(tallystickDocs, tallystickDoc)
		// make 4 copies
		for i := 0; i < 4; i++ {
			// issuer copy
			page := wkhtmltopdf.NewPage(assetViewURLInternal)
			// Set options for this page
			page.FooterRight.Set("[page]")
			page.FooterFontSize.Set(10)
			page.Zoom.Set(0.95)
			// Add to document
			pdfg.AddPage(page)
		}
	}
	// Create PDF document in internal buffer
	err = pdfg.Create()
	if err != nil {
		log.Fatal(err)
	}
	// Write buffer contents to file on disk
	certificateDoc := "certificate_" + uuid + ".pdf"
	err = pdfg.WriteFile(filePath + certificateDoc)
	if err != nil {
		log.Fatal(err)
	}
	zipBundle := "uncopied_"+uuid+".zip"
	out, err := os.Create(filePath + zipBundle)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer out.Close()
	// Create a new zip archive.
	w := zip.NewWriter(out)
	// Add files to the archive.
	{
		f, err := w.Create(certificateDoc)
		if err != nil {
			log.Fatal(err)
		}
		b, err := ioutil.ReadFile(filePath + certificateDoc) // just pass the file name
		if err != nil {
			fmt.Print(err)
		}
		_, err = f.Write(b)
		if err != nil {
			log.Fatal(err)
		}
	}
	for _, file := range tallystickDocs {
		f, err := w.Create(file.Filename)
		if err != nil {
			log.Fatal(err)
		}
		b, err := ioutil.ReadFile(filePath +file.Filename) // just pass the file name
		if err != nil {
			fmt.Print(err)
		}
		_, err = f.Write(b)
		if err != nil {
			log.Fatal(err)
		}
	}
	// Make sure to check the error on Close.
	err = w.Close()
	if err != nil {
		log.Fatal(err)
	}
	order.ZipBundle = zipBundle
	if order.IsDIY {
		order.DeliveryStatus="DELIVERED"
	} else {
		order.DeliveryStatus="PENDING_DELIVERY"
	}
	db.Updates(&order)
	c.HTML(http.StatusOK, "success.tmpl", gin.H{
		"order":          order,
		"certificateDoc": certificateDoc,
		"tallystickDocs": tallystickDocs,
		"zipBundle" : zipBundle,
	})
}

func cancel(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userName := c.MustGet("user")
	// check if user
	var user dbmodel.User
	if err := db.Where("user_name = ?", userName).First(&user).Error; err != nil {
		fmt.Println("User name not found ", userName)
		c.AbortWithStatus(409)
		return
	}
	id := c.Param("id")
	var asset dbmodel.AssetTemplate
	if err := db.Preload("Source.Issuer").Preload("Source").Preload("Assets").Where("id = ?", id).First(&asset).Error; err != nil {
		c.AbortWithStatus(404)
		return
	}
	if asset.Source.IssuerID != user.ID {
		fmt.Println("User doesnt own asset ", userName)
		c.AbortWithStatus(409)
		return
	}
	if asset.Source.Stamp == "" {
		fmt.Println("Asset is not stamped ")
		c.AbortWithStatus(409)
		return
	}
	// checkout the first
	var first = asset.Assets[0]
	pricing := Pricing{
		PriceDiy:  asset.EditionTotal,
		Price:     20 + asset.EditionTotal*5,
		CcySymbol: "€",
		Ccy:       "EUR",
	}
	c.HTML(http.StatusOK, "cancel.tmpl", gin.H{
		"asset":   first,
		"source":  asset.Source,
		"pricing": pricing,
	})
}
