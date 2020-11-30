package dbmodel
import (
	"gorm.io/gorm"
)

// Certificate data dbmodel
type Certificate struct {
	gorm.Model

	Documentation   string `sql:"type:text;"` // "friendly name" of asset

	Issuer User `gorm:"foreignKey:IssuerID"`
	IssuerID uint

}
