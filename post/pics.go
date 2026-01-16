package post

import (
	"database/sql"
	"goblog/service"

	"github.com/gofiber/fiber/v2"
	"github.com/minio/minio-go/v7"
)

func AddImage(db *sql.DB, mn *minio.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if role := c.Locals("role").(string); role != "Author" {
			return c.Status(403).SendString("User is not author")
		}

		key, err := c.ParamsInt("postId")
		if err != nil {
			return c.Status(409).SendString("Invalid request")
		}

		header, err := c.FormFile("image")
		if err != nil {
			return err
		}

		err, path := service.AddImageService(header, mn, db, c.Locals("mail").(string), key)
		if err != nil {
			if err.Error() == "Image is not picture" {
				return c.Status(400).SendString(err.Error())
			}
			if err.Error() == "Not found" {
				return c.Status(404).SendString(err.Error())
			}
		}

		return c.Status(201).JSON(fiber.Map{
			"message": "ok",
			"path":    path,
		})
	}
}

func DeleteImage(db *sql.DB, mn *minio.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		mail := c.Locals("mail").(string)
		if role := c.Locals("role").(string); role != "Author" {
			return c.Status(403).SendString("User is not author")
		}

		key, err := c.ParamsInt("postId")
		if err != nil {
			return c.Status(409).SendString("Invalid request")
		}

		err = service.DeleteImageService(mn, db, mail, key, c.Query("path"))
		if err != nil {
			if err.Error() == "Not found" {
				return c.Status(404).SendString(err.Error())
			}
			return err
		}

		return c.Status(200).SendString("ok")
	}
}
