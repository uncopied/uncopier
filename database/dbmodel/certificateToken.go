package dbmodel
import (
	"gorm.io/gorm"
)

// Certificate data dbmodel
type CertificateToken struct {
	gorm.Model

	CertificateID uint

	Role string `sql:"type:text;"` // role, ex. Issuer, Owner, etc.
	Token string `sql:"type:text;"` // token
	TokenHash string `sql:"type:text;"` // token hash
}