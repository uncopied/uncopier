package checkout

import (
	"bytes"
	"fmt"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/gin-gonic/gin"
	"github.com/uncopied/tallystick"
	"github.com/uncopied/uncopier/api/v1.0/cert"
	"github.com/uncopied/uncopier/database/dbmodel"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
)

const uncopied_root = "http://localhost:8081/"

func PermanentCertificateURL(token string) string {
	return uncopied_root + "/c/v/" + token
}

func PermanentCertificateTokenURL(token string) string {
	return uncopied_root + "/c/t/" + token
}

type Pricing struct {
	PriceDiy int
	Price int
	CcySymbol string
	Ccy string
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
		PrimaryLinkURL:                  PermanentCertificateURL("token"),
		SecondaryLinkURL:                 PermanentCertificateURL("token"),
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
		Price:      (20+asset.EditionTotal*5),
		CcySymbol: "€",
		Ccy:        "EUR",
	}
	c.HTML(http.StatusOK, "checkout.tmpl", gin.H{
		"asset":      first,
		"source":     asset.Source,
		"tallystick": tallystickHtml,
		"pricing": pricing,
	})
}

type TallystickDoc struct {
	CertificateLabel string
	OrderUUID string
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

	uuid := uuid.New().String()
	order := dbmodel.Order{
		OrderUUID : uuid,
		AssetTemplate:   asset,
		PaymentStatus:   "FREE",
		DeliveryStatus: "NEW",
	}

	// create the order for tracking
	db.Create(&order)

	// PDF document with all the certificates
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		c.AbortWithStatus(500)
		return
	}
	// Set global options
	pdfg.Dpi.Set(300)
	pdfg.Orientation.Set(wkhtmltopdf.OrientationPortrait)
	pdfg.Grayscale.Set(false)
	tallystickDocs := make([]TallystickDoc,0)

	filePath := "./public/doc/"+uuid+"/"
	err = os.MkdirAll(filePath,os.ModePerm)
	if err != nil {
		log.Fatal(err)
		return
	}

	// checkout the first
	for _, elem := range asset.Assets {
		token := ""+strconv.Itoa(int(elem.ID))

		certificate := cert.IssueCertificate(db, user, elem.CertificateLabel, order.IsDIY)
		certificateIssuance := dbmodel.CertificateIssuance{
			Asset:         elem,
			Order:         order,
			Certificate:   certificate,
		}
		db.Create(&certificateIssuance)
		// create a tally stick for checkout
		t := tallystick.Tallystick{
			CertificateLabel:                elem.CertificateLabel,
			PrimaryLinkURL:                  PermanentCertificateURL(token),
			SecondaryLinkURL:                 PermanentCertificateURL(token), //TODO change this
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
		err := tallystick.DrawPDF(&t, &buf)
		if err != nil {
			fmt.Println("Tallystick.DrawPDF failed ")
			c.AbortWithStatus(500)
			return
		}
		tallystickDoc := TallystickDoc{
			OrderUUID: uuid,
			CertificateLabel: elem.CertificateLabel,
			Filename:         "tallystick_"+uuid+"_"+token+".pdf",
		}
		out, err := os.Create(filePath+tallystickDoc.Filename)
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
		tallystickDocs = append(tallystickDocs,tallystickDoc)
		// make 4 copies
		for i := 0; i < 4; i++ {
			// issuer copy
			page := wkhtmltopdf.NewPage("http://localhost:8081/c/v/"+token)
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
	certificateDoc := "certificate_"+uuid+".pdf"
	err = pdfg.WriteFile(filePath+certificateDoc)
	if err != nil {
		log.Fatal(err)
	}
	c.HTML(http.StatusOK, "success.tmpl", gin.H{
		"order":      order,
		"certificateDoc": certificateDoc,
		"tallystickDocs": tallystickDocs,
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
		Price:      (20+asset.EditionTotal*5),
		CcySymbol: "€",
		Ccy:        "EUR",
	}
	c.HTML(http.StatusOK, "cancel.tmpl", gin.H{
		"asset":      first,
		"source":     asset.Source,
		"pricing": pricing,
	})
}
