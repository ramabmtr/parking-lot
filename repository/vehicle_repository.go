package repository

import (
	"database/sql"
	"errors"
	"sync"
	"time"

	"parking-lot/domain"
)

type vehicleRepo struct {
	db    *sql.DB
	mutex *sync.RWMutex
}

func NewVehicleRepository(db *sql.DB) domain.VehicleRepository {
	return &vehicleRepo{
		db:    db,
		mutex: &sync.RWMutex{},
	}
}

func (r *vehicleRepo) GetVehicleByLicensePlate(licensePlate string) (*domain.Vehicle, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	query := `
		SELECT id, license_plate, type, created_at, updated_at
		FROM vehicles
		WHERE license_plate = $1
	`

	var vehicle domain.Vehicle
	err := r.db.QueryRow(query, licensePlate).Scan(
		&vehicle.ID,
		&vehicle.LicensePlate,
		&vehicle.Type,
		&vehicle.CreatedAt,
		&vehicle.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &vehicle, nil
}

func (r *vehicleRepo) CreateVehicle(vehicle *domain.Vehicle) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	query := `
		INSERT INTO vehicles (license_plate, type, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	now := time.Now()
	err := r.db.QueryRow(
		query,
		vehicle.LicensePlate,
		vehicle.Type,
		now,
		now,
	).Scan(&vehicle.ID)

	if err != nil {
		return err
	}

	vehicle.CreatedAt = now
	vehicle.UpdatedAt = now

	return nil
}
