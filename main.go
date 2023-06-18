package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"

	"github.com/njwong/me-api/api"
	"github.com/njwong/me-api/database"
)

func main() {
	// Check Go env
	goEnv := os.Getenv("GO_ENV")

	// Load .env file
	if goEnv == "" || goEnv == "development" {
		if err := godotenv.Load(); err != nil {
			log.Fatal("(main) failed to load env - ", err)
		}
	}

	// Setup the connection to the database
	database.Setup()

	// Create app
	app := fiber.New()

	// Add middleware
	app.Use(logger.New())

	// Add routes
	api.AddHealthRoutes(app)
	api.AddCharactersRoutes(app)
	api.AddGendersEndpoints(app)

	// Get the port from the environment
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	// Run the app listening on the selected port
	log.Fatal(app.Listen("0.0.0.0:" + port))
}
