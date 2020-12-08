package dbmodel

import (
	"gorm.io/gorm"
)

// DigitalAssetSrc data dbmodel (ex. a raw image)
type DigitalAssetSrc struct {
	gorm.Model

	Issuer User `gorm:"foreignKey:IssuerID"`
	IssuerID uint

	IssuerClaimsAuthorship bool  // if asset was uploaded on behalf of author
	AuthorName  string `sql:"type:text;"` // "friendly name" of author
	AuthorProfileURL string `sql:"type:text;"` // public profile of author, ex. Wikipedia
	SourceLicense string `sql:"type:text;"` // license terms friendly name
	SourceLicenseURL string `sql:"type:text;"` // source file license terms
	IsProxy bool // the asset is a proxy

	// the Issuer will have several options to 'upload' the asset file
	IPFSUploader string `sql:"type:text;"`
	IPFSFilename string `sql:"type:text;"`
	IPFSHash string `gorm:"type:varchar(255);index"`
	IPFSMimetype string `sql:"type:text;"`

	// asset thumbnail
	IPFSHashThumbnail string `sql:"type:text;"`

	// once the digital asset is validated, stamp it with a MD5 checksum
	Stamp string `gorm:"type:char(32);index"`
	StampError string `sql:"type:text;"`

	ExternalSourceID string `sql:"type:text;"`

	// hashes for similarity indexing
	AverageHash uint64 `gorm:"type:numeric;"`
	DifferenceHash uint64 `gorm:"type:numeric;"`
	PerceptionHash uint64 `gorm:"type:numeric;"`

}

const IPFSRootURL = "https://ipfs.io/ipfs"
func ThumbnailURL(src DigitalAssetSrc) string {
	return IPFSRootURL+"/"+src.IPFSHashThumbnail
}