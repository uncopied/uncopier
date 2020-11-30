package model

import (
	"gorm.io/gorm"
)

// User data model
type User struct {
	gorm.Model
	UserName string
	DisplayName	string
	EmailAddress string
	PasswordHash string
}

