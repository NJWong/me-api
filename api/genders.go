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

func handleGetGenders(c *fiber.Ctx) error {
	db := database.Client

	query := "SELECT * FROM genders"

	res, err := db.Query(query)

	if err != nil {
		fmt.Print("(getGenders) db.Query - ", err)
		return err
	}

	defer res.Close()

	genders := []models.Gender{}

	for res.Next() {
		var gender models.Gender

		err := res.Scan(&gender.ID, &gender.Name)

		if err != nil {
			fmt.Print("(getGenders) res.Scan - ", err)
			return err
		}

		genders = append(genders, gender)
	}

	return c.JSON(genders)
}

func handleGetGender(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid gender ID",
		})
	}

	gender, err := findGenderByID(id)

	if err != nil {
		fmt.Print("(handleGetCharacter) findCharacterByID - ", err)
		return err
	}

	if gender == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Gender not found",
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
