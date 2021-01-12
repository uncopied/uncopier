package view

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/uncopied/uncopier/database/dbmodel"
	"gorm.io/gorm"
	"net/http"
)

func preview(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	uuid := c.Param("uuid")
	var assetTemplate dbmodel.AssetTemplate
	if err := db.Preload("Source.Issuer").Preload("Source").Preload("Assets").Where("object_uuid = ?", uuid).First(&assetTemplate).Error; err != nil {
		c.AbortWithStatus(404)
		return
	}

	if assetTemplate.Source.Stamp == "" {
		fmt.Println("Asset is not stamped ")
		c.AbortWithStatus(409)
		return
	}
	// view the first
	var first = assetTemplate.Assets[0]
	c.HTML(http.StatusOK, "view.tmpl", gin.H{
		"asset":  first,
		"source": assetTemplate.Source,
	})
}

func view(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	token := c.Param("token")

	// TODO change this : for now just use the ID as token
	var asset dbmodel.Asset
	if err := db.Where("id = ?", token).First(&asset).Error; err != nil {
		c.AbortWithStatus(404)
		return
	}

	var assetTemplate dbmodel.AssetTemplate
	if err := db.Preload("Source.Issuer").Preload("Source").Preload("Assets").Where("id = ?", asset.AssetTemplateID).First(&assetTemplate).Error; err != nil {
		c.AbortWithStatus(404)
		return
	}

	c.HTML(http.StatusOK, "view.tmpl", gin.H{
		"asset":  asset,
		"source": assetTemplate.Source,
	})
}
