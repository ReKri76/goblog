package register

import (
	"crypto/rsa"
	"database/sql"
	"errors"
	"goblog/keys"
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

		query := "SELECT RefreshTime FROM users WHERE Mail = $1"
		var data refresher
		err = db.QueryRow(query, claims["mail"]).Scan(&data.RefreshTime)
		if err != nil {
			return err
		}
		if data.RefreshTime < time.Now().Unix() {
			return c.Status(401).SendString("Invalid token")
		}

		access, err := data.CreateJWT(2, private)
		if err != nil {
			return err
		}

		refresh, err = data.CreateJWT(24*7, private)
		if err != nil {
			return err
		}

		c.Cookie(&fiber.Cookie{
			Name:    "refresh_token",
			Value:   refresh,
			Expires: time.Now().Add(24 * 7 * time.Hour),
		})

		return c.Status(200).JSON(fiber.Map{
			"message":      "Refresh successful",
			"access_token": access,
		})
	}
}
