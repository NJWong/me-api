package api

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/njwong/me-api/database"
	"github.com/njwong/me-api/models"
)

func authMiddleware(c *fiber.Ctx) error {
	// Get the JWT token from the Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"msg": "Missing Authorization header",
		})
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Check token is signed using the correct method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("token not signed using RSA")
		}

		// Get the public key from Auth0
		resp, err := http.Get("https://me-api.au.auth0.com/.well-known/jwks.json")
		if err != nil {
			return nil, fmt.Errorf("failed to get public key: %v", err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %v", err)
		}
		var jwks struct {
			Keys []struct {
				Kty string `json:"kty"`
				Kid string `json:"kid"`
				Use string `json:"use"`
				N   string `json:"n"`
				E   string `json:"e"`
			} `json:"keys"`
		}
		if err := json.Unmarshal(body, &jwks); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
		}
		var key *rsa.PublicKey
		for _, k := range jwks.Keys {
			if k.Kid == token.Header["kid"] && k.Kty == "RSA" && k.Use == "sig" {
				nb, err := base64.RawURLEncoding.DecodeString(k.N)
				if err != nil {
					return nil, fmt.Errorf("failed to decode public key modulus: %v", err)
				}
				eb, err := base64.RawURLEncoding.DecodeString(k.E)
				if err != nil {
					return nil, fmt.Errorf("failed to decode public key exponent: %v", err)
				}
				key = &rsa.PublicKey{
					N: big.NewInt(0).SetBytes(nb),
					E: int(big.NewInt(0).SetBytes(eb).Int64()),
				}
				break
			}
		}
		if key == nil {
			return nil, errors.New("public key not found")
		}
		return key, nil
	})

	const invalidMsg = "Invalid token"

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"msg": invalidMsg,
		})
	}

	if !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"msg": invalidMsg,
		})
	}

	claims := token.Claims.(jwt.MapClaims)
	audience, err := claims.GetAudience()

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"msg": invalidMsg,
		})
	}

	if contains(audience, "https://me-api.fly.dev/api") {
		// Call the next middleware function
		return c.Next()
	} else {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"msg": invalidMsg,
		})
	}
}

func contains(arr []string, target string) bool {
	for _, s := range arr {
		if s == target {
			return true
		}
	}
	return false
}

func AddSpeciesEndpoints(app *fiber.App) {
	apiGroup := app.Group("/api")

	// Public endpoints
	apiGroup.Get("/species", handleGetSpecies)
	apiGroup.Get("/species/:id", handleGetSpeciesById)

	// Protected endpoints
	apiGroup.Use(authMiddleware)
	apiGroup.Post("/species", handleCreateSpecies)
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
			"message": "Failed to create species",
		})
	}

	id, err := result.LastInsertId()

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get species ID",
		})
	}

	species.ID = int(id)
	return c.Status(fiber.StatusCreated).JSON(species)
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
			"message": "Failed to delete species",
		})
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to delete species",
		})
	}

	if rowsAffected == 0 {
		fmt.Printf("Error - \"%s\" for the following request:\n", err.Error())

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Species not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Species deleted"})
}
