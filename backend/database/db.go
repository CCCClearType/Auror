package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "admin")
	password := getEnv("DB_PASSWORD", "admin")
	name := getEnv("DB_NAME", "auror_vapor")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC", host, user, password, name, port)
	var database *gorm.DB
	var err error

	for i := 0; i < 10; i++ {
		database, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			sqlDB, err := database.DB()
			if err == nil && sqlDB.Ping() == nil {
				break
			}
		}
		log.Printf("Database not ready, retrying in 3 seconds... (%d/10)\n", i+1)
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		log.Fatal("Failed to connect to database after retries:", err)
	}
	DB = database

	fmt.Println("Database connection successfully opened")
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
