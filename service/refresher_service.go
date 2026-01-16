package service

import (
	"database/sql"
	"time"
)

func RefresherService(db *sql.DB, Mail string) error {
	var Refreshtime int64
	query := "SELECT RefreshTime FROM users WHERE Mail = $1"
	err := db.QueryRow(query, Mail).Scan(&Refreshtime)
	if err != nil {
		return err
	}
	if Refreshtime < time.Now().Unix() {
		return err
	}

	return nil
}
