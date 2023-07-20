package middleware

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
)

func JWTAuth(c *fiber.Ctx) error {
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
