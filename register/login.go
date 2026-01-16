package register

import (
	"crypto/rsa"
	"database/sql"
	"goblog/keys"
	"goblog/service"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Login(db *sql.DB, private *rsa.PrivateKey) fiber.Handler {
	return func(c *fiber.Ctx) error {
		type loginer struct {
			keys.Record
			Password string `json:"password"`
		}

		var src loginer
		if err := c.BodyParser(&src); err != nil {
			return err
		}

		err := service.LoginChek(db, src.Mail, src.Password)
		if err != nil {
			return c.Status(403).SendString("Invalid mail or password")
		}

		access, err := src.CreateJWT(2, private)
		if err != nil {
			return err
		}

		refresh, err := src.CreateJWT(24*7, private)
		if err != nil {
			return err
		}

		err = service.Login(db, refresh, src.Mail)
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
			"message":      "Auth successful",
			"access_token": access,
		})
	}
}
