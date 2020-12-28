package dbmodel

import (
	"gorm.io/gorm"

)

// Order fulfillment dbmodel
type Order struct {
	gorm.Model

	// the order UUID
	OrderUUID string `gorm:"type:char(36);index"`

	// asset template
	AssetTemplate AssetTemplate `gorm:"foreignKey:AssetTemplateID"`
	AssetTemplateID uint

	// payment/invoice status
	PaymentStatus string
	InvoiceURL string

	// delivery status
	DeliveryStatus string

	// quality check status
	QualityStatus string
	Quality int

	// the file delivry
	ZipBundle string

	// is do it yourself?
	IsDIY bool

	// Paypal Details
	PaypalDetails string `sql:"type:text;"`
}
