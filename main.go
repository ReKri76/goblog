package main

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func main() {
	app := fiber.New()

	authHandler := &AuthHandler{&AuthStorage{map[string]User{}}}

	app.Post("/register", authHandler.Register)

	logrus.Fatal(app.Listen(":80"))
}

type (
	// Обработчик HTTP-запросов на регистрацию и аутентификацию пользователей
	AuthHandler struct {
		storage *AuthStorage
	}

	// Хранилище зарегистрированных пользователей
	// Данные хранятся в оперативной памяти
	AuthStorage struct {
		users map[string]User
	}

	// Структура данных с информацией о пользователе
	User struct {
		Email    string
		Name     string
		password string
	}
)

// Структура HTTP-запроса на регистрацию пользователя
type RegisterRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// Обработчик HTTP-запросов на регистрацию пользователя
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	regReq := RegisterRequest{}
	if err := c.BodyParser(&regReq); err != nil {
		return fmt.Errorf("body parser: %w", err)
	}

	// Проверяем, что пользователь с таким email еще не зарегистрирован
	if _, exists := h.storage.users[regReq.Email]; exists {
		return errors.New("the user already exists")
	}

	// Сохраняем в память нового зарегистрированного пользователя
	h.storage.users[regReq.Email] = User{
		Email:    regReq.Email,
		Name:     regReq.Name,
		password: regReq.Password,
	}

	return c.SendStatus(fiber.StatusCreated)
}
