package database

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"simak-backend/config"
)

// DB is the global GORM database instance.
var DB *gorm.DB

// Connect initialises the PostgreSQL connection pool via GORM.
func Connect(cfg *config.Config) *gorm.DB {
	logLevel := logger.Silent
	if cfg.App.Debug {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Printf("⚠️ Failed to connect to database: %v", err)
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("⚠️ Failed to get underlying sql.DB: %v", err)
		return nil
	}

	// Connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	DB = db
	log.Println("Database connected successfully")
	return db
}

// Ping checks whether the database connection is alive.
func Ping(db *gorm.DB) error {
	if db == nil {
		return gorm.ErrInvalidDB
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
