package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"goblog/post"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Anus() (*multipart.Writer, *bytes.Buffer, error) {

	path := os.Getenv("TEST_IMAGE")

	data, err := os.Open(path)
	if err != nil && err != io.EOF {
		return nil, nil, err
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("image", "test")
	if err != nil {
		return nil, nil, err
	}

	_, err = io.Copy(part, data)
	if err != nil && err != io.EOF {
		return nil, nil, err
	}

	_, err = data.Seek(0, 0)
	if err != nil && err != io.EOF {
		return nil, nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, nil, err
	}

	return writer, &body, nil

}

func TestAllPics(t *testing.T) {
	db, mn, test, _, _ := Load()
	defer db.Close()

	role := "Author"
	mail := "test@"

	test.Use("/valid", func(c *fiber.Ctx) error {
		c.Locals("role", role)
		c.Locals("mail", mail)
		return c.Next()
	})
	test.Use("/invalidRole", func(c *fiber.Ctx) error {
		c.Locals("role", "invalid")
		c.Locals("mail", mail)
		return c.Next()
	})
	test.Use("/invalidMail", func(c *fiber.Ctx) error {
		c.Locals("role", role)
		c.Locals("mail", "invalid")
		return c.Next()
	})

	test.Post("/valid/:postId/test", post.AddImage(db, mn))
	test.Post("/invalidRole/:postId/test", post.AddImage(db, mn))
	test.Post("/invalidMail/:postId/test", post.AddImage(db, mn))

	test.Delete("/valid/:postId/:imagePath/test", post.DeleteImage(db, mn))
	test.Delete("/invalidRole/:postId/:imagePath/test", post.DeleteImage(db, mn))
	test.Delete("/invalidMail/:postId/:imagePath/test", post.DeleteImage(db, mn))

	query := `INSERT INTO posts (Author, Key, Title, Content, Created, Updated, Status, Images)
				SELECT $1, $2, $3, $4, $5, $6, $7, ARRAY[$8]`
	_, err := db.Exec(query, mail, 1, "testpost", "post for test images", time.Now().Unix(), time.Now().Unix(), "Draft", nil)
	defer db.Exec("DELETE FROM posts WHERE Key = 1")
	if err != nil {
		t.Fatal(err)
	}

	writer, body, err := Anus()
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/invalidRole/1/test", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err := test.Test(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 403 {
		t.Errorf("Invalid status code: %d", res.StatusCode)
	}

	writer, body, err = Anus()
	if err != nil {
		t.Fatal(err)
	}

	req, err = http.NewRequest("POST", "/valid/invalid/test", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err = test.Test(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 409 {
		t.Errorf("Invalid status code: %d", res.StatusCode)
	}
	writer, body, err = Anus()
	if err != nil {
		t.Fatal(err)
	}

	req, err = http.NewRequest("POST", "/invalidMail/1/test", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err = test.Test(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 404 {
		t.Errorf("Invalid status code: %d", res.StatusCode)
	}

	writer, body, err = Anus()
	if err != nil {
		t.Fatal(err)
	}

	req, err = http.NewRequest("POST", "/valid/0/test", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err = test.Test(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 404 {
		t.Errorf("Invalid status code: %d", res.StatusCode)
	}

	writer, body, err = Anus()
	if err != nil {
		t.Fatal(err)
	}

	req, err = http.NewRequest("POST", "/valid/1/test", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err = test.Test(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 201 {
		t.Errorf("Invalid status code: %d", res.StatusCode)
	}

	var msg struct {
		Message string `json:"message"`
		Path    string `json:"path"`
	}
	err = json.NewDecoder(res.Body).Decode(&msg)
	if err != nil {
		t.Fatal(err)
	}
	pathname := strings.TrimPrefix(msg.Path, "images/")

	url := fmt.Sprintf("/invalidRole/%d/%s/test", 1, pathname)
	req, err = http.NewRequest("DELETE", url, nil)
	if err != nil {
		t.Fatal(err)
	}

	res, err = test.Test(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 403 {
		t.Errorf("Invalid status code: %d", res.StatusCode)
	}

	url = fmt.Sprintf("/imvalidMail/%d/%s/test", 1, pathname)
	req, err = http.NewRequest("DELETE", url, nil)
	if err != nil {
		t.Fatal(err)
	}

	res, err = test.Test(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 404 {
		t.Errorf("Invalid status code: %d", res.StatusCode)
	}

	url = fmt.Sprintf("/valid/invalid/%s/test", pathname)
	req, err = http.NewRequest("DELETE", url, nil)
	if err != nil {
		t.Fatal(err)
	}

	res, err = test.Test(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 409 {
		t.Errorf("Invalid status code: %d", res.StatusCode)
	}

	url = fmt.Sprintf("/valid/%d/%s/test", 0, pathname)
	req, err = http.NewRequest("DELETE", url, nil)
	if err != nil {
		t.Fatal(err)
	}

	res, err = test.Test(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 404 {
		t.Errorf("Invalid status code: %d", res.StatusCode)
	}

	url = fmt.Sprintf("/valid/%d/%s/test", 1, pathname)
	req, err = http.NewRequest("DELETE", url, nil)
	if err != nil {
		t.Fatal(err)
	}

	res, err = test.Test(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 200 {
		t.Errorf("Invalid status code: %d", res.StatusCode)
	}

}
