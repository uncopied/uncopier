package asset

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/algorand/go-algorand-sdk/encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/uncopied/uncopier/database/dbmodel"
	"gorm.io/gorm"
	"os"
	"strconv"
	"text/template"
	"time"
)

type AssetBundle struct {
	Template *dbmodel.AssetTemplate
//	Params []TemplateParams
	Status string
	ErrorMessage []string
}

type AssetError struct{
	Message    string
}

//create asset data = {"name":"Allégorie Florale",
//"certificate_label":"{{.AssetName}} by {{.IssuerName}}, {{.CurrentYear}} ( {{.EditionNumber}} / {{.EditionTotal}} )",
//"edition_total":"3",
//"source_id":14}
func create(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	type RequestBody struct {
		SourceID int `json:"source_id" binding:"required"`
		Name string `json:"name" binding:"required"`
		CertificateLabel string  `json:"certificate_label" binding:"required"`
		AssetLabel string  `json:"asset_label" binding:"required"`
		AssetProperties map[string]string `json:"asset_properties" binding:"required"`
		Metadata string  `json:"metadata"`
		ExternalMetadataURL string  `json:"external_metadata_url"`
		Note string `json:"note"`
		// template parameters
		ExternalAssetId int `json:"external_asset_id"`
		EditionTotal int `json:"edition_total"`
	}

	var body RequestBody
	if err := c.BindJSON(&body); err != nil {
		fmt.Println("Failed to bind body")
		assetError := AssetError{
			Message: "Failed to bind "+err.Error(),
		}
		c.AbortWithStatusJSON(400, assetError)
		return
	}

	userName := c.MustGet("user")
	// check if user
	var user dbmodel.User
	if err := db.Where("user_name = ?", userName).First(&user).Error; err != nil {
		fmt.Println("User name not found ",userName)
		assetError := AssetError{
			Message: "User name not found",
		}
		c.AbortWithStatusJSON(409, assetError)
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
	propsJson := string(json.Encode(body.AssetProperties))

	uuid := uuid.New().String()
	assetTemplate := dbmodel.AssetTemplate{
		Metadata:            body.Metadata,
		ExternalMetadataURL: body.ExternalMetadataURL,
		ExternalAssetId:     body.ExternalAssetId,
		EditionTotal:        body.EditionTotal,
		Name:                body.Name,
		CertificateLabel:    body.CertificateLabel,
		AssetLabel:          body.AssetLabel,
		Note:                body.Note,
		AssetProperties:     propsJson,
		Source:              assetSrc,
		ObjectUUID:          uuid,
	}

	// if the work is not an edition then EditionTotal=1
	if assetTemplate.EditionTotal == 0 {
		assetTemplate.EditionTotal = 1
	}
	bundle,err := evaluate(&assetTemplate, body.AssetProperties)
	if err!=nil {
		c.AbortWithStatus(500)
		return
	}
	if bundle.Status == "CREATE" && len(bundle.ErrorMessage) == 0 {
		// the bundle is valid, so create in DB
		db.Create(&assetTemplate)
		//for _, elem := range bundle.Assets { db.Create(&elem) }
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
	AssetProperties map[string]string
	EditionTotal int
	EditionNumber int
	ExternalAssetId int
	AssetName string
	CertificateLabel string
	ExternalMetadataURL string
	CurrentYear string
	Note string
}

func execute(templateParams *TemplateParams, templateString string) (string, error) {

	tMetadata := template.Must(template.New("field").Parse(templateString))
	buf := new(bytes.Buffer)
	err := tMetadata.Execute(buf, templateParams)
	if err != nil {
		fmt.Println("oops evaluating "+templateString+" err="+err.Error())
		return "", err
	}
	return buf.String(), nil
}

const maxAssetNameLength = 32
const maxCertificateLabelLength = 128
const maxAssetLabelLength = 32
const maxNoteLength = 1000


func evaluate(assetTemplate *dbmodel.AssetTemplate, assetProperties map[string]string) (AssetBundle, error) {
  	assets := make([]dbmodel.Asset, 0)
  	params := make([]TemplateParams,0)
  	errors := make([]string,0)
	currentTime := time.Now()
	currentYear := strconv.Itoa(currentTime.Year())
	for i := 1; i<= assetTemplate.EditionTotal ;i++ {
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
			AssetProperties: assetProperties,
			EditionTotal:    assetTemplate.EditionTotal,
			ExternalAssetId: externalAssetId,
			EditionNumber:   i,
		}
		// evaluate asset name first
		assetName, err := execute(&templateParams, assetTemplate.Name)
		if err!=nil {
			errors = append(errors, "assetTemplate.Name templating error : "+err.Error())
		}
		if len(assetName) > maxAssetNameLength {
			errors = append(errors, "asset name length too long : "+assetName)
		}
		templateParams.AssetName = assetName

		// then evaluate certificate label
		certificateLabel,err := execute(&templateParams, assetTemplate.CertificateLabel)
		if err!=nil {
			errors = append(errors, "assetTemplate.CertificateLabel templating error : "+err.Error())
		}
		if len(certificateLabel) > maxCertificateLabelLength {
			errors = append(errors, "certificate label length too long : "+certificateLabel)
		}

		// then evaluate asset label
		assetLabel,err := execute(&templateParams, assetTemplate.AssetLabel)
		if err!=nil {
			errors = append(errors, "assetTemplate.AssetLabel templating error : "+err.Error())
		}
		if len(assetLabel) > maxAssetLabelLength {
			errors = append(errors, "asset label length too long : "+assetLabel)
		}
		templateParams.CertificateLabel = certificateLabel
		externalMetadataURL,err:= execute(&templateParams, assetTemplate.ExternalMetadataURL)
		if err!=nil {
			errors = append(errors, "assetTemplate.ExternalMetadataURL templating error : "+err.Error())
		}
		// TODO: call it and get it
		templateParams.ExternalMetadataURL = externalMetadataURL
		note,err:=execute(&templateParams, assetTemplate.Note)
		if err!=nil {
			errors = append(errors, "assetTemplate.Note templating error : "+err.Error())
		}
		if len(note) > maxNoteLength {
			errors = append(errors, "asset note too long : "+note)
		}
		metadata,err:=execute(&templateParams, assetTemplate.Metadata)
		if err!=nil {
			errors = append(errors, "assetTemplate.Metadata templating error : "+err.Error())
		}
		// upload metadata to IPFS
		ipfsNode := os.Getenv("LOCAL_IPFS_NODE_HOST")+":"+os.Getenv("LOCAL_IPFS_NODE_PORT")
		sh := shell.NewShell(ipfsNode)
		metadataHash, err := sh.Add(bytes.NewReader([]byte(metadata)))
		if err != nil {
			errors = append(errors, "assetTemplate.Metadata templating error : "+err.Error())
		}
		// MD5 stamp
		hasher := md5.New()
		hasher.Write([]byte(metadata))
		md5Hash := hex.EncodeToString(hasher.Sum(nil))
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
			IPFSHashMetadata: 	 metadataHash,
			MetadataHash32: 	 md5Hash,
		}
		assets = append(assets, asset)
		params = append(params, templateParams)
	}
	assetTemplate.Assets=assets
	bundle := AssetBundle{
		Template:     assetTemplate,
		Status:       "CREATE",
//		Params:       params,
		ErrorMessage: errors,
	}
	return bundle, nil
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
	c.JSON(200, asset)
}


