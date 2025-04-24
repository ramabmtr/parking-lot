package main

import (
	"log"

	"parking-lot/config"
)

func main() {
	config.InitAppConfig()

	log.Println("Starting database migration...")

	db, err := config.InitDBConnection()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	err = config.CreateTables(db)
	if err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	err = config.InitializeParkingSpots(db)
	if err != nil {
		log.Fatalf("Failed to initialize parking spots: %v", err)
	}

	log.Println("Database migration completed successfully.")
}
