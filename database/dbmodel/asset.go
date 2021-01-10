package dbmodel

import (
	"gorm.io/gorm"
)

// Asset data dbmodel (ex. a limited edition copy 1/15)
type Asset struct {
	gorm.Model

	// The asset template metadata
	Metadata         string `sql:"type:text;"`
	MD5HashMetadata  string `sql:"type:text;"`
	IPFSHashMetadata string `sql:"type:text;"`

	// An external metadata URL, ex. https://metadata.mintable.app/ao1YMdMJWwF39585fg9C/52
	ExternalMetadataURL  string `sql:"type:text;"`

	// An external asset ID for a sequencing ex. 52
	ExternalAssetId int

	// The number of copies when multiple copies of a piece of artwork are produced - e.g. for a limited edition of 20 prints, 'artEdition' refers to the total number of copies (in this example "20").
	EditionTotal int

	// The edition number in the edition total
	EditionNumber int

	// Friendly name (must be short)
	Name string `sql:"type:text;"`

	// Certificate label
	CertificateLabel string `sql:"type:text;"`

	// Asset label (max size : 32)
	AssetLabel string `sql:"type:text;"`

	// Arbitrary data to be stored in the transaction
	Note string  `sql:"type:text;"` //

	AssetTemplateID uint

}
