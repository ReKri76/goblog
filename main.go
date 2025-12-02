package main

import (
	"database/sql"
	"fmt"
	"goblog/register"

	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
)

func main() {
	app := fiber.New(fiber.Config{
		Prefork:       true,
		CaseSensitive: false,
		StrictRouting: false,
	})

	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=BhD13 dbname=postgres sslmode=disable")
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	defer db.Close()

	app.Post("/api/register", register.Regist(db))

	/*app.Post("/api/auth/login", func(c *fiber.Ctx) error {
		var src DBrecord
		if err := c.BodyParser(&src); err != nil {
			return err
		}
		var data DBrecord
		query := "SELECT Mail, Role, Password FROM users WHERE Mail = $1"
		err = db.QueryRow(query, src.Mail).Scan(&data.Mail, &data.Role, &data.Password)
		if err != nil {
			return c.Status(403).SendString("Invalid mail or password")
		}
		err = bcrypt.CompareHashAndPassword([]byte(data.Password), []byte(src.Password+src.Mail))
		if err != nil {
			return c.Status(403).SendString("Invalid mail or password")
		}
		access, err := data.CreateJWT(2)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Could not create access token",
			})
		}
		refresh, err := data.CreateJWT(24 * 7)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Could not create refresh token",
			})
		}
		data.RefreshToken = refresh
		data.RefreshTime = time.Now().Add(time.Hour * time.Duration(24*7)).Unix()
		insertQuery := "UPDATE users SET RefreshToken = $1, RefreshTime = $2 WHERE Mail = $3"
		_, err = db.Exec(insertQuery,
			data.RefreshToken,
			data.RefreshTime,
			data.Mail,
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
			"message":       "Auth successful",
			"access_token":  access,
			"refresh_token": refresh,
		})
	})

	app.Post("/api/auth/refresh-token", func(c *fiber.Ctx) error {
		var src DBrecord
		if err := c.BodyParser(&src); err != nil {
			return err
		}
		query := "SELECT RefreshToken, RefreshTime, Mail, Role FROM users WHERE Mail = $1"
		var data DBrecord
		err = db.QueryRow(query, src.Mail).Scan(&data.RefreshToken, &data.RefreshTime, &data.Mail, &data.Role)
		if err != nil {
			return err
		}
		if data.RefreshTime < src.RefreshTime {
			return c.Status(400).SendString("Invalid refresh time")
		}
		access, err := data.CreateJWT(2)
		if err != nil {
			return err
		}
		refresh, err := data.CreateJWT(24 * 7)
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
	})
	*/
	app.Listen(":8080")

}
