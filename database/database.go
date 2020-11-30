package database

import (
	"fmt"
	"os"

	"gorm.io/gorm"
	"gorm.io/driver/postgres"
)

// Initialize initializes the database
func Initialize() (*gorm.DB, error) {
	dsn := os.Getenv("DB_CONFIG")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected to database")
	db.AutoMigrate(&User{}, &DigitalAssetRoot{}, &DigitalAsset{})
	adminUser :=&User{
		Username:     "uncopied",
		DisplayName:  "Uncopied Admin",
		EmailAddress: "contact@uncopied.art",
		PasswordHash: "####",
	}
	db.Create(adminUser)
	uncopiedAsset := &DigitalAssetRoot{
		AssetNamePrefix:   "uncopied",
		UnitNamePrefix:    "uncopied",
		NotePrefix:        "uncopied",
		AssetURL:          "http://uncopied.art",
		AssetMetadataHash: "####",
		User:              *adminUser,
	}
	db.Create(uncopiedAsset)

	return db, err
}

