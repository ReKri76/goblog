package service

import (
	"database/sql"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func LoginChek(db *sql.DB, Mail string, Password string) error {

	type loginer struct {
		Role     string
		Password string
	}

	var Lole loginer

	query := "SELECT Role, Password FROM users WHERE Mail = $1"
	err := db.QueryRow(query, Mail).Scan(&Lole.Role, &Lole.Password)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(Lole.Password), []byte(Password+Mail))
	if err != nil {
		return err
	}

	return nil
}

func Login(db *sql.DB, refresh string, Mail string) error {

	query := "UPDATE users SET RefreshToken = $1, RefreshTime = $2 WHERE Mail = $3"
	_, err := db.Exec(query,
		refresh,
		time.Now().Add(time.Hour*time.Duration(24*7)).Unix(),
		Mail,
	)
	if err != nil {
		return err
	}

	return nil
}
