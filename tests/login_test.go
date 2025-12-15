package tests

import (
	"fmt"
	"goblog/register"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestLogin(t *testing.T) {
	db, _, test, _, private := Load()
	defer db.Close()

	test.Post("/register", register.Regist(db, private))
	test.Post("/test", register.Login(db, private))

	mail := "test@"
	role := "Author"
	password := "test"

	Register := fmt.Sprintf(`{"mail":"%s","role":"%s","password":"%s"}`, mail, role, password)
	req, err := http.NewRequest(http.MethodPost, "/register", strings.NewReader(Register))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		t.Fatal(err)
	}

	_, err = test.Test(req)
	defer db.Exec("DELETE FROM users WHERE Mail = $1", mail)
	if err != nil {
		t.Fatal(err)
	}

	Loginer := fmt.Sprintf(`{"mail":"%s","role":"%s","password":"%s"}`, mail, role, password)
	reqValid, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader(Loginer))
	reqValid.Header.Set("Content-Type", "application/json")
	if err != nil {
		t.Fatal(err)
	}

	res, err := test.Test(reqValid)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 200 {
		t.Errorf("Status != 200 : %d", res.StatusCode)
	}

	LoginerInvalidMail := fmt.Sprintf(`{"mail":"%s","role":"%s","password":"%s"}`, "invalid", role, password)
	reqInvalidMail, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader(LoginerInvalidMail))
	reqInvalidMail.Header.Set("Content-Type", "application/json")
	if err != nil {
		t.Fatal(err)
	}

	res, err = test.Test(reqInvalidMail)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 403 {
		t.Errorf("Status != 403 : %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(body) != "Invalid mail or password" {
		t.Errorf("Invalid mail or password: %s", string(body))

	}

	LoginerInvalidPassword := fmt.Sprintf(`{"mail":"%s","role":"%s","password":"%s"}`, mail, role, "invalid")
	reqInvalidPassword, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader(LoginerInvalidPassword))
	reqInvalidPassword.Header.Set("Content-Type", "application/json")
	if err != nil {
		t.Fatal(err)
	}

	res, err = test.Test(reqInvalidPassword)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 403 {
		t.Errorf("Status != 403 : %d", res.StatusCode)
	}

	body, err = io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(body) != "Invalid mail or password" {
		t.Errorf("Invalid mail or password: %s", string(body))

	}

	LoginerInvalidRole := fmt.Sprintf(`{"mail":"%s","role":"%s","password":"%s"}`, mail, "invalid", password)
	reqInvalidRole, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader(LoginerInvalidRole))
	reqInvalidRole.Header.Set("Content-Type", "application/json")
	if err != nil {
		t.Fatal(err)
	}

	res, err = test.Test(reqInvalidRole)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 403 {
		t.Errorf("Status != 403 : %d", res.StatusCode)
	}
}
