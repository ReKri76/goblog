package register

import (
	"crypto/rsa"
	"database/sql"
	"goblog/keys"
	"goblog/service"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Regist(db *sql.DB, private *rsa.PrivateKey) fiber.Handler {
	return func(c *fiber.Ctx) error {
		type Register struct {
			keys.Record
			Password string `json:"password"`
		}

		var data Register
		if err := c.BodyParser(&data); err != nil {
			return err
		}

		if data.Role != "Author" && data.Role != "Reader" {
			return c.Status(400).SendString("invalid role")
		}

		//Зачем вообще регистрировать по мейлу? Может еще и по пасспорту сразу? Моей проверки достаточно. Не условие задания, то я бы вообще не делал проверки.
		if !strings.Contains(data.Mail, "@") {
			return c.Status(400).SendString("Mail address is not valid")
		}

		access, err := data.CreateJWT(2, private)
		if err != nil {
			return err
		}

		refresh, err := data.CreateJWT(24*7, private)
		if err != nil {
			return err
		}

		err = service.Regist(data.Mail, data.Password, data.Role, refresh, db)
		if err != nil {
			if err.Error() == "Mail already used" {
				return c.Status(400).SendString("Mail already used")
			}
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
			"message":      "Registration successful",
			"access_token": access,
		})
	}
}
