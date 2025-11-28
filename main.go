package main

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

var key = []byte("tmp_key") //////////////////////

type DBrecord struct {
	Id       int
	Mail     string `json:"mail"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// ///////////////////////////////////////////////////////////////////
func (src DBrecord) CreateJWTAcess() (string, error) {
	claims := jwt.MapClaims{
		"user_id": src.Mail,
		"role":    src.Role,
		"exp":     time.Now().Add(time.Hour * 2).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(key)
}

func (src DBrecord) CreateJWTRefresh() (string, error) {
	claims := jwt.MapClaims{
		"user_id": src.Mail,
		"role":    src.Role,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(key)
}

// ////////////////////////////////////////////////////////////////////////////////
func main() {

	app := fiber.New(fiber.Config{
		Prefork:       true,
		CaseSensitive: false,
		StrictRouting: false,
	})

	app.Post("/api/register", func(c *fiber.Ctx) error {
		var data DBrecord
		if err := c.BodyParser(&data); err != nil {
			return err
		}
		if !strings.Contains(data.Mail, "@") {
			return c.Status(400).SendString("Mail address is not valid")
		}

		//И если data.Mail не содержится в базе данных
		//Добавить в базу данных запись(пока не знаю как рабоать с postgresql здесь, завтра разберусь)

		access, err := data.CreateJWTAcess()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Could not create access token",
			})
		}
		refresh, err := data.CreateJWTRefresh()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Could not create refresh token",
			})
		}
		c.Cookie(&fiber.Cookie{
			Name:     "refresh_token",
			Value:    refresh,
			Expires:  time.Now().Add(24 * 7 * time.Hour),
			HTTPOnly: true,
			Secure:   true,
			SameSite: "Lax",
		})

		// Возвращаем access token в теле ответа
		return c.Status(200).JSON(fiber.Map{
			"message":      "Registration successful",
			"access_token": access,
		})
	})
	app.Listen(":8080")

}
