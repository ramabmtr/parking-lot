package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"parking-lot/config"
	"parking-lot/domain"
)

type ParkingHandler struct {
	parkingService domain.ParkingService
}

func NewParkingHandler(parkingService domain.ParkingService) *ParkingHandler {
	return &ParkingHandler{
		parkingService: parkingService,
	}
}

func (h *ParkingHandler) ParkVehicle(c echo.Context) error {
	var req domain.ParkRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, domain.ParkResponse{
			Success: false,
			Message: "Invalid request format",
		})
	}

	if req.LicensePlate == "" {
		return c.JSON(http.StatusBadRequest, domain.ParkResponse{
			Success: false,
			Message: "License plate is required",
		})
	}

	if req.VehicleType != domain.Motorcycle && req.VehicleType != domain.Bicycle && req.VehicleType != domain.Car {
		return c.JSON(http.StatusBadRequest, domain.ParkResponse{
			Success: false,
			Message: "Invalid vehicle type. Must be 'motorcycle', 'bicycle', or 'car'",
		})
	}

	spot, err := h.parkingService.ParkVehicle(req.LicensePlate, req.VehicleType)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, domain.ParkResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, domain.ParkResponse{
		Success:     true,
		Message:     "Vehicle parked successfully",
		ParkingSpot: spot,
	})
}

func (h *ParkingHandler) UnparkVehicle(c echo.Context) error {
	var req domain.UnparkRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, domain.UnparkResponse{
			Success: false,
			Message: "Invalid request format",
		})
	}

	if req.LicensePlate == "" {
		return c.JSON(http.StatusBadRequest, domain.UnparkResponse{
			Success: false,
			Message: "License plate is required",
		})
	}

	err := h.parkingService.UnparkVehicle(req.LicensePlate)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, domain.UnparkResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, domain.UnparkResponse{
		Success: true,
		Message: "Vehicle unparked successfully",
	})
}

func (h *ParkingHandler) GetAvailableSpots(c echo.Context) error {
	spots, err := h.parkingService.GetAllAvailableSpots()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, domain.AvailableSpotsResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	// group spots by vehicle type
	parkingConfig := config.GetAppConfig().Parking
	spotMap := map[domain.VehicleType][]domain.ParkingSpot{}
	for _, spot := range spots {
		vehicleType := parkingConfig.FloorVehicleMap[spot.Floor]
		spotMap[vehicleType] = append(spotMap[vehicleType], spot)
	}

	return c.JSON(http.StatusOK, domain.AvailableSpotsResponse{
		Success: true,
		Message: "Available spots retrieved successfully",
		ParkingSpots: domain.ParkingSpotByVehicle{
			Car:        spotMap[domain.Car],
			Motorcycle: spotMap[domain.Motorcycle],
			Bicycle:    spotMap[domain.Bicycle],
		},
	})
}

func (h *ParkingHandler) SearchVehicle(c echo.Context) error {
	licensePlate := c.QueryParam("license_plate")
	if licensePlate == "" {
		return c.JSON(http.StatusBadRequest, domain.SearchResponse{
			Success: false,
			Message: "License plate is required",
		})
	}

	spot, isParked, err := h.parkingService.SearchVehicle(licensePlate)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, domain.SearchResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, domain.SearchResponse{
		Success:     true,
		Message:     "Vehicle found",
		ParkingSpot: spot,
		IsParked:    isParked,
	})
}
