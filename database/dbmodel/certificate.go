package dbmodel
import (
	"gorm.io/gorm"
)

// Certificate data dbmodel
type Certificate struct {
	gorm.Model

	// short text at center of certificate
	CertificateLabel string `sql:"type:text;"`

	// the user issuing the certificate
	Issuer User `gorm:"foreignKey:IssuerID"`
	IssuerID uint

	// the user printing the certificate (also responsible to send proofs to the primary & secondary conservator)
	Printer User `gorm:"foreignKey:PrinterID"`
	PrinterID uint

	// the primary conservator (normally : UNCOPIED)
	PrimaryConservator User `gorm:"foreignKey:PrimaryConservatorID"`
	PrimaryConservatorID uint

	// the secondary conservator (ex. a trusted physical and digital archiver)
	SecondaryConservator User `gorm:"foreignKey:SecondaryConservatorID"`
	SecondaryConservatorID uint

	// where the document is stored immutably
	ImmutableDocumentPrimaryURL   string `sql:"type:text;"` // stable URL to immutable PDF document
	ImmutableDocumentSecondaryURL   string `sql:"type:text;"` // stable URL to immutable PDF document backup
	ImmutableDocumentHash   string `sql:"type:text;"` // "friendly name" of asset

	// blockchain IDs
	CertificateSecondaryScheme  string `sql:"type:text;"` // the certificate's secondary system ex. Algorand blockchain
	CertificateSecondaryID  string `sql:"type:text;"` // the certificate's ID in secondary system ex. Algorand blockchain assetId
	CertificateSecondaryURL  string `sql:"type:text;"` // link to the certificate in secondary system ex. stable link to blockchain explorer

	// role tokens : qrcodes split
	IssuerToken CertificateToken `gorm:"foreignKey:IssuerTokenID"`
	IssuerTokenID uint
	OwnerToken CertificateToken `gorm:"foreignKey:OwnerTokenID"`
	OwnerTokenID uint
	PrimaryAssetVerifierToken CertificateToken `gorm:"foreignKey:PrimaryAssetVerifierTokenID"`
	PrimaryAssetVerifierTokenID uint
	SecondaryAssetVerifierToken CertificateToken `gorm:"foreignKey:SecondaryAssetVerifierTokenID"`
	SecondaryAssetVerifierTokenID uint
	PrimaryOwnerVerifierToken CertificateToken `gorm:"foreignKey:PrimaryOwnerVerifierTokenID"`
	PrimaryOwnerVerifierTokenID uint
	SecondaryOwnerVerifierToken CertificateToken `gorm:"foreignKey:SecondaryOwnerVerifierTokenID"`
	SecondaryOwnerVerifierTokenID uint
	PrimaryIssuerVerifierToken CertificateToken `gorm:"foreignKey:PrimaryIssuerVerifierTokenID"`
	PrimaryIssuerVerifierTokenID uint
	SecondaryIssuerVerifierToken CertificateToken `gorm:"foreignKey:SecondaryIssuerVerifierTokenID"`
	SecondaryIssuerVerifierTokenID uint

}
