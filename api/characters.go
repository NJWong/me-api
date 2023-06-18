package api

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/njwong/me-api/database"
	"github.com/njwong/me-api/models"
)

func AddCharactersRoutes(app *fiber.App) {
	apiGroup := app.Group("/api")
	apiGroup.Get("/characters", handleGetCharacters)
	apiGroup.Get("/characters/:id", handleGetCharacter)
	apiGroup.Post("/characters", handleCreateCharacter)
}

func handleGetCharacters(c *fiber.Ctx) error {
	db := database.Client

	query := "SELECT * FROM characters"

	res, err := db.Query(query)

	if err != nil {
		log.Fatal("(getCharacters) db.Query - ", err)
	}

	defer res.Close()

	characters := []models.Character{}

	for res.Next() {
		var character models.Character

		err := res.Scan(&character.ID, &character.Name, &character.Species, &character.Gender, &character.Class)

		if err != nil {
			log.Fatal("(getCharacters) res.Scan - ", err)
		}

		characters = append(characters, character)
	}

	return c.JSON(characters)
}

func handleGetCharacter(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid character ID",
		})
	}

	character, err := findCharacterByID(id)

	if err != nil {
		log.Fatal("(handleGetCharacter) findCharacterByID - ", err)
	}

	if character == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Character not found",
		})
	}

	return c.JSON(character)
}

func findCharacterByID(id int) (*models.Character, error) {
	db := database.Client

	var character models.Character

	query := fmt.Sprintf("SELECT * FROM characters WHERE id = %d", id)
	err := db.QueryRow(query).Scan(&character.ID, &character.Name, &character.Species, &character.Gender, &character.Class)

	return &character, err
}

func handleCreateCharacter(c *fiber.Ctx) error {
	var character models.Character

	err := c.BodyParser(&character)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	db := database.Client

	query := fmt.Sprintf("INSERT INTO characters (name, species, gender, class) VALUES (\"%s\", %d, %d, \"%s\")", character.Name, character.Species, character.Gender, character.Class)

	result, err := db.Exec(query)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to create character",
		})
	}

	id, err := result.LastInsertId()

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get character ID",
		})
	}

	character.ID = int(id)
	return c.Status(fiber.StatusCreated).JSON(character)
}
