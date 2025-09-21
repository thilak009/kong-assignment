package db

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func Init(opts ...gorm.Option) {
	var err error

	dsn := fmt.Sprintf("postgres://%s/%s?sslmode=disable&user=%s&password=%s", os.Getenv("DB_HOST"), os.Getenv("DB_NAME"), os.Getenv("DB_USER"), os.Getenv("DB_PASS"))
	db, err = gorm.Open(postgres.Open(dsn), opts...)
	if err != nil {
		panic("failed to connect to database @" + os.Getenv("DB_HOST") + " error: " + err.Error())
	}
}


// RunMigrations runs database migrations for provided models
func RunMigrations(models ...interface{}) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Auto-migrate all provided models
	err := db.AutoMigrate(models...)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func GetDB() *gorm.DB {
	return db
}
