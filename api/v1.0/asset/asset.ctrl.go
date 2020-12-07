package asset

import (
	"bytes"
	"fmt"
	"github.com/uncopied/uncopier/database/dbmodel"
	"strconv"
	"text/template"
	"gorm.io/gorm"
	"github.com/gin-gonic/gin"
	"time"
)

type AssetBundle struct {
	Template dbmodel.AssetTemplate
	Assets []dbmodel.Asset
	Status string
	ErrorMessage []string
}

func create(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	type RequestBody struct {
		SourceID int `json:"source_id" binding:"required"`
		Name string `json:"name" binding:"required"`
		CertificateLabel string  `json:"certificate_label" binding:"required"`
		Metadata string  `json:"metadata"`
		ExternalMetadataURL string  `json:"external_metadata_url"`
		Note string `json:"note"`
		// template parameters
		ExternalAssetId int `json:"external_asset_id"`
		EditionTotal int `json:"edition_total"`
	}

	var body RequestBody
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithStatus(400)
		return
	}

	userName := c.MustGet("user")
	// check if user
	var user dbmodel.User
	if err := db.Where("user_name = ?", userName).First(&user).Error; err != nil {
		fmt.Println("User name not found ",userName)
		c.AbortWithStatus(409)
		return
	}

	// check source, status and ownership
	var assetSrc dbmodel.DigitalAssetSrc
	if err := db.Preload("Issuer").Where("id = ?", body.SourceID).First(&assetSrc).Error; err != nil {
		c.AbortWithStatus(404)
		return
	}

	if assetSrc.IssuerID !=user.ID {
		// that's not the user's template
		c.AbortWithStatus(409)
		return
	}
	if assetSrc.Stamp == "" {
		// that template isn't stamped
		c.AbortWithStatus(409)
		return
	}

	assetTemplate := dbmodel.AssetTemplate{
		Metadata:            body.Metadata,
		ExternalMetadataURL: body.ExternalMetadataURL,
		ExternalAssetId:     body.ExternalAssetId,
		EditionTotal:        body.EditionTotal,
		Name:                body.Name,
		CertificateLabel:    body.CertificateLabel,
		Note:                body.Note,
	}
	// if the work is not an edition then EditionTotal=1
	if assetTemplate.EditionTotal == 0 {
		assetTemplate.EditionTotal = 1
	}
	bundle,err := evaluate(&assetTemplate)
	if err!=nil {
		c.AbortWithStatus(500)
		return
	}
	if bundle.Status == "CREATE" && len(bundle.ErrorMessage) == 0 {
		// the bundle is valid, so create in DB
		db.Create(&assetTemplate)
		for _, elem := range bundle.Assets {
			db.Create(&elem)
		}
	}
	c.JSON(200, &bundle)
}

type TemplateParams struct {
	ExternalSourceID string
	IssuerUserName string
	IssuerEthereumAddress string
	IssuerName string
	AuthorName string
	ThumbnailURL string
	EditionTotal int
	EditionNumber int
	ExternalAssetId int
	AssetName string
	CertificateLabel string
	ExternalMetadataURL string
	CurrentYear string
	Note string
}

func execute(templateParams *TemplateParams, templateString string) string {
	tMetadata := template.Must(template.New("field").Parse(templateString))
	buf := new(bytes.Buffer)
	tMetadata.Execute(buf, templateParams)
	return buf.String()
}

const maxAssetNameLength = 32
const maxCertificateLabelLength = 128
const maxNoteLength = 1000

func evaluate(assetTemplate *dbmodel.AssetTemplate) (AssetBundle, error) {
  	assets := make([]dbmodel.Asset, assetTemplate.EditionTotal)
	bundle := AssetBundle{
		Template:     *assetTemplate,
		Assets:       assets,
		Status:       "CREATE",
		ErrorMessage: make([]string,0),
	}
	currentTime := time.Now()
	currentYear := strconv.Itoa(currentTime.Year())
	for i := 0; i< assetTemplate.EditionTotal ;i++ {
		externalAssetId := 0
		if assetTemplate.ExternalAssetId > 0 {
			externalAssetId = assetTemplate.ExternalAssetId+i
		}
		templateParams := TemplateParams{
			CurrentYear: currentYear,
			ThumbnailURL: dbmodel.ThumbnailURL(assetTemplate.Source),
			AuthorName: assetTemplate.Source.AuthorName,
			IssuerName: assetTemplate.Source.Issuer.DisplayName,
			IssuerEthereumAddress : assetTemplate.Source.Issuer.EthereumAddress,
			ExternalSourceID : assetTemplate.Source.ExternalSourceID,
			IssuerUserName: assetTemplate.Source.Issuer.UserName,

			EditionTotal:    assetTemplate.EditionTotal,
			ExternalAssetId: externalAssetId,
			EditionNumber:   i,
		}
		// evaluate asset name first
		assetName := execute(&templateParams, assetTemplate.Name)
		if len(assetName) > maxAssetNameLength {
			bundle.ErrorMessage = append(bundle.ErrorMessage, "asset name length too long : "+assetName)
		}
		templateParams.AssetName = assetName
		// then evaluate certificate label
		certificateLabel := execute(&templateParams, assetTemplate.CertificateLabel)
		if len(certificateLabel) > maxCertificateLabelLength {
			bundle.ErrorMessage = append(bundle.ErrorMessage, "asset label length too long : "+certificateLabel)
		}
		templateParams.CertificateLabel = certificateLabel
		externalMetadataURL:= execute(&templateParams, assetTemplate.ExternalMetadataURL)
		// TODO: call it and get it
		templateParams.ExternalMetadataURL = externalMetadataURL
		note:=execute(&templateParams, assetTemplate.Note)
		if len(note) > maxNoteLength {
			bundle.ErrorMessage = append(bundle.ErrorMessage, "asset note too long : "+note)
		}
		metadata:=execute(&templateParams, assetTemplate.Metadata)
		// TODO : json schema validation an comparison with value at externalMetadataURL
		asset := dbmodel.Asset{
			EditionTotal:        assetTemplate.EditionTotal,
			EditionNumber:       i,
			ExternalAssetId:     externalAssetId,
			Name:                assetName, // evaluated first
			CertificateLabel: certificateLabel, // evaluated second
			ExternalMetadataURL: externalMetadataURL,
			Note:                note,
			Metadata:            metadata,
			Template:            *assetTemplate,
		}
		assets = append(assets, asset)
	}
	return bundle, nil
}

func list(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userName := c.MustGet("user")
	// check if user
	var user dbmodel.User
	if err := db.Where("user_name = ?", userName).First(&user).Error; err != nil {
		fmt.Println("User name not found ",userName)
		c.AbortWithStatus(409)
		return
	}
	var assetSrc []dbmodel.DigitalAssetSrc
	if err := db.Where("issuer_id = ?", user.ID).Order("id asc").Find(&assetSrc).Error; err != nil {
		c.AbortWithStatus(500)
		return
	}
	c.JSON(200, assetSrc)
}

func read(c *gin.Context) {
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
	var assetSrc dbmodel.DigitalAssetSrc

	// auto preloads the related dbmodel
	// http://gorm.io/docs/preload.html#Auto-Preloading
	if err := db.Set("gorm:auto_preload", true).Where("id = ?", id).First(&assetSrc).Error; err != nil {
		c.AbortWithStatus(404)
		return
	}
	c.JSON(200, assetSrc)
}

