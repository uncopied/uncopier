package dbmodel

import (
	"gorm.io/gorm"
)

// DigitalAsset data dbmodel (ex. a limited edition copy 1/15)
type DigitalAsset struct {
	gorm.Model

	AssetName   string `sql:"type:text;"` // "friendly name" of asset
	UnitName string  `sql:"type:text;"` // used to display asset units to user
	Note string  `sql:"type:text;"` // arbitrary data to be stored in the transaction; here, none is stored

	AssetURL string  `sql:"type:text;"`  // optional string pointing to a URL relating to the asset. 32 character length.
	AssetMetadataHash  string  `sql:"type:text;"` // optional hash commitment of some sort relating to the asset. 32 character length.

	DigitalAssetRoot   DigitalAssetRoot `gorm:"foreignKey:DigitalAssetRootID"`
	DigitalAssetRootID uint

}
