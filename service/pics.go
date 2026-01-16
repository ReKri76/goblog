package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

func AddImageService(header *multipart.FileHeader, mn *minio.Client, db *sql.DB, mail string, key int) (error, string) {

	file, err := header.Open()
	defer file.Close()
	if err != nil {
		return err, ""
	}

	buf := make([]byte, 512)
	_, err = file.Read(buf)
	if err != nil && err != io.EOF {
		return err, ""
	}

	ext := http.DetectContentType(buf)
	if ext != "image/jpeg" && ext != "image/png" && ext != "image/gif" && ext != "image/webp" && ext != "image/tiff" && ext != "image/svg+xml" && ext != "image/pjpeg" {
		return errors.New("Image is not picture"), "" //400
	}
	ext = strings.TrimPrefix(ext, "image/")
	_, err = file.Seek(0, 0)
	if err != nil {
		return err, ""
	}

	name := header.Filename
	name = "[" + name + "][" + fmt.Sprint(time.Now().Unix()) + "]." + ext
	path := "images/" + name

	_, err = mn.PutObject(
		context.Background(),
		"images",
		name,
		file,
		header.Size,
		minio.PutObjectOptions{
			ContentType: "image/" + ext,
		},
	)
	if err != nil {
		return err, ""
	}

	//слово не воробей
	query := "UPDATE posts SET images = array_append(images, $1) where author=$2 AND key=$3 AND status='Draft'"
	res, err := db.Exec(query, path, mail, key)
	if err != nil {
		return err, ""
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err, ""
	}

	if rows == 0 {
		return errors.New("Not found"), "" //404
	}

	return nil, path
}
func DeleteImageService(mn *minio.Client, db *sql.DB, mail string, key int, imagePath string) error {
	query := "UPDATE posts SET images=array_remove(images,$3) where author=$1 AND key=$2 AND status='Draft'"
	res, err := db.Exec(query, mail, key, "images"+imagePath)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("Not found") //404
	}

	err = mn.RemoveObject(context.Background(), "images", imagePath, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}

	return nil
}
