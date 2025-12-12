package register

import (
	"crypto/rsa"
	"database/sql"
	"goblog/keys"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
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

		var data loginer
		query := "SELECT Role, Password FROM users WHERE Mail = $1"
		err := db.QueryRow(query, src.Mail).Scan(&data.Role, &data.Password)
		if err != nil {
			time.Sleep(time.Second * 5)
			return c.Status(403).SendString("Invalid mail or password")
		}

		err = bcrypt.CompareHashAndPassword([]byte(data.Password), []byte(src.Password+src.Mail))
		if err != nil {
			time.Sleep(time.Second * 5)
			return c.Status(403).SendString("Invalid mail or password")
		}

		if src.Role != data.Role {
			return c.Status(500).SendString("Invalid role")
		}
		access, err := src.CreateJWT(2, private)
		if err != nil {
			return err
		}
		refresh, err := src.CreateJWT(24*7, private)
		if err != nil {
			return err
		}
		insertQuery := "UPDATE users SET RefreshToken = $1, RefreshTime = $2 WHERE Mail = $3"
		_, err = db.Exec(insertQuery,
			refresh,
			time.Now().Add(time.Hour*time.Duration(24*7)).Unix(),
			src.Mail,
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
			"message":      "Auth successful",
			"access_token": access,
		})
	}
}
