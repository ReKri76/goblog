package register

import (
	"crypto/rsa"
	"database/sql"
	"errors"
	"goblog/keys"
	"goblog/service"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func Refresh(db *sql.DB, private *rsa.PrivateKey, public *rsa.PublicKey) fiber.Handler {
	return func(c *fiber.Ctx) error {
		type refresher struct {
			keys.Record
			RefreshTime int64
		}

		refresh := c.Cookies("refresh_token")
		token, err := jwt.Parse(refresh, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, errors.New("deer_penis")
			}
			return public, nil
		})
		if err != nil || token == nil || !token.Valid {
			return c.Status(401).SendString("Invalid token")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(401).SendString("Invalid token")
		}

		err = service.RefresherService(db, claims["mail"].(string))
		if err != nil {
			return c.Status(401).SendString("Invalid token")
		}

		var data keys.Record
		data.Mail = claims["mail"].(string)
		data.Role = claims["role"].(string)

		access, err := data.CreateJWT(2, private)
		if err != nil {
			return err
		}

		refresh, err = data.CreateJWT(24*7, private)
		if err != nil {
			return err
		}

		c.Cookie(&fiber.Cookie{
			Name:     "refresh_token",
			Value:    refresh,
			Expires:  time.Now().Add(24 * 7 * time.Hour),
			HTTPOnly: true,
			Secure:   true,
			SameSite: "Lax",
		})

		return c.Status(200).JSON(fiber.Map{
			"message":      "Refresh successful",
			"access_token": access,
		})
	}
}
