package tests

import (
	"fmt"
	"goblog/keys"
	"goblog/register"
	"net/http"
	"strings"
	"testing"
)

func TestRefresh(t *testing.T) {
	db, _, test, public, private := Load()
	defer db.Close()

	test.Post("/register", register.Regist(db, private))
	test.Post("/test", register.Refresh(db, private, public))

	mail := "test@"
	role := "Author"
	password := "test"

	Register := fmt.Sprintf(`{"mail":"%s","role":"%s","password":"%s"}`, mail, role, password)
	req, err := http.NewRequest("POST", "/register", strings.NewReader(Register))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := test.Test(req)
	defer db.Exec("DELETE FROM users WHERE mail=$1", mail)
	if err != nil {
		t.Fatal(err)
	}

	refresh := res.Cookies()
	if len(refresh) != 1 {
		t.Fatalf("refresh cookie length %d != 1", len(refresh))
	}

	reqValid, err := http.NewRequest("POST", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	reqValid.AddCookie(refresh[0])

	res, err = test.Test(reqValid)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 200 {
		t.Errorf("status code %d != 200", res.StatusCode)
	}

	cookie := res.Cookies()
	if len(cookie) != 1 {
		t.Fatalf("cookie length %d != 1", len(cookie))
	}

	if cookie[0].Name != "refresh_token" {
		t.Errorf("cookie name %s != refresh_token", cookie[0].Name)
	}

	parts := strings.Split(cookie[0].Value, ".")
	if len(parts) != 3 {
		t.Error("cookie does not contain a JWT token")
	}

	reqEmptyCookie, err := http.NewRequest("POST", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	reqEmptyCookie.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: "",
	})

	res, err = test.Test(reqEmptyCookie)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 401 {
		t.Errorf("status code %d != 401", res.StatusCode)
	}

	reqInvalidCookie, err := http.NewRequest("POST", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	reqInvalidCookie.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: refresh[0].Value + "invalid",
	})

	res, err = test.Test(reqInvalidCookie)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 401 {
		t.Errorf("status code %d != 401", res.StatusCode)
	}

	reqInvalidTime, err := http.NewRequest("POST", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	inavalid, err := keys.Record{mail, role}.CreateJWT(0, private)
	if err != nil {
		t.Fatal(err)
	}

	reqInvalidTime.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: inavalid,
	})

	res, err = test.Test(reqInvalidTime)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 401 {
		t.Errorf("status code %d != 401", res.StatusCode)
	}

	reqInvalidMail, err := http.NewRequest("POST", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	inavalid, err = keys.Record{"invalid", role}.CreateJWT(1337, private)
	if err != nil {
		t.Fatal(err)
	}

	reqInvalidTime.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: inavalid,
	})

	res, err = test.Test(reqInvalidMail)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 401 {
		t.Errorf("status code %d != 401", res.StatusCode)
	}

}
