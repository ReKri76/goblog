package tests

import (
	"fmt"
	"goblog/post"
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestCreatePost(t *testing.T) {
	db, _, test, _, _ := Load()
	defer db.Close()

	mail := "test@"
	role := "Author"

	test.Use("/valid", func(c *fiber.Ctx) error {
		c.Locals("role", role)
		c.Locals("mail", mail)
		return c.Next()
	})

	test.Use("/InvalidRole", func(c *fiber.Ctx) error {
		c.Locals("role", "invalid")
		c.Locals("mail", mail)
		return c.Next()
	})

	test.Post("/InvalidRole/test", post.CreatePost(db))
	test.Post("/valid/test", post.CreatePost(db))

	Post := fmt.Sprintf(`{"idempotencyKey":1, "title":" ","body": " " }`)
	req, err := http.NewRequest("POST", "/valid/test", strings.NewReader(Post))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := test.Test(req)
	defer db.Exec("delete from posts where key=1")
	defer res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 201 {
		t.Errorf("Status is not 201: %d", res.StatusCode)
	}

	req, err = http.NewRequest("POST", "/InvalidRole/test", strings.NewReader(Post))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err = test.Test(req)
	defer db.Exec("delete from posts where key=1")
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 403 {
		t.Errorf("Status is not 403: %d", res.StatusCode)
	}

	req, err = http.NewRequest("POST", "/valid/test", strings.NewReader(Post))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err = test.Test(req)
	defer db.Exec("delete from posts where key=1")
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 409 {
		t.Errorf("Status is not 409: %d", res.StatusCode)
	}

}
