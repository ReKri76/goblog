package post

import (
	"database/sql"
	"goblog/service"

	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
)

type Post struct {
	Id      int
	Key     int
	Author  string
	Title   string
	Content string
	Created int64
	Updated int64
	Status  string
	Images  pq.StringArray
}

func CreatePost(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role").(string)
		if role != "Author" {
			return c.Status(403).SendString("User is not author")
		}

		mail := c.Locals("mail").(string)

		type Post struct {
			Key     int    `json:"idempotencyKey"`
			Title   string `json:"title"`
			Content string `json:"body"`
		}

		var src Post
		if err := c.BodyParser(&src); err != nil {
			return err
		}

		err := service.CreatePostsService(db, mail, src.Key, src.Title, src.Content)
		if err != nil {
			if err.Error() == "Key already used" {
				return c.Status(409).SendString("Key already used")
			}
			return err
		}

		return c.Status(201).SendString("Successfully created post")
	}
}

func PublicPost(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role").(string)
		if role != "Author" {
			return c.Status(403).SendString("User is not author")
		}

		if c.Params("status") != "Published" {
			return c.Status(400).SendString("Invalid status request")
		}

		Key, err := c.ParamsInt("postId")
		if err != nil {
			return c.Status(409).SendString("Invalid postId request")
		}

		err = service.PublicPostService(db, c.Locals("mail").(string), Key)
		if err != nil {
			if err.Error() == "Post not found" {
				return c.Status(404).SendString("Post not found")
			}
		}

		return c.Status(200).SendString("Post published")
	}
}

func ChangePost(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role").(string)
		if role != "Author" {
			return c.Status(403).SendString("User is not author")
		}

		type Post struct {
			Title   string `json:"title"`
			Content string `json:"body"`
		}

		var src Post
		if err := c.BodyParser(&src); err != nil {
			return err
		}

		Key, err := c.ParamsInt("postId")
		if err != nil {
			return c.Status(409).SendString("Invalid request")
		}

		err = service.ChangePostService(db, c.Locals("mail").(string), Key, src.Title, src.Content)
		if err != nil {
			if err.Error() == "Post not found" {
				return c.Status(404).SendString("Post not found")
			}
		}

		return c.Status(200).SendString("Successfully changed post")
	}
}

func ReadPost(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		//ну тут уже вроде как напрямую запрос в бд и ответ, так что сервис не нужен
		limit := 16
		page := c.QueryInt("page")

		var data []Post
		rows, err := db.Query("SELECT * FROM posts WHERE Author=$1 OR Status<>'Draft' ORDER BY Created DESC LIMIT $2 OFFSET $3", c.Locals("mail"), limit, limit*page)
		if err != nil {
			return err
		}

		for rows.Next() {
			var post Post
			if err = rows.Scan(&post.Id, &post.Key, &post.Title, &post.Content, &post.Created, &post.Updated, &post.Status, &post.Images, &post.Author); err != nil {
				return err
			}
			data = append(data, post)
		}
		return c.Status(200).JSON(fiber.Map{
			"data": &data,
			"page": page,
		})
	}
}
