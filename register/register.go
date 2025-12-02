package register

import (
	"database/sql"
	"goblog/keys"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func Regist(db *sql.DB) fiber.Handler {
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
		if !strings.Contains(data.Mail, "@") {
			return c.Status(400).SendString("Mail address is not valid")
		}
		var exists bool
		query := "SELECT EXISTS(SELECT 1 FROM users WHERE Mail = $1);"
		err := db.QueryRow(query, data.Mail).Scan(&exists)
		if err != nil {
			return err
		}
		if exists {
			return c.Status(403).SendString("Mail already used")
		}

		bytes, err := bcrypt.GenerateFromPassword([]byte(data.Password+data.Mail), 14)
		if err != nil {
			return err
		}
		data.Password = string(bytes)

		access, err := data.CreateJWT(2)
		if err != nil {
			return err
		}
		refresh, err := data.CreateJWT(24 * 7)
		if err != nil {
			return err
		}

		insertQuery := `
    INSERT INTO users (Mail, Password, Role, RefreshToken, RefreshTime)
    VALUES ($1, $2, $3, $4, $5)`

		_, err = db.Exec(insertQuery,
			data.Mail,
			data.Password,
			data.Role,
			refresh,
			time.Now().Add(time.Hour*time.Duration(24*7)).Unix(),
		)
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
			"message":      "Registration successful",
			"access_token": access,
		})
	}
}
