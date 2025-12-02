package keys

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type DBrecord struct {
	Mail string `json:"mail"`
	Role string `json:"role"`
}

func loadKey(filename string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func (src DBrecord) CreateJWT(ttl int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": src.Mail,
		"role":    src.Role,
		"exp":     time.Now().Add(time.Hour * time.Duration(ttl)).Unix(),
	}
	key, err := loadKey("./private.pem")
	if err != nil {
		return "", err
	}
	return jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)
}
func chekAccess(acess string) bool {

}
