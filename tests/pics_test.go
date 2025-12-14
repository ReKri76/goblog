package tests

import (
	"goblog/post"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestAddImage(t *testing.T) {

	db, mn, test, _, _ := Load()
	defer db.Close()

	test.Use("/valid", func(c *fiber.Ctx) error {
		c.Locals("role", "Author")
		return c.Next()
	})

	test.Post("/valid/test", post.AddImage(db, mn))

	Validreq, err := http.NewRequest("GET", "/valid/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	query := "INSERT INTO posts (key) values (1)"
	_, err = db.Exec(query)
	if err != nil {
		t.Fatal(err)
	}

	res, err := test.Test(Validreq)
	if err != nil {
		t.Fatal(err)
	}

}
