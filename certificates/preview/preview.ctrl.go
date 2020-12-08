package preview

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/uncopied/tallystick"
	"github.com/uncopied/uncopier/api/v1.0/cert"
	"github.com/uncopied/uncopier/database/dbmodel"
	"gorm.io/gorm"
	"net/http"
	"html/template"
)

const uncopied_root = "https://uncopied.org/"
func PermanentURL(certId string) string {
	return uncopied_root+"/certificates/"+certId
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

	if asset.Source.Stamp=="" {
		fmt.Println("Asset is not stamped ")
		c.AbortWithStatus(409)
		return
	}

	// preview the first
	var first = asset.Assets[0]

	// create a tally stick for preview
	t := tallystick.Tallystick{
		CertificateLabel:                first.CertificateLabel,
		PrimaryLinkURL:                  PermanentURL("PrimaryIDDummy"),
		SecondaryLinkURL:                PermanentURL("SecondaryIDDummy"),
		IssuerTokenURL:                  PermanentURL("IssuerTokenURLDummy"),
		OwnerTokenURL:                   PermanentURL("OwnerTokenURLDummy"),
		PrimaryAssetVerifierTokenURL:    PermanentURL("PrimaryAssetVerifierTokenURLDummy"),
		SecondaryAssetVerifierTokenURL:  PermanentURL("SecondaryAssetVerifierTokenURLDummy"),
		PrimaryOwnerVerifierTokenURL:    PermanentURL("PrimaryOwnerVerifierTokenURLDummy"),
		SecondaryOwnerVerifierTokenURL:  PermanentURL("SecondaryOwnerVerifierTokenURLDummy"),
		PrimaryIssuerVerifierTokenURL:   PermanentURL("PrimaryIssuerVerifierTokenURLDummy"),
		SecondaryIssuerVerifierTokenURL: PermanentURL("SecondaryIssuerVerifierTokenURLDummy"),
		MailToContentLeft:               cert.MailTo(),
		MailToContentRight:              cert.MailTo(),
	}
	var buf bytes.Buffer
	err := tallystick.DrawSVG(&t,&buf)
	if err!=nil {
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
	c.HTML(http.StatusOK, "preview.tmpl", gin.H{
		"certificateLabel": first.CertificateLabel,
		"tallystick": tallystickHtml,
		"thumbnail": asset.Source.IPFSHashThumbnail,
	})
}

