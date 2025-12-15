package tests

import (
	"fmt"
	"goblog/register"
	"net/http"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestRegist(t *testing.T) {
	db, _, test, _, private := Load()
	defer db.Close()

	test.Post("/test", register.Regist(db, private))

	mail := "test@"
	role := "Author"
	password := "test"

	Register := fmt.Sprintf(`{"mail":"%s","role":"%s","password":"%s"}`, mail, role, password)
	reqValid, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader(Register))
	if err != nil {
		t.Fatal(err)
	}
	reqValid.Header.Set("Content-Type", "application/json; charset=utf-8")

	res, err := test.Test(reqValid)
	defer db.Exec("DELETE FROM users WHERE mail=$1", mail)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 200 {
		t.Errorf("status code != 200: %d", res.StatusCode)
	}

	var pass string
	err = db.QueryRow("SELECT Password FROM users WHERE Mail=$1", mail).Scan(&pass)
	if err != nil {
		t.Fatal(err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(pass), []byte(password+mail))
	if err != nil {
		t.Fatal(err)
	}

	Register = fmt.Sprintf(`{"mail":"%s","role":"%s","password":"%s"}`, mail, "invalid", password)
	invalidRole, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader(Register))
	if err != nil {
		t.Fatal(err)
	}
	invalidRole.Header.Set("Content-Type", "application/json; charset=utf-8")

	res, err = test.Test(invalidRole)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 400 {
		t.Errorf("status code != 400: %d", res.StatusCode)
	}

	Register = fmt.Sprintf(`{"mail":"%s","role":"%s","password":"%s"}`, "invalid_mail", role, password)
	invalidMail, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader(Register))
	if err != nil {
		t.Fatal(err)
	}
	invalidMail.Header.Set("Content-Type", "application/json; charset=utf-8")

	res, err = test.Test(invalidMail)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 400 {
		t.Errorf("status code != 400: %d", res.StatusCode)
	}

	Register = fmt.Sprintf(`{"mail":"%s","role":"%s","password":"%s"}`, mail, role, password)
	invalidDouble, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader(Register))
	if err != nil {
		t.Fatal(err)
	}
	invalidDouble.Header.Set("Content-Type", "application/json; charset=utf-8")

	res, err = test.Test(invalidDouble)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 403 {
		t.Errorf("status code != 400: %d", res.StatusCode)
	}
}
