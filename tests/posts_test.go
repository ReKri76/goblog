package tests

import (
	"encoding/json"
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

	test.Use("/InvalidMail", func(c *fiber.Ctx) error {
		c.Locals("role", role)
		c.Locals("mail", "invalid")
		return c.Next()
	})

	test.Post("/InvalidRole/test/:postId/:status", post.PublicPost(db))
	test.Post("/valid/test/:postId/:status", post.PublicPost(db))
	test.Post("/InvalidMail/test/:postId/:status", post.PublicPost(db))

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

	reqInValidMail, err := http.NewRequest("POST", "/InvalidMail/test/1/Published", nil)
	if err != nil {
		t.Fatal(err)
	}

	res, err = test.Test(reqInValidMail)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 404 {
		t.Errorf("Status is not 404: %d", res.StatusCode)
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

	if res.StatusCode != 409 {
		t.Errorf("Status is not 409: %d", res.StatusCode)
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

// как же я заебался писать это

func TestChangePost(t *testing.T) {
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

	// Добавляем middleware для проверки email
	test.Use("/InvalidMail", func(c *fiber.Ctx) error {
		c.Locals("role", role)
		c.Locals("mail", "invalid")
		return c.Next()
	})

	test.Post("/InvalidRole/test/:postId/", post.ChangePost(db))
	test.Post("/valid/test/:postId/", post.ChangePost(db))
	test.Post("/InvalidMail/test/:postId/", post.ChangePost(db))

	query := `INSERT INTO posts (Author, Key, Title, Content, Created, Updated, Status, Images)
				values ($1, $2, $3, $4, $5, $6, $7, ARRAY[$8])`
	_, err := db.Exec(query, mail, 1, "TestTitle", "TestBody", time.Now().Unix(), time.Now().Unix(), "Draft", "")
	defer db.Exec("delete from posts where key=1")
	if err != nil {
		t.Fatal(err)
	}

	Post := fmt.Sprintf(`{"title":"new","body": "new" }`)
	reqValid, err := http.NewRequest("POST", "/valid/test/1/", strings.NewReader(Post))
	if err != nil {
		t.Fatal(err)
	}
	reqValid.Header.Set("Content-Type", "application/json")

	res, err := test.Test(reqValid)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 200 {
		t.Errorf("Status is not 200: %d", res.StatusCode)
	}

	reqInvalidMail, err := http.NewRequest("POST", "/InvalidMail/test/1/", strings.NewReader(Post))
	if err != nil {
		t.Fatal(err)
	}
	reqInvalidMail.Header.Set("Content-Type", "application/json")

	res, err = test.Test(reqInvalidMail)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 404 {
		t.Errorf("Status is not 404 when mail doesn't match: %d", res.StatusCode)
	}

	reqInValid, err := http.NewRequest("POST", "/InvalidRole/test/1/", strings.NewReader(Post))
	if err != nil {
		t.Fatal(err)
	}
	reqInValid.Header.Set("Content-Type", "application/json")

	res, err = test.Test(reqInValid)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 403 {
		t.Errorf("Status is not 403: %d", res.StatusCode)
	}

	reqInValidId, err := http.NewRequest("POST", "/valid/test/invalid", strings.NewReader(Post))
	if err != nil {
		t.Fatal(err)
	}
	reqInValidId.Header.Set("Content-Type", "application/json")

	res, err = test.Test(reqInValidId)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 409 {
		t.Errorf("Status is not 409: %d", res.StatusCode)
	}

	req404, err := http.NewRequest("POST", "/valid/test/0", strings.NewReader(Post))
	if err != nil {
		t.Fatal(err)
	}
	req404.Header.Set("Content-Type", "application/json")

	res, err = test.Test(req404)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 404 {
		t.Errorf("Status is not 404: %d", res.StatusCode)
	}

	reqDraft, err := http.NewRequest("POST", "/valid/test/1", strings.NewReader(Post))
	if err != nil {
		t.Fatal(err)
	}
	reqDraft.Header.Set("Content-Type", "application/json")
	_, err = db.Exec("update posts set Status='invalid'  where key=1")
	if err != nil {
		t.Fatal(err)
	}

	res, err = test.Test(reqDraft)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 404 {
		t.Errorf("Status is not 404: %d", res.StatusCode)
	}
}

func TestReadPost(t *testing.T) {
	db, _, test, _, _ := Load()
	defer db.Close()

	mail1 := "test@1"
	mail2 := "test@2"

	for i := 1; i < 128; i++ {

		var status string
		if i%3 == 0 {
			status = "Draft"
		} else {
			status = "Published"
		}
		if i%2 == 0 {
			query := `INSERT INTO posts (Author, Key, Title, Content, Created, Updated, Status, Images)
				values ($1, $2, $3, $4, $5, $6, $7, ARRAY[$8])`
			_, err := db.Exec(query, mail1, i, "TestTitle", "TestBody", time.Now().Unix(), time.Now().Unix(), status, "")
			defer db.Exec(fmt.Sprintf("delete from posts where key=%d", i))
			if err != nil {
				t.Fatal(err)
			}
		} else {
			query := `INSERT INTO posts (Author, Key, Title, Content, Created, Updated, Status, Images)
				values ($1, $2, $3, $4, $5, $6, $7, ARRAY[$8])`
			_, err := db.Exec(query, mail2, i, "TestTitle", "TestBody", time.Now().Unix(), time.Now().Unix(), "Draft", "")
			defer db.Exec(fmt.Sprintf("delete from posts where key=%d", i))
			if err != nil {
				t.Fatal(err)
			}

		}
	}
	test.Use(func(c *fiber.Ctx) error {
		c.Locals("mail", mail1)
		return c.Next()
	})
	test.Get("/test", post.ReadPost(db))

	for i := range make([]bool, 64) {
		url := fmt.Sprintf("/test/?page=%d", i)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			t.Fatal(err)
		}

		res, err := test.Test(req)
		if err != nil {
			t.Fatal(err)
		}

		var body struct {
			Data []post.Post `json:"data"`
			Page int         `json:"page"`
		}
		err = json.NewDecoder(res.Body).Decode(&body)
		if err != nil {
			t.Fatal(err)
		}

		if len(body.Data) > 16 {
			t.Errorf("Post #%d is too large", i)
		}

		for _, v := range body.Data {
			if v.Status == "Draft" && v.Author != mail1 {
				t.Errorf("Post #%d is draft", i)
			}
		}
	}

}
