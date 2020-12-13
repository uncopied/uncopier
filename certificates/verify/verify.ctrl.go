package verify

import (
	"github.com/gin-gonic/gin"
	"github.com/uncopied/uncopier/database/dbmodel"
	"gorm.io/gorm"
	"net/http"
	"os"
)

func verify(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	token := c.Param("id")
	// TODO change this : for now just use the ID as token
	var certificateIssuance dbmodel.CertificateIssuance
	if err := db.Preload("Asset").Where("id = ?", token).First(&certificateIssuance).Error; err != nil {
		c.AbortWithStatus(404)
		return
	}
	var assetTemplate dbmodel.AssetTemplate
	if err := db.Preload("Source.Issuer").Preload("Source").Preload("Assets").Where("id = ?", certificateIssuance.Asset.AssetTemplateID).First(&assetTemplate).Error; err != nil {
		c.AbortWithStatus(404)
		return
	}
	localIPFSHost:=os.Getenv("LOCAL_IPFS_HOST")
	localIPFSPort:=os.Getenv("LOCAL_IPFS_PORT")
	c.HTML(http.StatusOK, "verify.tmpl", gin.H{
		"certificateIssuance": certificateIssuance,
		"assetTemplate" : assetTemplate,
		"source" : assetTemplate.Source,
		"localIPFSHost":localIPFSHost,
		"localIPFSPort":localIPFSPort,
	})
}
