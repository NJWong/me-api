package api

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/njwong/me-api/database"
	"github.com/njwong/me-api/models"
)

func AddSpeciesEndpoints(app *fiber.App) {
	apiGroup := app.Group("/api")
	apiGroup.Get("/species", handleGetSpecies)
	apiGroup.Get("/species/:id", handleGetSpeciesById)
}

func AddAdminSpeciesEndpoints(app *fiber.App) {
	apiGroup := app.Group("/api")
	apiGroup.Post("/species", handleCreateSpecies)
	apiGroup.Put("/species/:id", handleUpdateSpecies)
	apiGroup.Delete("/species/:id", handleDeleteSpeciesById)
}

func handleGetSpecies(c *fiber.Ctx) error {
	db := database.Client

	query := "SELECT * FROM species"

	res, err := db.Query(query)

	if err != nil {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "Internal server error",
		})
	}

	defer res.Close()

	speciesList := []models.Species{}

	for res.Next() {
		var species models.Species

		err := res.Scan(&species.ID, &species.Name)

		if err != nil {
			fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"msg": "Internal server error",
			})
		}

		speciesList = append(speciesList, species)
	}

	return c.JSON(speciesList)
}

func handleGetSpeciesById(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")

	if err != nil {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": "Bad request - invalid id",
		})
	}

	species, err := findSpeciesByID(id)

	if species == nil || err != nil {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"msg": "Species not found",
		})
	}

	return c.JSON(species)
}

func findSpeciesByID(id int) (*models.Species, error) {
	db := database.Client

	var species models.Species

	query := fmt.Sprintf("SELECT * FROM species WHERE id = %d", id)
	err := db.QueryRow(query).Scan(&species.ID, &species.Name)

	return &species, err
}

func handleCreateSpecies(c *fiber.Ctx) error {
	var species models.Species

	err := c.BodyParser(&species)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	db := database.Client

	query := fmt.Sprintf("INSERT INTO species (name) VALUES (\"%s\")", species.Name)

	result, err := db.Exec(query)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "Failed to create species",
		})
	}

	id, err := result.LastInsertId()

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "Failed to get species ID",
		})
	}

	species.ID = int(id)
	return c.Status(fiber.StatusCreated).JSON(species)
}

func handleUpdateSpecies(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")

	if err != nil {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": "Bad request - invalid id",
		})
	}

	var species models.Species
	err = c.BodyParser(&species)

	if err != nil {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": "Bad request - invalid data",
		})
	}

	db := database.Client

	query := fmt.Sprintf("UPDATE species SET name = '%s' WHERE id = %d", species.Name, id)

	result, err := db.Exec(query)

	if err != nil {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "Failed to update species",
		})
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "Failed to update species",
		})
	}

	if rowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"msg": "Species not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"msg": "Species updated"})
}

func handleDeleteSpeciesById(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")

	if err != nil {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": "Bad request - invalid id",
		})
	}

	db := database.Client

	query := fmt.Sprintf("DELETE FROM species WHERE id = %d", id)

	result, err := db.Exec(query)

	if err != nil {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "Failed to delete species",
		})
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "Failed to delete species",
		})
	}

	if rowsAffected == 0 {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"msg": "Species not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"msg": "Species deleted"})
}
