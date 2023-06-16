package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
)

type Character struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Species int `json:"species"`
	Gender int `json:"gender"`
	Class string `json:"class"`
}

var characters = []Character{
	{
		ID: 1,
		Name: "Commander Shepard",
		Species: 1,
		Gender: 1,
		Class: "N7 Alliance Marine / Spectre",
	},
	{
		ID: 2,
		Name: "Liara T'Soni",
		Species: 2,
		Gender: 3,
		Class: "Asari Scientist",
	},
}

func main() {
	app := fiber.New()

	app.Get("/characters", getCharacters)
	app.Get("/characters/:id", getCharacter)

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	log.Fatal(app.Listen("0.0.0.0:" + port))
}

func getCharacters(c *fiber.Ctx) error {
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
	for _, character := range characters {
		if character.ID == id {
			return &character
		}
	}

	return nil
}