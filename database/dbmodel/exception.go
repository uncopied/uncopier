package dbmodel
import (
	"gorm.io/gorm"
)

// DigitalAssetSrc data dbmodel (ex. a raw image)
type UncopierException struct {
	gorm.Model

	Source1 DigitalAssetSrc `gorm:"foreignKey:DigitalAssetSrcID"`
	Source1ID uint

	Source2 DigitalAssetSrc `gorm:"foreignKey:DigitalAssetSrcID"`
	Source2ID uint

	Status string `sql:"type:text;"`
	Message string `sql:"type:text;"`
}
