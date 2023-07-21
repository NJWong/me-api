package api

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/njwong/me-api/database"
	"github.com/njwong/me-api/models"
)

func AddGendersEndpoints(app *fiber.App) {
	apiGroup := app.Group("/api")

	apiGroup.Get("/genders", handleGetGenders)
	apiGroup.Get("/genders/:id", handleGetGender)
}

func AddAdminGendersEndpoints(app *fiber.App) {
	apiGroup := app.Group("/api")

	apiGroup.Post("/genders", handleCreateGender)
	apiGroup.Put("/genders/:id", handleUpdateGender)
	apiGroup.Delete("/genders/:id", handleDeleteGender)
}

func handleGetGenders(c *fiber.Ctx) error {
	db := database.Client

	query := "SELECT * FROM genders"

	res, err := db.Query(query)

	if err != nil {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "Internal server error",
		})
	}

	defer res.Close()

	genders := []models.Gender{}

	for res.Next() {
		var gender models.Gender

		err := res.Scan(&gender.ID, &gender.Name)

		if err != nil {
			fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"msg": "Internal server error",
			})
		}

		genders = append(genders, gender)
	}

	return c.JSON(genders)
}

func handleGetGender(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")

	if err != nil {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": "Bad request - invalid id",
		})
	}

	gender, err := findGenderByID(id)

	if gender == nil || err != nil {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"msg": "Gender not found",
		})
	}

	return c.JSON(gender)
}

func findGenderByID(id int) (*models.Gender, error) {
	db := database.Client

	var gender models.Gender

	query := fmt.Sprintf("SELECT * FROM genders WHERE id = %d", id)
	err := db.QueryRow(query).Scan(&gender.ID, &gender.Name)

	return &gender, err
}

func handleCreateGender(c *fiber.Ctx) error {
	var gender models.Gender

	err := c.BodyParser(&gender)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	db := database.Client

	query := fmt.Sprintf("INSERT INTO genders (name) VALUES (\"%s\")", gender.Name)

	result, err := db.Exec(query)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create gender",
		})
	}

	id, err := result.LastInsertId()

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get gender ID",
		})
	}

	gender.ID = int(id)
	return c.Status(fiber.StatusCreated).JSON(gender)
}

func handleUpdateGender(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")

	if err != nil {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": "Bad request - invalid id",
		})
	}

	var gender models.Gender
	err = c.BodyParser(&gender)

	if err != nil {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": "Bad request - invalid data",
		})
	}

	db := database.Client

	query := fmt.Sprintf("UPDATE genders SET name = '%s' WHERE id = %d", gender.Name, id)

	result, err := db.Exec(query)

	if err != nil {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "Failed to update gender",
		})
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "Failed to update gender",
		})
	}

	if rowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"msg": "Gender not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"msg": "Gender updated"})
}

func handleDeleteGender(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")

	if err != nil {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": "Bad request - invalid id",
		})
	}

	db := database.Client

	query := fmt.Sprintf("DELETE FROM genders WHERE id = %d", id)

	result, err := db.Exec(query)

	if err != nil {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "Failed to delete gender",
		})
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "Failed to delete gender",
		})
	}

	if rowsAffected == 0 {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"msg": "Gender not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"msg": "Gender deleted"})
}
