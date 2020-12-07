package dbmodel

import (
	"gorm.io/gorm"
)

// Alert data dbmodel
type Alert struct {
	gorm.Model

	Source DigitalAssetSrc `gorm:"foreignKey:SourceID"`
	SourceID uint

	OtherSource DigitalAssetSrc `gorm:"foreignKey:OtherSourceID"`
	OtherSourceID uint

	Status string `sql:"type:text;"`
	Message string `sql:"type:text;"`
}
