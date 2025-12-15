package tests

import (
	"fmt"
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

	mail := "test"
	role := "Author"
	password := "test"

	Register := fmt.Sprintf(`{"mail":"%s","role":"%s","password":"%s"}`, mail, role, password)
	req, err := http.NewRequest("POST", "/register", strings.NewReader(Register))
	if err != nil {
		t.Fatal(err)
	}
	//TODO - принять с куков нужное и копипастить проверки статусов

	_, err = test.Test(req)
	defer db.Exec("DELETE FROM users WHERE mail=$1", mail)
	if err != nil {
		t.Fatal(err)
	}
}
