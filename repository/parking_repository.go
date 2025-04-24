package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/lib/pq"
	"parking-lot/domain"
)

type parkingRepo struct {
	db    *sql.DB
	mutex *sync.RWMutex
}

func NewParkingRepository(db *sql.DB) domain.ParkingRepository {
	return &parkingRepo{
		db:    db,
		mutex: &sync.RWMutex{},
	}
}

func (r *parkingRepo) GetAvailableSpots(floors ...int) ([]domain.ParkingSpot, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	floorWhere := ""
	if len(floors) > 0 {
		floorWhere = "AND ps.floor = ANY($1)"
	}

	query := fmt.Sprintf(`
		SELECT ps.id, ps.floor, ps.row, ps.column, ps.is_active, ps.created_at, ps.updated_at
		FROM parking_spots ps
		LEFT JOIN (
			SELECT parking_spot_id
			FROM parking_records
			WHERE exit_time IS NULL
		) pr ON ps.id = pr.parking_spot_id
		WHERE ps.is_active = true AND pr.parking_spot_id IS NULL %s
		ORDER BY ps.floor, ps.row, ps.column
	`, floorWhere)

	var args []any
	if floors != nil && len(floors) > 0 {
		args = append(args, pq.Array(floors))
	}
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var spots []domain.ParkingSpot
	for rows.Next() {
		var spot domain.ParkingSpot
		err := rows.Scan(
			&spot.ID,
			&spot.Floor,
			&spot.Row,
			&spot.Column,
			&spot.IsActive,
			&spot.CreatedAt,
			&spot.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		spots = append(spots, spot)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return spots, nil
}

func (r *parkingRepo) GetSpotByID(id int64) (*domain.ParkingSpot, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	query := `
		SELECT id, floor, row, "column", is_active, created_at, updated_at
		FROM parking_spots
		WHERE id = $1
	`

	var spot domain.ParkingSpot
	err := r.db.QueryRow(query, id).Scan(
		&spot.ID,
		&spot.Floor,
		&spot.Row,
		&spot.Column,
		&spot.IsActive,
		&spot.CreatedAt,
		&spot.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &spot, nil
}

func (r *parkingRepo) GetSpotByPosition(floor, row, column int) (*domain.ParkingSpot, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	query := `
		SELECT id, floor, row, "column", is_active, created_at, updated_at
		FROM parking_spots
		WHERE floor = $1 AND row = $2 AND "column" = $3
	`

	var spot domain.ParkingSpot
	err := r.db.QueryRow(query, floor, row, column).Scan(
		&spot.ID,
		&spot.Floor,
		&spot.Row,
		&spot.Column,
		&spot.IsActive,
		&spot.CreatedAt,
		&spot.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &spot, nil
}

func (r *parkingRepo) UpdateSpotStatus(id int64, isActive bool) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	query := `
		UPDATE parking_spots
		SET is_active = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(query, isActive, time.Now(), id)
	return err
}

func (r *parkingRepo) CreateParkingRecord(record *domain.ParkingRecord) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	query := `
		INSERT INTO parking_records (vehicle_id, parking_spot_id, entry_time, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	now := time.Now()
	err := r.db.QueryRow(
		query,
		record.VehicleID,
		record.ParkingSpotID,
		record.EntryTime,
		now,
		now,
	).Scan(&record.ID)

	return err
}

func (r *parkingRepo) UpdateParkingRecord(record *domain.ParkingRecord) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	query := `
		UPDATE parking_records
		SET exit_time = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(query, record.ExitTime, time.Now(), record.ID)
	return err
}

func (r *parkingRepo) GetLastParkingRecordByVehicleID(vehicleID int64) (*domain.ParkingRecord, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	query := `
		SELECT id, vehicle_id, parking_spot_id, entry_time, exit_time, created_at, updated_at
		FROM parking_records
		WHERE vehicle_id = $1
		ORDER BY entry_time DESC
		LIMIT 1
	`

	var record domain.ParkingRecord
	err := r.db.QueryRow(query, vehicleID).Scan(
		&record.ID,
		&record.VehicleID,
		&record.ParkingSpotID,
		&record.EntryTime,
		&record.ExitTime,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &record, nil
}
