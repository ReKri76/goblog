package post

import (
	"database/sql"
	"time"

	"github.com/gofiber/fiber/v2"
)

func CreatePost(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role").(string)
		if role != "Author" {
			return c.Status(403).SendString("User is not author")
		}
		mail := c.Locals("mail").(string)
		type Post struct {
			Key     string `json:"idempotencyKey"`
			Title   string `json:"title"`
			Content string `json:"body"`
		}
		var src Post
		var exists bool
		if err := c.BodyParser(&src); err != nil {
			return err
		}
		query := "SELECT EXISTS(SELECT 1 FROM posts WHERE Key=$1)"
		err := db.QueryRow(query, src.Key).Scan(&exists)
		if err != nil {
			return err
		}
		if exists {
			return c.Status(409).SendString("Key already used")
		}
		query = "INSERT INTO posts (Author, Key, Title, Content, Created, Updated, Status) VALUES ($1, $2, $3, $4, $5, $6, $7)"
		_, err = db.Exec(query, mail, src.Key, src.Title, src.Content, time.Now(), time.Now(), "Draft")
		if err != nil {
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
			return c.Status(400).SendString("Invalid request")
		}
		Key, err := c.ParamsInt("postId")
		if err != nil {
			return c.Status(400).SendString("Invalid request")
		}
		query := "UPDATE posts SET Status = $3 WHERE Key = $1 AND Author = $2"
		res, err := db.Exec(query, Key, c.Locals("Mail").(string), "Published")
		if err != nil {
			return err
		}
		if rows, _ := res.RowsAffected(); rows == 0 {
			return c.Status(404).SendString("Post not found")
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
			return c.Status(400).SendString("Invalid request")
		}
		query := "UPDATE posts SET Title=$3, Content=$4, Updated=$5 WHERE Key = $1 AND Author = $2"
		res, err := db.Exec(query, Key, c.Locals("Mail").(string), src.Title, src.Content, time.Now().Unix())
		if err != nil {
			return err
		}
		if rows, _ := res.RowsAffected(); rows == 0 {
			return c.Status(404).SendString("Post not found")
		}
		return c.Status(200).SendString("Successfully changed post")
	}
}

func ReadPost(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		type Post struct {
			Id      int
			Author  string
			Title   string
			Content string
			Key     string
			Created string
			Updated string
			Status  string
		}
		//всегда будут возвращаться одни и те посты
		//нужно исправить
		var data = make([]Post, 0, 16)
		rows, err := db.Query("SELECT * FROM posts ORDER BY Created DESC LIMIT 16")
		if err != nil {
			return err
		}
		for rows.Next() {
			var post Post
			if err = rows.Scan(&post.Id, &post.Author, &post.Title, &post.Content, &post.Key, &post.Created, &post.Updated, &post.Status); err != nil {
				return err
			}
			if post.Status != "Draft" || post.Author == c.Locals("mail").(string) {
				data = append(data, post)
			}
		}
		return c.Status(200).JSON(&data)
	}
}
