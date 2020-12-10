package dbmodel

import (
	"gorm.io/gorm"

)

// Order fulfillment dbmodel
type Order struct {
	gorm.Model

	OrderUUID string

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

	// is do it yourself?
	IsDIY bool
}
