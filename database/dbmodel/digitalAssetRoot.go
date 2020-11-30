package dbmodel

import (
	"gorm.io/gorm"
)

// DigitalAssetRoot data dbmodel (ex. a raw image)
type DigitalAssetRoot struct {
	gorm.Model

	AssetNamePrefix   string `sql:"type:text;"` // "friendly name" of asset
	UnitNamePrefix string  `sql:"type:text;"` // used to display asset units to user
	NotePrefix string  `sql:"type:text;"` // arbitrary data to be stored in the transaction; here, none is stored

	User User `gorm:"foreignKey:UserID"`
	UserID uint

}

