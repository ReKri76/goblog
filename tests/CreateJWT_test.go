package tests

import (
	"goblog/keys"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func TestJWT(t *testing.T) {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	private, err := keys.LoadPrivateKey(os.Getenv("PRIVATE_KEY"))
	if err != nil {
		log.Fatal(err)
	}
	pub, err := keys.LoadPublicKey(os.Getenv("PUBLIC_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	mail := "test"
	role := "test"
	record := keys.Record{mail, role}
	token, err := record.CreateJWT(1337, private)
	if err != nil {
		t.Fatalf("%v", err)
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
		t.Fatalf("%v", err)
	}

	reqValid.Header.Set("Authorization", "Bearer "+token)

	res, err := test.Test(reqValid)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if res.StatusCode != 200 {
		t.Errorf("Status incorrect. Code:1. Status: %v", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}

	parts := strings.Split(string(body), " ")
	if len(parts) != 2 {
		t.Errorf("Incorrect number of parts. Expected: 2, got: %d", len(parts))
	}

	if parts[0] != mail {
		t.Errorf("Incorrect mail address")
	}

	if parts[1] != role {
		t.Errorf("Incorrect role")
	}

	reqInValid1, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	reqInValid1.Header.Set("InvalidHeader", "Bearer "+token)

	res, err = test.Test(reqInValid1)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if res.StatusCode != 401 {
		t.Errorf("Status incorrect. Code:2. Status: %v", res.Status)
	}

	reqInValid2, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	reqInValid2.Header.Set("Authorization", token)

	res, err = test.Test(reqInValid2)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if res.StatusCode != 401 {
		t.Errorf("Status incorrect. Code:3. Status: %v", res.Status)
	}

	reqInValid3, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	reqInValid3.Header.Set("InvalidHeader", "Bearer "+token+"invalid_part")

	res, err = test.Test(reqInValid3)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if res.StatusCode != 401 {
		t.Errorf("Status incorrect. Code:4. Status: %v", res.Status)
	}

}
