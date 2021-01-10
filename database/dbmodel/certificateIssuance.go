package dbmodel

import (
	"gorm.io/gorm"
)

// CertificateIssuance dbmodel
type CertificateIssuance struct {
	gorm.Model

	Asset Asset `gorm:"foreignKey:AssetID"`
	AssetID uint

	Order Order `gorm:"foreignKey:OrderID"`
	OrderID uint

	Certificate Certificate `gorm:"foreignKey:CertificateID"`
	CertificateID uint

	AlgorandTransactionID string `sql:"type:text;"`

	Metadata string  `sql:"type:text;"`
	MetadataHash string  `sql:"type:text;"`
}