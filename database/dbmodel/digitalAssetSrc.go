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
	IPFSHash string `sql:"type:text;"`
	IPFSMimetype string `sql:"type:text;"`

	// asset thumbnail
	IPFSHashThumbnail string `sql:"type:text;"`

	// once the digital asset is validated, stamp it with a MD5 checksum
	Stamp string `sql:"type:text;"`
	StampError string `sql:"type:text;"`

}