package config

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// InitDBConnection initializes only the database connection without running migrations
func InitDBConnection() (*sql.DB, error) {
	dbConfig := GetAppConfig().DB

	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.DBName, dbConfig.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	log.Println("Connected to database successfully")
	return db, nil
}

// CreateTables creates the necessary tables if they don't exist
func CreateTables(db *sql.DB) error {
	// Create vehicles table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS vehicles (
			id SERIAL PRIMARY KEY,
			license_plate VARCHAR(50) UNIQUE NOT NULL,
			type VARCHAR(20) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return err
	}

	// Create parking_spots table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS parking_spots (
			id SERIAL PRIMARY KEY,
			floor INT NOT NULL,
			row INT NOT NULL,
			"column" INT NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			UNIQUE(floor, row, "column")
		)
	`)
	if err != nil {
		return err
	}

	// Create parking_records table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS parking_records (
			id SERIAL PRIMARY KEY,
			vehicle_id INT NOT NULL REFERENCES vehicles(id),
			parking_spot_id INT NOT NULL REFERENCES parking_spots(id),
			entry_time TIMESTAMP NOT NULL DEFAULT NOW(),
			exit_time TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

// InitializeParkingSpots initializes parking spots based on configuration
func InitializeParkingSpots(db *sql.DB) error {
	parkingConfig := GetAppConfig().Parking

	var err error

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Prepare statement for inserting parking spots
	stmt, err := tx.Prepare(`
		INSERT INTO parking_spots (floor, row, "column", is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (floor, row, "column") DO NOTHING
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Insert parking spots
	for f := 1; f <= parkingConfig.Floors; f++ {
		for r := 1; r <= parkingConfig.Rows; r++ {
			for c := 1; c <= parkingConfig.Columns; c++ {
				_, err = stmt.Exec(f, r, c, true)
				if err != nil {
					return err
				}
			}
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return err
	}

	log.Printf("Initialized parking spots: %d floors, %d rows, %d columns\n", parkingConfig.Floors, parkingConfig.Rows, parkingConfig.Columns)
	return nil
}
