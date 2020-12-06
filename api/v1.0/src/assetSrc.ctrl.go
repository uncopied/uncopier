package src

import (
	"../../../database/dbmodel"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/algorand/go-algorand-sdk/encoding/json"
	"github.com/auyer/steganography"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"github.com/ipfs/go-ipfs-api"
	"github.com/nfnt/resize"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)


func create(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	type RequestBody struct {
		IssuerClaimsAuthorship bool `json:"issuer_claims_ownership"`
		AuthorName string  `json:"author_name"`
		AuthorProfileURL string  `json:"author_profile_url"`
		SourceLicense string  `json:"source_license"`
		SourceLicenseURL string `json:"source_license_url"`
		IPFSUploader string `json:"ipfs_uploader"`
		IPFSFilename string `json:"ipfs_filename"`
		IPFSHash string  `json:"ipfs_hash" binding:"required"`
		IPFSMimetype string `json:"ipfs_mimetype"`
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

	asset := dbmodel.DigitalAssetSrc{
		Issuer:                 user,
		IssuerClaimsAuthorship: false,
		AuthorName:             body.AuthorName,
		AuthorProfileURL:       body.AuthorProfileURL,
		SourceLicense:          body.SourceLicense,
		SourceLicenseURL:       body.SourceLicenseURL,
		IPFSUploader:           body.IPFSUploader,
		IPFSFilename:           body.IPFSFilename,
		IPFSHash:               body.IPFSHash,
		IPFSMimetype:           body.IPFSMimetype,
	}

	db.Create(&asset)
	md5, err := stamp(&asset)
	asset.Stamp = md5
	if err != nil {
		// should we return some error code when stamping failed?
		asset.StampError = err.Error()
		fmt.Println("error "+err.Error())
	}
	db.Updates(&asset)
	c.JSON(200, &asset)
}

const IPFSRootURL = "https://ipfs.io/ipfs"
const IPFSNode = "localhost:5001"

const LocalCacheDIR = "d:/ifps_cache"
const MaxContentLength = 25000000
const ThumbnailWidthHeight = 720

func stamp(asset *dbmodel.DigitalAssetSrc) (string, error) {
	sourceURL := IPFSRootURL +"/"+asset.IPFSHash;
	resp, err := http.Get(sourceURL)
	if err!=nil {
		return "",err
	}
	if resp.StatusCode != 200 {
		return "",errors.New("IPFS failed to get "+sourceURL+", got status "+resp.Status)
	}
	// check max size
	if resp.ContentLength > MaxContentLength {
		return "",errors.New("IPFS content length at "+sourceURL+", too large")
	}
	defer resp.Body.Close()

	// read the body into mem
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	mime := mimetype.Detect(bodyBytes)
	if asset.IPFSMimetype == "" {
		asset.IPFSMimetype = mime.String()
	} else if asset.IPFSMimetype != mime.String() {
		return "",errors.New("IPFS file mimetype mismatch "+sourceURL+", "+asset.IPFSMimetype+" != "+mime.String())
	}
	if asset.IPFSFilename == "" {
		asset.IPFSFilename = asset.IPFSHash+mime.Extension()
	} else if ! strings.HasSuffix(asset.IPFSFilename, mime.Extension()) {
		return "",errors.New("IPFS file name not valid for mimetype  "+sourceURL+", "+asset.IPFSMimetype+" fileName "+asset.IPFSFilename+" should end with extension "+ mime.Extension())
	}

	var img image.Image
	var imgConfig image.Config
	// create thumbnail
	if mime.Extension() == ".png" {
		// decode jpeg into image.Image
		img, err = png.Decode(bytes.NewReader(bodyBytes))
		if err != nil {
			return "",errors.New("IPFS PNG file could not be decoded from  "+sourceURL)
		}
		imgConfig, err = png.DecodeConfig(bytes.NewReader(bodyBytes))
		if err != nil {
			return "",errors.New("IPFS PNG file config could not be decoded from  "+sourceURL)
		}
	} else if mime.Extension() == ".jpg" {
		// decode jpeg into image.Image
		img, err = jpeg.Decode(bytes.NewReader(bodyBytes))
		if err != nil {
			return "",errors.New("IPFS JPG file could not be decoded from  "+sourceURL)
		}
		imgConfig, err = jpeg.DecodeConfig(bytes.NewReader(bodyBytes))
		if err != nil {
			return "",errors.New("IPFS JPG file config could not be decoded from  "+sourceURL)
		}
	} else {
		return "",errors.New("IPFS thumbnailing : currently, only PNG and JPG files are supported "+sourceURL)
	}

	// Create the file
	filePath := LocalCacheDIR+"/"+asset.Issuer.UserName+"/"
	fileName := asset.IPFSHash+mime.Extension()
	err = os.MkdirAll(filePath,os.ModePerm)
	if err != nil {
		return "",err
	}
	out, err := os.Create(filePath+fileName)
	if err != nil {
		return "",err
	}
	defer out.Close()

	// Write the body to file
	_, err = out.Write(bodyBytes)
	if err != nil {
		return "",err
	}

	// resize using Lanczos resampling
	// and preserve aspect ratio
	var thumbnail image.Image
	if imgConfig.Width >= imgConfig.Height {
		thumbnail = resize.Resize(ThumbnailWidthHeight, 0, img, resize.Lanczos3)
	} else {
		thumbnail = resize.Resize(0, ThumbnailWidthHeight, img, resize.Lanczos3)
	}

	// MD5 stamp
	hasher := md5.New()
	hasher.Write(bodyBytes)
	md5 := hex.EncodeToString(hasher.Sum(nil))

	// add steganography
	type Stegano struct {
		Source string
		SourceType string
		IPFSHash string
		Stamp string
	}
	stegano := Stegano{
		Source:   "uncopied.org",
		SourceType:   "thumbnail",
		IPFSHash: asset.IPFSHash,
		Stamp: md5,
	}
	message := json.Encode(stegano)
	sizeOfMessage := steganography.GetMessageSizeFromImage(thumbnail)

	if int(sizeOfMessage) < len(message) {
		if err != nil {
			return "",errors.New("IPFS thumbnail too small for stegano "+sourceURL)
		}
	} else {
		w := new(bytes.Buffer)
		err := steganography.Encode(w, thumbnail, message) // Encode the message into the image
		if err != nil {
			return "",err
		}
		sh := shell.NewShell(IPFSNode)
		cid, err := sh.Add(bytes.NewReader(w.Bytes()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s", err)
			os.Exit(1)
		}
		asset.IPFSHashThumbnail = cid
	}
	return md5,err
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

	var asset dbmodel.DigitalAssetSrc
	if err := db.Where("id = ?", id).First(&asset).Error; err != nil {
		c.AbortWithStatus(404)
		return
	}

	if asset.Issuer.ID != user.ID {
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
		IssuerClaimsAuthorship bool
		AuthorName string
		AuthorProfileURL string
		SourceLicense string
		SourceLicenseURL string
	}

	var body RequestBody
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithStatus(400)
		return
	}

	var asset dbmodel.DigitalAssetSrc
	if err := db.Preload("User").Where("id = ?", id).First(&asset).Error; err != nil {
		c.AbortWithStatus(404)
		return
	}

	if asset.Issuer.ID != user.ID {
		c.AbortWithStatus(403)
		return
	}
	asset.IssuerClaimsAuthorship = body.IssuerClaimsAuthorship
	asset.AuthorName =  body.AuthorName
	asset.AuthorProfileURL = body.AuthorProfileURL
	asset.SourceLicense =  body.SourceLicense
	asset.SourceLicenseURL =  body.SourceLicenseURL

	db.Save(&asset)
	c.JSON(200, asset)
}
