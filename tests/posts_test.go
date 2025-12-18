package tests

import (
	"fmt"
	"goblog/post"
	"net/http"
	"strings"
	"testing"
	"time"

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

func TestPublicPost(t *testing.T) {
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

	test.Post("/InvalidRole/test/:postId/:status", post.PublicPost(db))
	test.Post("/valid/test/:postId/:status", post.PublicPost(db))

	query := `INSERT INTO posts (Author, Key, Title, Content, Created, Updated, Status, Images)
				values ($1, $2, $3, $4, $5, $6, $7, ARRAY[$8])`
	_, err := db.Exec(query, mail, 1, "TestTitle", "TestBody", time.Now().Unix(), time.Now().Unix(), "Draft", "")
	defer db.Exec("delete from posts where key=1")
	if err != nil {
		t.Fatal(err)
	}

	reqValid, err := http.NewRequest("POST", "/valid/test/1/Published", nil)
	if err != nil {
		t.Fatal(err)
	}

	res, err := test.Test(reqValid)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 200 {
		t.Errorf("Status is not 200: %d", res.StatusCode)
	}

	reqInValidRole, err := http.NewRequest("POST", "/InvalidRole/test/1/Published", nil)
	if err != nil {
		t.Fatal(err)
	}

	res, err = test.Test(reqInValidRole)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 403 {
		t.Errorf("Status is not 403: %d", res.StatusCode)
	}

	reqInvalidStatus, err := http.NewRequest("POST", "/valid/test/1/invalid", nil)
	if err != nil {
		t.Fatal(err)
	}

	res, err = test.Test(reqInvalidStatus)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 400 {
		t.Errorf("Status is not 400: %d", res.StatusCode)
	}

	reqInValidId, err := http.NewRequest("POST", "/valid/test/invalid/Published", nil)
	if err != nil {
		t.Fatal(err)
	}

	res, err = test.Test(reqInValidId)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 400 {
		t.Errorf("Status is not 400: %d", res.StatusCode)
	}

	req404, err := http.NewRequest("POST", "/valid/test/0/Published", nil)
	if err != nil {
		t.Fatal(err)
	}

	res, err = test.Test(req404)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 404 {
		t.Errorf("Status is not 404: %d", res.StatusCode)
	}
}
