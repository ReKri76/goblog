package tests

import (
	"goblog/post"
	"log"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestAddAllPics(t *testing.T) {
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
		c.Locals("role", "invalid")
		c.Locals("mail", "invalid")
		return c.Next()
	})

	test.Post("/valid/:postId/test", post.AddImage(db, mn))
	test.Post("/invalidRole/:postId/test", post.AddImage(db, mn))
	test.Post("/invalidMail/:postId/test", post.AddImage(db, mn))

	test.Delete("/valid/:postId/:imagePath", post.DeleteImage(db, mn))
	test.Delete("/invalidRole/:postId/:imagePath", post.DeleteImage(db, mn))
	test.Delete("/invalidMail/:postId/:imagePath", post.DeleteImage(db, mn))

	path := os.Getenv("TEST_IMAGE")

	data, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer data.Close()

}
