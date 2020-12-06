package database

import (
	"fmt"
	"os"
	"gorm.io/gorm"
	"gorm.io/driver/postgres"
	"github.com/uncopied/uncopier/dbmodel"
)

// Initialize initializes the database
func Initialize() (*gorm.DB, error) {
	dsn := os.Getenv("DB_CONFIG")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected to database")
	db.AutoMigrate(&dbmodel.User{}, &dbmodel.DigitalAssetSrc{}, &dbmodel.DigitalAsset{}, &dbmodel.Certificate{}, &dbmodel.CertificateToken{})
	adminUser :=&dbmodel.User{
		UserName:     "uncopied",
		DisplayName:  "Elian Carsenat",
		EmailAddress: "contact@uncopied.art",
	}
	db.Create(adminUser)
	return db, err
}

