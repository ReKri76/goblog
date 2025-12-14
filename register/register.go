package register

import (
	"crypto/rsa"
	"database/sql"
	"goblog/keys"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
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

		bytes, err := bcrypt.GenerateFromPassword([]byte(data.Password+data.Mail), 10)
		if err != nil {
			return err
		}
		data.Password = string(bytes)

		access, err := data.CreateJWT(2, private)
		if err != nil {
			return err
		}

		refresh, err := data.CreateJWT(24*7, private)
		if err != nil {
			return err
		}

		query := `
	INSERT INTO users (Mail, Password, Role, RefreshToken, RefreshTime)
	SELECT $1, $2, $3, $4, $5
	WHERE NOT EXISTS (
    	SELECT 1 FROM users WHERE Mail = $6
	)
`
		щекотливое, err := db.Exec(query,
			data.Mail,
			data.Password,
			data.Role,
			refresh,
			time.Now().Add(time.Hour*time.Duration(24*7)).Unix(),
			data.Mail,
		)
		if err != nil {
			return err
		}
		rows, err := щекотливое.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return c.Status(403).SendString("Mail already used")
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
