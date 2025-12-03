package register

import (
	"crypto/rsa"
	"database/sql"
	"goblog/keys"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Refresh(db *sql.DB, private *rsa.PrivateKey, public *rsa.PublicKey) fiber.Handler {
	return func(c *fiber.Ctx) error {
		type refresher struct {
			keys.Record
			RefreshToken string `json:"refresh_token"`
			RefreshTime  int64  `json:"refresh_time"`
		}
		var src refresher
		if err := c.BodyParser(&src); err != nil {
			return err
		}
		refresh := c.Cookies("refresh_token")
		query := "SELECT RefreshToken, RefreshTime FROM users WHERE Mail = $1"
		var data refresher
		err := db.QueryRow(query, src.Mail).Scan(&data.RefreshToken, &data.RefreshTime)
		if err != nil {
			return err
		}
		//проверить подпись токена
		//проверить время токена
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
