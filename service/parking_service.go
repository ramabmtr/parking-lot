package service

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"parking-lot/config"
	"parking-lot/domain"
)

type parkingService struct {
	parkingRepo domain.ParkingRepository
	vehicleRepo domain.VehicleRepository
	mutex       *sync.Mutex
}

func NewParkingService(
	parkingRepo domain.ParkingRepository,
	vehicleRepo domain.VehicleRepository,
) domain.ParkingService {
	return &parkingService{
		parkingRepo: parkingRepo,
		vehicleRepo: vehicleRepo,
		mutex:       &sync.Mutex{},
	}
}

func (s *parkingService) ParkVehicle(licensePlate string, vehicleType domain.VehicleType) (*domain.ParkingSpot, error) {
	// Check if the vehicle is already parked
	vehicle, err := s.vehicleRepo.GetVehicleByLicensePlate(licensePlate)
	if err != nil {
		return nil, fmt.Errorf("error getting vehicle: %w", err)
	}

	// Create vehicle if it doesn't exist
	if vehicle == nil {
		vehicle = &domain.Vehicle{
			LicensePlate: licensePlate,
			Type:         vehicleType,
		}
		err = s.vehicleRepo.CreateVehicle(vehicle)
		if err != nil {
			return nil, fmt.Errorf("error creating vehicle: %w", err)
		}
	}

	// Check if the vehicle is already parked
	lastRecord, err := s.parkingRepo.GetLastParkingRecordByVehicleID(vehicle.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting last parking record: %w", err)
	}

	if lastRecord.IsParked() {
		spot, err := s.parkingRepo.GetSpotByID(lastRecord.ParkingSpotID)
		if err != nil {
			return nil, fmt.Errorf("error getting parking spot: %w", err)
		}
		return nil, fmt.Errorf("vehicle is already parked at spot %d-%d-%d", spot.Floor, spot.Row, spot.Column)
	}

	// Use mutex to prevent race conditions when finding available spots
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// get floors based on vehicle type
	var floors []int
	parkingConfig := config.GetAppConfig().Parking
	for floor, t := range parkingConfig.FloorVehicleMap {
		if t == vehicleType {
			floors = append(floors, floor)
		}
	}

	// Get available spots based on floor(s)
	availableSpots, err := s.parkingRepo.GetAvailableSpots(floors...)
	if err != nil {
		return nil, fmt.Errorf("error getting available spots: %w", err)
	}

	if availableSpots == nil || len(availableSpots) == 0 {
		return nil, errors.New("no available parking spots")
	}

	record := &domain.ParkingRecord{
		VehicleID:     vehicle.ID,
		ParkingSpotID: availableSpots[0].ID,
		EntryTime:     time.Now(),
	}

	err = s.parkingRepo.CreateParkingRecord(record)
	if err != nil {
		return nil, fmt.Errorf("error creating parking record: %w", err)
	}

	return &availableSpots[0], nil
}

func (s *parkingService) UnparkVehicle(licensePlate string) error {
	// Get vehicle
	vehicle, err := s.vehicleRepo.GetVehicleByLicensePlate(licensePlate)
	if err != nil {
		return fmt.Errorf("error getting vehicle: %w", err)
	}

	if vehicle == nil {
		return fmt.Errorf("vehicle with license plate %s not found", licensePlate)
	}

	// Check if the vehicle is parked
	lastRecord, err := s.parkingRepo.GetLastParkingRecordByVehicleID(vehicle.ID)
	if err != nil {
		return fmt.Errorf("error checking active parking: %w", err)
	}

	if !lastRecord.IsParked() {
		return fmt.Errorf("vehicle with license plate %s is not parked", licensePlate)
	}

	// Update parking record with exit time
	lastRecord.ExitTime = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}

	err = s.parkingRepo.UpdateParkingRecord(lastRecord)
	if err != nil {
		return fmt.Errorf("error updating parking record: %w", err)
	}

	return nil
}

func (s *parkingService) GetAllAvailableSpots() ([]domain.ParkingSpot, error) {
	spots, err := s.parkingRepo.GetAvailableSpots()
	if err != nil {
		return nil, fmt.Errorf("error getting available spots: %w", err)
	}
	return spots, nil
}

func (s *parkingService) SearchVehicle(licensePlate string) (*domain.ParkingSpot, bool, error) {
	vehicle, err := s.vehicleRepo.GetVehicleByLicensePlate(licensePlate)
	if err != nil {
		return nil, false, fmt.Errorf("error getting vehicle: %w", err)
	}

	if vehicle == nil {
		return nil, false, fmt.Errorf("vehicle with license plate %s not found", licensePlate)
	}

	lastRecord, err := s.parkingRepo.GetLastParkingRecordByVehicleID(vehicle.ID)
	if err != nil {
		return nil, false, fmt.Errorf("error getting last parking record: %w", err)
	}

	if lastRecord == nil {
		return nil, false, fmt.Errorf("no parking history found for vehicle with license plate %s", licensePlate)
	}

	spot, err := s.parkingRepo.GetSpotByID(lastRecord.ParkingSpotID)
	if err != nil {
		return nil, false, fmt.Errorf("error getting parking spot: %w", err)
	}

	return spot, !lastRecord.IsParked(), nil
}
