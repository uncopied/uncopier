	package upload

import (
	"fmt"
	"github.com/gin-gonic/gin"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/uncopied/uncopier/database/dbmodel"
	"gorm.io/gorm"
	"net/http"
	"os"
)

const must_auth = true

type IPFSResponse struct{
	IPFSHash   string
	Message    string
}

// https://chenyitian.gitbooks.io/gin-web-framework/content/docs/12.html
// curl -X POST http://localhost:8081/upload -F "upload[]=mstile-310x310.png" -H "Content-Type: multipart/form-data"
//curl -X POST http://localhost:8081/upload/  -F "file=mstile-310x310.png" -H "Content-Type: multipart/form-data"
func upload(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	if must_auth {
		userName := c.MustGet("user")

		// check if user
		var user dbmodel.User
		if err := db.Where("user_name = ?", userName).First(&user).Error; err != nil {
			fmt.Println("User name not found ", userName)
			c.AbortWithStatus(409)
			return
		}
	}
	response := IPFSResponse {
		IPFSHash: "",
		Message:  "",
	}
	// single file
	file, _ := c.FormFile("file")
	src, err := file.Open()
	if err != nil {
		response.Message =err.Error()
		c.JSON(http.StatusInternalServerError, response)
	}
	defer src.Close()

	ipfsNode := os.Getenv("LOCAL_IPFS_NODE_HOST")+":"+os.Getenv("LOCAL_IPFS_NODE_PORT")
	sh := shell.NewShell(ipfsNode)

	cid, err := sh.Add(src) //bytes.NewReader(fileData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		c.JSON(http.StatusInternalServerError, response)
	}
	response.IPFSHash = cid
	c.JSON(http.StatusOK, response)
}



