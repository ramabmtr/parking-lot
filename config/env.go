package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
	"parking-lot/domain"
)

var (
	envOnce   sync.Once
	appConfig AppConfig
)

type AppConfig struct {
	DB      DBConfig
	Parking ParkingConfig
	Server  ServerConfig
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type ParkingConfig struct {
	Floors          int
	Rows            int
	Columns         int
	FloorVehicleMap map[int]domain.VehicleType
}

type ServerConfig struct {
	Port string
}

// InitAppConfig is a syntax sugar to initialize the application configuration
func InitAppConfig() {
	_ = GetAppConfig()
}

// GetAppConfig returns the application configuration
func GetAppConfig() AppConfig {
	envOnce.Do(func() {
		err := godotenv.Load()
		if err != nil {
			log.Println("Warning: .env file not found, using default environment variables")
		}

		appConfig = AppConfig{
			DB:      getDBConfig(),
			Parking: getParkingConfig(),
			Server:  getServerConfig(),
		}
	})

	return appConfig
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getDBConfig() DBConfig {
	port, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		port = 5432
	}

	return DBConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     port,
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "parking_lot"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

func getParkingConfig() ParkingConfig {
	floors, err := strconv.Atoi(getEnv("PARKING_FLOORS", "3"))
	if err != nil {
		floors = 3
	}

	rows, err := strconv.Atoi(getEnv("PARKING_ROWS", "5"))
	if err != nil {
		rows = 5
	}

	columns, err := strconv.Atoi(getEnv("PARKING_COLUMNS", "5"))
	if err != nil {
		columns = 5
	}

	// get floor vehicle map
	floorVehicleMap := make(map[int]domain.VehicleType)
	for f := 1; f <= floors; f++ {
		vehicleType := domain.VehicleType(getEnv(fmt.Sprintf("PARKING_FLOOR_%v_VEHICLE_TYPE", f), string(domain.Car)))
		if vehicleType.IsValid() {
			floorVehicleMap[f] = vehicleType
		} else {
			log.Printf("Warning: invalid vehicle type for floor %v, using default vehicle type %v\n", f, domain.Car)
			floorVehicleMap[f] = domain.Car
		}
	}

	return ParkingConfig{
		Floors:          floors,
		Rows:            rows,
		Columns:         columns,
		FloorVehicleMap: floorVehicleMap,
	}
}

func getServerConfig() ServerConfig {
	return ServerConfig{
		Port: getEnv("PORT", "8080"),
	}
}
