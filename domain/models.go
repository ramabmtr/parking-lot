package domain

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type VehicleType string

const (
	Motorcycle VehicleType = "motorcycle"
	Bicycle    VehicleType = "bicycle"
	Car        VehicleType = "car"
)

func (t VehicleType) IsValid() bool {
	return t == Motorcycle || t == Bicycle || t == Car
}

type Vehicle struct {
	ID           int64       `json:"id"`
	LicensePlate string      `json:"license_plate"`
	Type         VehicleType `json:"type"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

type ParkingSpot struct {
	ID        int64     `json:"id"`
	Floor     int       `json:"floor"`
	Row       int       `json:"row"`
	Column    int       `json:"column"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (p ParkingSpot) MarshalJSON() ([]byte, error) {
	type Alias ParkingSpot

	return json.Marshal(&struct {
		Alias
		SpotID string `json:"spot_id"`
	}{
		Alias:  Alias(p),
		SpotID: fmt.Sprintf("%v-%v-%v", p.Floor, p.Row, p.Column),
	})
}

type ParkingRecord struct {
	ID            int64        `json:"id"`
	VehicleID     int64        `json:"vehicle_id"`
	ParkingSpotID int64        `json:"parking_spot_id"`
	EntryTime     time.Time    `json:"entry_time"`
	ExitTime      sql.NullTime `json:"exit_time"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

func (p ParkingRecord) IsParked() bool {
	return !p.ExitTime.Valid
}

// ParkingRepository defines the interface for parking spot operations
type ParkingRepository interface {
	GetAvailableSpots(floors ...int) ([]ParkingSpot, error)
	GetSpotByID(id int64) (*ParkingSpot, error)
	GetSpotByPosition(floor, row, column int) (*ParkingSpot, error)
	UpdateSpotStatus(id int64, isActive bool) error
	CreateParkingRecord(record *ParkingRecord) error
	UpdateParkingRecord(record *ParkingRecord) error
	GetLastParkingRecordByVehicleID(vehicleID int64) (*ParkingRecord, error)
}

// VehicleRepository defines the interface for vehicle operations
type VehicleRepository interface {
	GetVehicleByLicensePlate(licensePlate string) (*Vehicle, error)
	CreateVehicle(vehicle *Vehicle) error
}

// ParkingService defines the interface for parking business logic
type ParkingService interface {
	ParkVehicle(licensePlate string, vehicleType VehicleType) (*ParkingSpot, error)
	UnparkVehicle(licensePlate string) error
	GetAllAvailableSpots() ([]ParkingSpot, error)
	SearchVehicle(licensePlate string) (*ParkingSpot, bool, error)
}

type ParkRequest struct {
	LicensePlate string      `json:"license_plate"`
	VehicleType  VehicleType `json:"vehicle_type"`
}

type ParkResponse struct {
	Success     bool         `json:"success"`
	Message     string       `json:"message"`
	ParkingSpot *ParkingSpot `json:"parking_spot,omitempty"`
}

type UnparkRequest struct {
	LicensePlate string `json:"license_plate"`
}

type UnparkResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type SearchResponse struct {
	Success     bool         `json:"success"`
	Message     string       `json:"message"`
	ParkingSpot *ParkingSpot `json:"parking_spot,omitempty"`
	IsParked    bool         `json:"is_parked"`
}

type AvailableSpotsResponse struct {
	Success      bool                 `json:"success"`
	Message      string               `json:"message"`
	ParkingSpots ParkingSpotByVehicle `json:"parking_spots,omitempty"`
}

type ParkingSpotByVehicle struct {
	Car        []ParkingSpot `json:"car"`
	Motorcycle []ParkingSpot `json:"motorcycle"`
	Bicycle    []ParkingSpot `json:"bicycle"`
}
