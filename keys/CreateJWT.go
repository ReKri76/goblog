package keys

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type Record struct {
	Mail string `json:"mail"`
	Role string `json:"role"`
}

func LoadPrivateKey(filename string) (*rsa.PrivateKey, error) {
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

func LoadPublicKey(filename string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}
	pub, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return pub, nil
}

func (src Record) CreateJWT(ttl int, key *rsa.PrivateKey) (string, error) {
	claims := jwt.MapClaims{
		"mail": src.Mail,
		"role": src.Role,
		"exp":  time.Now().Add(time.Hour * (time.Duration(ttl))).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)
}

func ChekJWT(public *rsa.PublicKey) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			return c.Status(401).SendString("Authorization required")
		}
		parts := strings.Split(auth, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(401).SendString("Invalid Authorization header")
		}
		src := parts[1]

		token, err := jwt.Parse(src, func(token *jwt.Token) (interface{}, error) {
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

		mail, ok := claims["mail"].(string)
		if !ok || mail == "" {
			return c.Status(401).SendString("Invalid token")
		}
		role, ok := claims["role"].(string)
		if !ok || role == "" {
			return c.Status(401).SendString("Invalid token")
		}

		c.Locals("mail", mail)
		c.Locals("role", role)
		return c.Next()
	}
}
