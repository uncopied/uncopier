package dbmodel

import (
	"gorm.io/gorm"
)

// User data dbmodel
type User struct {
	gorm.Model
	UserName     string
	DisplayName  string
	EmailAddress string
	PasswordHash string
	// blockchain specifics
	EthereumAddress       string
}

