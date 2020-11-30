package assets

import (
	"../../../database/dbmodel"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)


func create(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	type RequestBody struct {
		AssetNamePrefix    string `json:"asset_name_p" binding:"required"`
		UnitNamePrefix string `json:"unit_name_p" binding:"required"`
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

	asset := dbmodel.DigitalAssetRoot{
		AssetNamePrefix: body.AssetNamePrefix,
		UnitNamePrefix:  body.UnitNamePrefix,
		User:            user,
	}
	db.Create(&asset)
	c.JSON(200, &asset)
}

func list(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	cursor := c.Query("cursor")
	recent := c.Query("recent")

	var assets []dbmodel.DigitalAssetRoot

	if cursor == "" {
		if err := db.Preload("User").Limit(10).Order("id desc").Find(&assets).Error; err != nil {
			c.AbortWithStatus(500)
			return
		}
	} else {
		condition := "id < ?"
		if recent == "1" {
			condition = "id > ?"
		}
		if err := db.Preload("User").Limit(10).Order("id desc").Where(condition, cursor).Find(&assets).Error; err != nil {
			c.AbortWithStatus(500)
			return
		}
	}

	c.JSON(200, assets)
}

func read(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")
	var asset dbmodel.DigitalAssetRoot

	// auto preloads the related dbmodel
	// http://gorm.io/docs/preload.html#Auto-Preloading
	if err := db.Set("gorm:auto_preload", true).Where("id = ?", id).First(&asset).Error; err != nil {
		c.AbortWithStatus(404)
		return
	}

	c.JSON(200, asset)
}

func remove(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")

	userName := c.MustGet("user")
	// check if user
	var user dbmodel.User
	if err := db.Where("user_name = ?", userName).First(&user).Error; err == nil {
		c.AbortWithStatus(409)
		return
	}

	var asset dbmodel.DigitalAssetRoot
	if err := db.Where("id = ?", id).First(&asset).Error; err != nil {
		c.AbortWithStatus(404)
		return
	}

	if asset.UserID != user.ID {
		c.AbortWithStatus(403)
		return
	}

	db.Delete(&asset)
	c.Status(204)
}

func update(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")

	userName := c.MustGet("user")
	// check if user
	var user dbmodel.User
	if err := db.Where("user_name = ?", userName).First(&user).Error; err == nil {
		c.AbortWithStatus(409)
		return
	}

	type RequestBody struct {
		AssetNamePrefix    string `json:"asset_name_p" binding:"required"`
		UnitNamePrefix string `json:"unit_name_p" binding:"required"`
	}

	var body RequestBody
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithStatus(400)
		return
	}

	var asset dbmodel.DigitalAssetRoot
	if err := db.Preload("User").Where("id = ?", id).First(&asset).Error; err != nil {
		c.AbortWithStatus(404)
		return
	}

	if asset.UserID != user.ID {
		c.AbortWithStatus(403)
		return
	}
	asset.AssetNamePrefix = body.AssetNamePrefix
	asset.UnitNamePrefix =  body.UnitNamePrefix

	db.Save(&asset)
	c.JSON(200, asset)
}
