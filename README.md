# Parking Lot System

A parking lot management system built with Go, Echo framework, and PostgreSQL.

## Features

- Multiple floors for different vehicle types (motorcycle, bicycle, car)
- Parking spots arranged in rows and columns
- Ability to park and unpark vehicles
- Check available parking spots
- Search for vehicles by license plate
- Configurable number of floors, rows, and columns
- Concurrent access handling for multiple gates

### Concurrency Handling

The system uses mutex locks to handle concurrent access from multiple gates, ensuring that:

- Two vehicles cannot be assigned the same parking spot
- Vehicle status is accurately tracked during parking and unparking operations

## API Endpoints

- `POST /park`: Park a vehicle
- `POST /unpark`: Unpark a vehicle
- `GET /available`: Get available parking spots
- `GET /search`: Search for a vehicle by license plate

## Configuration

The system can be configured using environment variables:

- `PARKING_FLOORS`: Number of floors in the parking lot (default: 3)
- `PARKING_ROWS`: Number of rows per floor (default: 5)
- `PARKING_COLUMNS`: Number of columns per floor (default: 5)
- `PARKING_FLOOR_X_VEHICLE_TYPE`: Vehicle type for `X` floor (default: car)
- `DB_HOST`: Database host (default: localhost)
- `DB_PORT`: Database port (default: 5432)
- `DB_USER`: Database user (default: postgres)
- `DB_PASSWORD`: Database password (default: postgres)
- `DB_NAME`: Database name (default: parking_lot)
- `DB_SSLMODE`: Database SSL mode (default: disable)
- `PORT`: Server port (default: 8080)

Note: if parking configuration is changed, you must rerun the migrations.

## Getting Started

### Prerequisites

- Go 1.24 or higher
- PostgreSQL

### Installation

1. Clone the repository
2. Create a PostgreSQL database named `parking_lot`
3. Copy `.env.example` to `.env` and configure the environment variables
4. Run database migrations:

```bash
go run cmd/migrate/main.go
```

Note: If you change parking lot configuration, you must truncate the `parking_spots` and `parking_record` table
and run the migrations again.

5. Run the application:

```bash
go run main.go
```

## API Usage Examples

### Park a Vehicle

```bash
curl -X POST http://localhost:8080/park \
  -H "Content-Type: application/json" \
  -d '{"license_plate": "ABC123", "vehicle_type": "car"}'
```

### Unpark a Vehicle

```bash
curl -X POST http://localhost:8080/unpark \
  -H "Content-Type: application/json" \
  -d '{"license_plate": "ABC123"}'
```

### Get Available Spots

```bash
curl -X GET http://localhost:8080/available
```

### Search for a Vehicle

```bash
curl -X GET http://localhost:8080/search?license_plate=ABC123
```
