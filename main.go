package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
)

type record interface {
	CreateJWT() (string, error)
}

type DBrecord struct {
	Mail         string `json:"mail"`
	Password     string `json:"password"`
	Role         string `json:"role"`
	RefreshToken string `json:"refresh_token"`
	RefreshTime  int64  `json:"refresh_time"`
}

func loadPrivateKey(filename string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func (src DBrecord) CreateJWT(ttl int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": src.Mail,
		"role":    src.Role,
		"exp":     time.Now().Add(time.Hour * time.Duration(ttl)).Unix(),
	}
	key, err := loadPrivateKey("./keys/private.pem")
	if err != nil {
		return "", err
	}
	return jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)
}

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

	app.Post("/api/register", func(c *fiber.Ctx) error {
		var data DBrecord
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

		insertQuery := `
    INSERT INTO users (Mail, Password, Role, RefreshToken, RefreshTime)
    VALUES ($1, $2, $3, $4, $5)`

		_, err = db.Exec(insertQuery,
			data.Mail,
			data.Password,
			data.Role,
			data.RefreshToken,
			data.RefreshTime,
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
	})

	app.Post("/api/auth/login", func(c *fiber.Ctx) error {
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

	app.Listen(":8080")

}
