package service

import (
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func Regist(Mail, Password, Role, RefreshToken string, db *sql.DB) error {
	query := `
	INSERT INTO users (Mail, Password, Role, RefreshToken, RefreshTime)
	SELECT $1, $2, $3, $4, $5
	WHERE NOT EXISTS (
    	SELECT 1 FROM users WHERE Mail = $6
	)
`
	bytes, err := bcrypt.GenerateFromPassword([]byte(Password+Mail), 10)
	if err != nil {
		return err
	}
	Password = string(bytes)

	щекотливое, err := db.Exec(query,
		Mail,
		Password,
		Role,
		RefreshToken,
		time.Now().Add(time.Hour*time.Duration(24*7)).Unix(),
		Mail,
	)
	if err != nil {
		return err
	}
	rows, err := щекотливое.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("Mail already used")
	}

	return nil

}
