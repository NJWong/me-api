package api

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/njwong/me-api/database"
	"github.com/njwong/me-api/models"
)

func AddCharacterEndpoints(app *fiber.App) {
	app.Get("/characters", handleGetCharacters)
	app.Get("/characters/:id", handleGetCharacter)
}

func handleGetCharacters(c *fiber.Ctx) error {
	db := database.Client

	query := "SELECT * FROM characters"

	res, err := db.Query(query)

	if err != nil {
		log.Fatal("(getCharacters) db.Query", err)
	}

	defer res.Close()

	characters := []models.Character{}

	for res.Next() {
		var c models.Character

		err := res.Scan(&c.ID, &c.Name, &c.Species, &c.Gender, &c.Class)

		if err != nil {
			log.Fatal("(getCharacters) res.scan", err)
		}

		characters = append(characters, c)
	}

	return c.JSON(characters)
}

func handleGetCharacter(c *fiber.Ctx) error {
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

func findCharacterByID(id int) *models.Character {
	db := database.Client

	var character models.Character

	query := fmt.Sprintf("SELECT * FROM characters WHERE id = %d", id)
	err := db.QueryRow(query).Scan(&character.ID, &character.Name, &character.Species, &character.Gender, &character.Class)

	if err != nil {
		log.Fatal("(getCharacter) db.Query", err)
	}

	return &character
}
