package tests

import (
	"crypto/rsa"
	"database/sql"
	"fmt"
	"goblog/keys"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func Load() (db *sql.DB, mn *minio.Client, app *fiber.App, public *rsa.PublicKey, private *rsa.PrivateKey) {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	app = fiber.New(fiber.Config{
		Prefork:       true,
		CaseSensitive: false,
		StrictRouting: false,
	})

	private, err := keys.LoadPrivateKey(os.Getenv("PRIVATE_KEY"))
	if err != nil {
		log.Fatal(err)
	}
	public, err = keys.LoadPublicKey(os.Getenv("PUBLIC_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))

	db, err = sql.Open("postgres", dsn)
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	mn, err = minio.New(os.Getenv("MINIO_HOST"), &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("MINIO_USER"), os.Getenv("MINIO_PASSWORD"), ""),
		Secure: false,
	})
	if err != nil {
		log.Fatal(err)
	}

	return db, mn, app, public, private

}
