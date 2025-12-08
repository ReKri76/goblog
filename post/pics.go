package post

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/minio/minio-go/v7"
)

func AddImage(db *sql.DB, mn *minio.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		mail := c.Locals("mail").(string)
		if role := c.Locals("role").(string); role != "Author" {
			return c.Status(403).SendString("User is not author")
		}

		header, err := c.FormFile("image")
		if err != nil {
			return err
		}

		file, err := header.Open()
		if err != nil {
			return err
		}

		buf := make([]byte, 512)
		_, err = file.Read(buf)
		if err != nil {
			return err
		}

		ext := http.DetectContentType(buf)
		if ext != "image/jpeg" && ext != "image/png" && ext != "image/gif" && ext != "image/webp" && ext != "image/tiff" && ext != "image/svg+xml" && ext != "image/pjpeg" {
			return c.Status(400).SendString("File is not a picture")
		}
		ext = strings.TrimPrefix(ext, "image/")
		_, err = file.Seek(0, 0)
		if err != nil {
			return err
		}

		name := header.Filename
		name = "<" + name + "><" + fmt.Sprint(time.Now().Unix()) + ">." + ext
		path := "images/" + name

		_, err = mn.PutObject(
			context.Background(),
			"images",
			name,
			file,
			header.Size,
			minio.PutObjectOptions{
				ContentType: "image/" + ext,
			},
		)
		if err != nil {
			return err
		}

		query := "UPDATE posts SET images = array_append(images, $1) where author=$2"
		_, err = db.Exec(query, path, mail)
		if err != nil {
			return err
		}

		return c.Status(201).JSON(fiber.Map{
			"message": "ok",
			"path":    path,
		})
	}
}
