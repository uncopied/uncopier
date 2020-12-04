package dbmodel
import (
	"gorm.io/gorm"
)

// Certificate data dbmodel
type Certificate struct {
	gorm.Model

	Issuer User `gorm:"foreignKey:IssuerID"`
	IssuerID uint

	Printer User `gorm:"foreignKey:PrinterID"`
	PrinterID uint

	PrimaryConservator User `gorm:"foreignKey:PrimaryConservatorID"`
	PrimaryConservatorID uint

	SecondaryConservator User `gorm:"foreignKey:SecondaryConservatorID"`
	SecondaryConservatorID uint

	CertificateLabel string `sql:"type:text;"` // short text at center of certificate

	// where the document is stored
	ImmutableDocumentPrimaryURL   string `sql:"type:text;"` // stable URL to immutable PDF document
	ImmutableDocumentSecondaryURL   string `sql:"type:text;"` // stable URL to immutable PDF document backup
	ImmutableDocumentHash   string `sql:"type:text;"` // "friendly name" of asset

	// blockchain IDs
	CertificateSecondaryScheme  string `sql:"type:text;"` // the certificate's secondary system ex. Algorand blockchain
	CertificateSecondaryID  string `sql:"type:text;"` // the certificate's ID in secondary system ex. Algorand blockchain assetId
	CertificateSecondaryURL  string `sql:"type:text;"` // link to the certificate in secondary system ex. stable link to blockchain explorer

	// role tokens : qrcodes split
	IssuerTokenID uint
	OwnerTokenID uint
	PrimaryAssetVerifierTokenID uint
	SecondaryAssetVerifierTokenID uint
	PrimaryOwnerVerifierTokenID uint
	SecondaryOwnerVerifierTokenID uint
	PrimaryIssuerVerifierTokenID uint
	SecondaryIssuerVerifierTokenID uint

}
