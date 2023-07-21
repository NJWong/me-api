package api

import (
	"database/sql"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/njwong/me-api/database"
	"github.com/njwong/me-api/models"
)

func AddCharactersRoutes(app *fiber.App) {
	apiGroup := app.Group("/api")
	apiGroup.Get("/characters", handleGetCharacters)
	apiGroup.Get("/characters/:id", handleGetCharacter)
}

func AddAdminCharacterRoutes(app *fiber.App) {
	apiGroup := app.Group("/api")
	apiGroup.Post("/characters", handleCreateCharacter)
	apiGroup.Put("/characters/:id", handleUpdateCharacter)
	apiGroup.Delete("/characters/:id", handleDeleteCharacterById)
}

func handleGetCharacters(c *fiber.Ctx) error {
	db := database.Client

	query := "SELECT characters.id, characters.name, characters.class, species.id, species.name, genders.id, genders.name FROM characters LEFT JOIN species ON characters.species = species.id LEFT JOIN genders ON characters.gender = genders.id"

	res, err := db.Query(query)

	if err != nil {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "Internal server error",
		})
	}

	defer res.Close()

	characters := []models.CharacterObject{}

	for res.Next() {
		var character models.CharacterObject
		var speciesID sql.NullInt64
		var speciesName sql.NullString
		var genderID sql.NullInt64
		var genderName sql.NullString

		err := res.Scan(&character.ID, &character.Name, &character.Class, &speciesID, &speciesName, &genderID, &genderName)

		if err != nil {
			fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"msg": "Internal server error",
			})
		}

		if speciesID.Valid {
			character.Species = &models.SpeciesObject{
				ID:   int(speciesID.Int64),
				Name: speciesName.String,
				URL:  fmt.Sprintf("https://me-api.fly.dev/api/species/%d", speciesID.Int64),
			}
		}

		if genderID.Valid {
			character.Gender = &models.GenderObject{
				ID:   int(genderID.Int64),
				Name: genderName.String,
				URL:  fmt.Sprintf("https://me-api.fly.dev/api/genders/%d", genderID.Int64),
			}
		}

		characters = append(characters, character)
	}

	return c.JSON(characters)
}

func handleGetCharacter(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")

	if err != nil {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": "Bad request - invalid id",
		})
	}

	character, err := findCharacterByID(id)

	if character == nil || err != nil {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"msg": "Character not found",
		})
	}

	return c.JSON(character)
}

func findCharacterByID(id int) (*models.Character, error) {
	db := database.Client

	var character models.Character

	// TODO - LEFT JOIN to populate the gender and species fields with data
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

func handleDeleteCharacterById(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")

	if err != nil {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": "Bad request - invalid id",
		})
	}

	db := database.Client

	query := fmt.Sprintf("DELETE FROM characters WHERE id = %d", id)

	result, err := db.Exec(query)

	if err != nil {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "Failed to delete character",
		})
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "Failed to delete character",
		})
	}

	if rowsAffected == 0 {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"msg": "Character not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"msg": "Character deleted"})
}

func handleUpdateCharacter(c *fiber.Ctx) error {
	// Get the id from params
	id, err := c.ParamsInt("id")

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": "Bad request - invalid id",
		})
	}

	var character models.Character
	err = c.BodyParser(&character)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": "Bad request - invalid data",
		})
	}

	db := database.Client

	// Using parameterized queries to avoid escaping special characters like `'`
	query := "UPDATE characters SET name = ?, species = ?, gender = ?, class = ? WHERE id = ?"

	stmt, err := db.Prepare(query)

	if err != nil {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "Internal server error",
		})
	}

	defer stmt.Close()

	_, err = stmt.Exec(character.Name, character.Species, character.Gender, character.Class, id)

	if err != nil {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "Internal server error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"msg": "Character updated"})
}
