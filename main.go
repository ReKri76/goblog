package main

import (
	"database/sql"
	"fmt"
	"goblog/keys"
	"goblog/register"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
)

func main() {
	app := fiber.New(fiber.Config{
		Prefork:       true,
		CaseSensitive: false,
		StrictRouting: false,
	})
	private, err := keys.LoadPrivateKey("keys/private.pem")
	if err != nil {
		log.Fatal(err)
	}
	public, err := keys.LoadPublicKey("keys/public.pem")
	if err != nil {
		log.Fatal(err)
	}
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=BhD13 dbname=postgres sslmode=disable")
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	defer db.Close()

	go func() {
		for {
			time.Sleep(time.Hour * 24)
			db.Exec("DELETE FROM users WHERE RefreshTime < NOW()")
		}
	}()

	app.Post("/api/register", register.Regist(db, private))

	app.Use("/api/auth", keys.ChekJWT(public))

	app.Post("/api/auth/login", register.Login(db, private))

	app.Post("/api/auth/refresh-token", register.Refresh(db, private, public))

	app.Listen(":8080")

}
