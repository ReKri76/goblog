package main

import (
	"context"
	"database/sql"
	"fmt"
	"goblog/keys"
	"goblog/post"
	"goblog/register"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	app := fiber.New(fiber.Config{
		Prefork:       true,
		CaseSensitive: false,
		StrictRouting: false,
	})

	private, err := keys.LoadPrivateKey(os.Getenv("PRIVATE_KEY"))
	if err != nil {
		log.Fatal(err)
	}
	public, err := keys.LoadPublicKey(os.Getenv("PUBLIC_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))

	db, err := sql.Open("postgres", dsn)
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mn, err := minio.New(os.Getenv("MINIO_HOST"), &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("MINIO_USER"), os.Getenv("MINIO_PASSWORD"), ""),
		Secure: false,
	})
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			_, err = db.Exec("DELETE FROM users WHERE RefreshTime < $1", time.Now().Unix())
			if err != nil {
				log.Printf("[CLEANUP] Error deleting users: %v", err)
			}

			time.Sleep(time.Hour * 24)
		}
	}()

	go func() {
		var images pq.StringArray

		rows, err := db.Query("DELETE FROM posts WHERE ($1- Updated)<$2 RETURNING images", time.Now().Unix(), time.Hour*24*365)

		for rows.Next() {
			if err = rows.Scan(&images); err != nil {
				log.Printf("[CLEANUP] Error deleting posts: %v", err)
			}
			for _, image := range images {
				err := mn.RemoveObject(context.Background(), "images", image, minio.RemoveObjectOptions{})
				if err != nil {
					log.Printf("[CLEANUP] Error deleting posts: %v", err)
				}
			}
		}

		if err != nil {
			log.Printf("[CLEANUP] Error deleting users: %v", err)
		}

		time.Sleep(time.Hour * 24 * 7)

	}()

	app.Use("/api/chek", keys.ChekJWT(public))

	app.Post("/api/register", register.Regist(db, private))

	app.Post("/api/auth/login", register.Login(db, private))

	app.Post("/api/auth/refresh-token", register.Refresh(db, private, public))

	app.Post("/api/chek/post", post.CreatePost(db))

	app.Put("/api/chek/post/:postID", post.ChangePost(db))

	app.Patch("/api/chek/post/:postId/:status", post.PublicPost(db))

	app.Get("/api/chek/posts/:limit", post.ReadPost(db))

	app.Post("/api/chek/post/:postId/image", post.AddImage(db, mn))

	app.Delete("/api/chek/post/:postId/image/:imagePath", post.DeleteImage(db, mn))

	app.Listen(os.Getenv("FIBER_PORT"))

}
