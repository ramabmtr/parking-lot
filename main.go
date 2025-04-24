package main

import (
	"fmt"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"parking-lot/config"
	"parking-lot/handler"
	"parking-lot/repository"
	"parking-lot/service"
)

func main() {
	// Get application configuration
	appConfig := config.GetAppConfig()

	// Initialize database connection
	db, err := config.InitDBConnection()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	parkingRepo := repository.NewParkingRepository(db)
	vehicleRepo := repository.NewVehicleRepository(db)

	parkingService := service.NewParkingService(parkingRepo, vehicleRepo)
	parkingHandler := handler.NewParkingHandler(parkingService)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Routes
	e.POST("/park", parkingHandler.ParkVehicle)
	e.POST("/unpark", parkingHandler.UnparkVehicle)
	e.GET("/available", parkingHandler.GetAvailableSpots)
	e.GET("/search", parkingHandler.SearchVehicle)

	// Start server
	port := appConfig.Server.Port
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", port)))
}
