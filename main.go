package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	"github.com/njwong/me-api/handlers"
)

type Character struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Species int    `json:"species"`
	Gender  int    `json:"gender"`
	Class   string `json:"class"`
}

var db *sql.DB

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load env", err)
	}

	// Open a connection to the database
	db, err = sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		log.Fatal("Failed to open db connection", err)
	}

	// Create app and define routes
	app := fiber.New()

	app.Get("/health", handlers.HandleHealthCheck)

	app.Get("/characters", getCharacters)
	app.Get("/characters/:id", getCharacter)

	// Get the port from the environment
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	// Run the app listening on the selected port
	log.Fatal(app.Listen("0.0.0.0:" + port))
}

func getCharacters(c *fiber.Ctx) error {
	query := "SELECT * FROM characters"

	res, err := db.Query(query)

	if err != nil {
		log.Fatal("(getCharacters) db.Query", err)
	}

	defer res.Close()

	characters := []Character{}

	for res.Next() {
		var c Character

		err := res.Scan(&c.ID, &c.Name, &c.Species, &c.Gender, &c.Class)

		if err != nil {
			log.Fatal("(getCharacters) res.scan", err)
		}

		characters = append(characters, c)
	}

	return c.JSON(characters)
}

func getCharacter(c *fiber.Ctx) error {
	id := getCharacterID(c)

	if id == -1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid character ID",
		})
	}

	character := findCharacterByID(id)

	if character == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Character not found",
		})
	}

	return c.JSON(character)
}

func getCharacterID(c *fiber.Ctx) int {
	id, err := c.ParamsInt("id")

	if err != nil {
		return -1
	}

	return id
}

func findCharacterByID(id int) *Character {
	var character Character

	query := fmt.Sprintf("SELECT * FROM characters WHERE id = %d", id)
	err := db.QueryRow(query).Scan(&character.ID, &character.Name, &character.Species, &character.Gender, &character.Class)

	if err != nil {
		log.Fatal("(getCharacter) db.Query", err)
	}

	return &character
}
