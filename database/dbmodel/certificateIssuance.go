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

}