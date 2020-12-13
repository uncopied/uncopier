package src

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/algorand/go-algorand-sdk/encoding/json"
	"github.com/auyer/steganography"
	"github.com/corona10/goimagehash"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"github.com/ipfs/go-ipfs-api"
	"github.com/nfnt/resize"
	"github.com/pkg/errors"
	"github.com/uncopied/uncopier/database/dbmodel"
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
		ExternalSourceID string  `json:"external_source_id"`
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
		ExternalSourceID: 		body.ExternalSourceID,
		SourceLicense:          body.SourceLicense,
		SourceLicenseURL:       body.SourceLicenseURL,
		IPFSUploader:           body.IPFSUploader,
		IPFSFilename:           body.IPFSFilename,
		IPFSHash:               body.IPFSHash,
		IPFSMimetype:           body.IPFSMimetype,
	}

	db.Create(&asset)
	md5Hash, err := stamp(&asset, db)
	asset.Stamp = md5Hash
	if err != nil {
		// should we return some error code when stamping failed?
		asset.StampError = err.Error()
		fmt.Println("error "+err.Error())
	}
	tx := db.Updates(&asset)
	if tx.Error != nil {
		fmt.Println("DB transaction error ")
		c.AbortWithStatus(500)
		return
	}
	c.JSON(200, &asset)
}




const MaxContentLength = 25000000
const ThumbnailWidthHeight = 720

func stamp(asset *dbmodel.DigitalAssetSrc, db *gorm.DB) (string, error) {

	// check for any duplicate based on MD5
	// check if exists
	var exists dbmodel.DigitalAssetSrc
	if err := db.Where("ip_fs_hash = ? AND stamp <> '' ", asset.IPFSHash).First(&exists).Error; err == nil {
		msg := "IPFS already contains stamped file with IPFSHash "+ asset.IPFSHash
		// create an exception
		exception := dbmodel.Alert{
			Source : *asset,
			OtherSource : exists,
			Status : "NEW",
			Message : msg,
		}
		db.Create(&exception)
		return "",errors.New(msg)
	}
	//const IPFSRootURL = "https://ipfs.io/ipfs"
	ipfsRootURL := os.Getenv("IPFS_ROOT_URL")
	sourceURL := ipfsRootURL +"/"+asset.IPFSHash;
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

	// MD5 stamp
	hasher := md5.New()
	hasher.Write(bodyBytes)
	md5Hash := hex.EncodeToString(hasher.Sum(nil))

	if err := db.Where("stamp = ?", md5Hash).First(&exists).Error; err == nil {
		msg := "IPFS already contains file with stamp "+ md5Hash +", legit collision is unlikely - could be plagiarism"
		// create an exception
		exception := dbmodel.Alert{
			Source : *asset,
			OtherSource : exists,
			Status : "NEW",
			Message : msg,
		}
		db.Create(&exception)
		return "",errors.New(msg)
	}

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

	// Save the main file in cache
	ipfsCacheDir := os.Getenv("LOCAL_IPFS_CACHE")
	filePath := ipfsCacheDir+"/"
	{
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
	}


	// resize using Lanczos resampling
	// and preserve aspect ratio
	var thumbnail image.Image
	if imgConfig.Width >= imgConfig.Height {
		thumbnail = resize.Resize(ThumbnailWidthHeight, 0, img, resize.Lanczos3)
	} else {
		thumbnail = resize.Resize(0, ThumbnailWidthHeight, img, resize.Lanczos3)
	}

	// compute thumbnail hashes for similarity indexing
	averageHash, _ := goimagehash.AverageHash(thumbnail)
	differenceHash, _ := goimagehash.DifferenceHash(thumbnail)
	perceptionHash, _ := goimagehash.PerceptionHash(thumbnail)
	asset.AverageHash = averageHash.GetHash()
	asset.DifferenceHash = differenceHash.GetHash()
	asset.PerceptionHash = perceptionHash.GetHash()


	// add steganography
	type Stegano struct {
		Source string
		SourceType string
		IPFSHash string
		Stamp string
	}
	stegano := Stegano{
		Source:     "uncopied.org",
		SourceType: "thumbnail",
		IPFSHash:   asset.IPFSHash,
		Stamp:      md5Hash,
	}
	message := json.Encode(stegano)
	sizeOfMessage := steganography.GetMessageSizeFromImage(thumbnail)
	if int(sizeOfMessage) < len(message) {
		return "",errors.New("IPFS thumbnail too small for stegano "+sourceURL)
	} else {
		w := new(bytes.Buffer)
		err := steganography.Encode(w, thumbnail, message) // Encode the message into the image
		if err != nil {
			return "",err
		}
		ipfsNode := os.Getenv("LOCAL_IPFS_NODE_HOST")+":"+os.Getenv("LOCAL_IPFS_NODE_PORT")
		sh := shell.NewShell(ipfsNode)
		cid, err := sh.Add(bytes.NewReader(w.Bytes()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s", err)
			os.Exit(1)
		}
		asset.IPFSHashThumbnail = cid
		// Save the thumbnail file in cache
		filePath := ipfsCacheDir+"/"
		{
			fileName := cid	+".png"
			out, err := os.Create(filePath+fileName)
			if err != nil {
				return "",err
			}
			defer out.Close()
			// Write the body to file
			_, err = out.Write(w.Bytes())
			if err != nil {
				return "",err
			}
		}
		// pin the original image
		err = sh.Pin(asset.IPFSHash)
		if err!= nil {
			return "",err
		}
	}
	return md5Hash,err
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

