package database

import (
	"fmt"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"gorm.io/gorm"
	"gorm.io/driver/postgres"
	"github.com/uncopied/uncopier/database/dbmodel"
	"time"
)

// Initialize initializes the database
func Initialize() (*gorm.DB, error) {
	fmt.Println("init in database")
	dsn := os.Getenv("DB_CONFIG")
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second,   // Slow SQL threshold
			LogLevel:      logger.Warn, // Log level
			Colorful:      false,         // Disable color
		},
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		fmt.Println("Failed to open database")
		panic(err)
	}
	fmt.Println("Connected to database")
	db.AutoMigrate(&dbmodel.User{},
					&dbmodel.DigitalAssetSrc{},
					&dbmodel.AssetTemplate{},
					&dbmodel.Asset{},
					&dbmodel.Certificate{},
					&dbmodel.Alert{},
					&dbmodel.CertificateToken{},
					&dbmodel.Order{},
					&dbmodel.CertificateIssuance{},
	)
	admin := "uncopied"
	var exists dbmodel.User
	if err := db.Where("user_name = ?", admin).First(&exists).Error; err != nil {
		adminUser :=&dbmodel.User{
			UserName:     "uncopied",
			DisplayName:  "Elian Carsenat",
			EmailAddress: "contact@uncopied.art",
		}
		db.Create(adminUser)
	}
	return db, err
}

