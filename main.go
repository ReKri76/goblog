package main

import (
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
	_ "github.com/lib/pq"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	godotenv.Load()
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

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, name)
	db, err := sql.Open("postgres", dsn)
	if err := db.Ping(); err != nil {
		log.Fatal("DB connection error:", err)
	}
	if err != nil {
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
				log.Fatal(err)
			}
			time.Sleep(time.Hour * 24)
		}
	}()

	app.Use("/api/chek", keys.ChekJWT(public))

	app.Post("/api/register", register.Regist(db, private))

	app.Post("/api/auth/login", register.Login(db, private))

	app.Post("/api/auth/refresh-token", register.Refresh(db, private, public))

	app.Post("/api/chek/post", post.CreatePost(db))

	app.Put("/api/chek/post/:postID", post.ChangePost(db))

	app.Patch("/api/chek/post/:postId/:status", post.PublicPost(db))

	app.Get("/api/chek/posts", post.ReadPost(db))

	app.Post("/api/chek/post/:postId/image", post.AddImage(db, mn))

	app.Listen(":8080")

}
