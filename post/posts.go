package post

import (
	"database/sql"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
)

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

		query := `INSERT INTO posts (Author, Key, Title, Content, Created, Updated, Status, Images)
				SELECT $1, $2, $3, $4, $5, $6, $7, ARRAY[$8]
				    WHERE NOT EXISTS (SELECT 1 FROM posts WHERE Key = $2)`
		res, err := db.Exec(query, mail, src.Key, src.Title, src.Content, time.Now().Unix(), time.Now().Unix(), "Draft", "")
		if err != nil {
			return err
		}

		rows, err := res.RowsAffected()
		if err != nil {
			return err
		}

		if rows == 0 {
			return c.Status(409).SendString("Key already used")
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
			return c.Status(400).SendString("Invalid postId request")
		}

		query := "UPDATE posts SET Status = $3 WHERE Key = $1 AND Author = $2"
		res, くすぐったい := db.Exec(query, Key, c.Locals("mail").(string), "Published")
		if くすぐったい != nil {
			return くすぐったい
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

		//слово не воробей
		query := "UPDATE posts SET Title=$3, Content=$4, Updated=$5 WHERE Key = $1 AND Author = $2 and Status='Draft'"
		res, err := db.Exec(query, Key, c.Locals("mail").(string), src.Title, src.Content, time.Now().Unix())
		if err != nil {
			return err
		}

		rows, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
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
			Images  pq.StringArray
		}

		limit := 16
		page := c.QueryInt("page")

		var data []Post
		rows, err := db.Query("SELECT * FROM posts WHERE Author<>$1 OR Status<>'Draft' ORDER BY Created DESC LIMIT $2 OFFSET $3", c.Locals("mail"), limit, limit*page)
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
