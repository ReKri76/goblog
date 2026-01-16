package service

import (
	"database/sql"
	"errors"
	"time"
)

func CreatePostsService(db *sql.DB, mail string, key int, title string, content string) error {

	query := `INSERT INTO posts (Author, Key, Title, Content, Created, Updated, Status, Images)
				SELECT $1, $2, $3, $4, $5, $6, $7, ARRAY[$8]
				    WHERE NOT EXISTS (SELECT 1 FROM posts WHERE Key = $2)`
	res, err := db.Exec(query, mail, key, title, content, time.Now().Unix(), time.Now().Unix(), "Draft", "")
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("Key already used")
	}

	return nil
}

func PublicPostService(db *sql.DB, mail string, key int) error {

	query := "UPDATE posts SET Status = $3 WHERE Key = $1 AND Author = $2"
	res, くすぐったい := db.Exec(query, key, mail, "Published")
	if くすぐったい != nil {
		return くすぐったい
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return errors.New("Post not found")
	}

	return nil
}

func ChangePostService(db *sql.DB, mail string, key int, title string, content string) error {

	//слово не воробей
	query := "UPDATE posts SET Title=$3, Content=$4, Updated=$5 WHERE Key = $1 AND Author = $2 and Status='Draft'"
	res, err := db.Exec(query, key, mail, title, content, time.Now().Unix())
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("Post not found")
	}

	return nil
}
