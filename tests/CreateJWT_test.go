package tests

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"goblog/keys"
	"io"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func JWT_test(pub *rsa.PublicKey, private *rsa.PrivateKey) (bool, error) {
	mail := "test"
	role := "test"
	record := keys.Record{mail, role}

	token, err := record.CreateJWT(1337, private)
	if err != nil {
		return false, err
	}

	test := fiber.New()

	test.Use(keys.ChekJWT(pub))

	test.Get("/test", func(c *fiber.Ctx) error {
		Mail, ok := c.Locals("mail").(string)
		Role, ok2 := c.Locals("role").(string)
		if !ok || !ok2 {
			return c.SendStatus(500)
		}
		return c.Status(200).SendString(Mail + " " + Role)
	})

	reqValid, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		return false, err
	}

	reqValid.Header.Set("Authorization", "Bearer "+token)

	res, err := test.Test(reqValid)
	if err != nil {
		return false, err
	}

	if res.StatusCode != 200 {
		return false, errors.New("[ERROR] Status incorrect. Code:1. Status: " + res.Status)
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return false, err
	}

	parts := strings.Split(string(body), " ")
	if len(parts) != 2 {
		return false, fmt.Errorf("[ERROR] Incorrect number of parts. Expected: 2, got: %d", len(parts))
	}

	if parts[0] != mail {
		return false, errors.New("[ERROR] Incorrect mail address")
	}

	if parts[1] != role {
		return false, errors.New("[ERROR] Incorrect role")
	}

	reqInValid1, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		return false, err
	}

	reqInValid1.Header.Set("InvalidHeader", "Bearer "+token)

	res, err = test.Test(reqInValid1)
	if err != nil {
		return false, err
	}

	if res.StatusCode != 401 {
		return false, errors.New("[ERROR] Status incorrect. Code:2. Status: " + res.Status)
	}

	reqInValid2, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		return false, err
	}

	reqInValid2.Header.Set("Authorization", token)

	res, err = test.Test(reqInValid2)
	if err != nil {
		return false, err
	}

	if res.StatusCode != 401 {
		return false, errors.New("[ERROR] Status incorrect. Code:3. Status: " + res.Status)
	}

	reqInValid3, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		return false, err
	}

	reqInValid3.Header.Set("InvalidHeader", "Bearer "+token+"invalid_part")

	res, err = test.Test(reqInValid3)
	if err != nil {
		return false, err
	}

	if res.StatusCode != 401 {
		return false, errors.New("[ERROR] Status incorrect. Code:4. Status: " + res.Status)
	}

	return true, nil
}
