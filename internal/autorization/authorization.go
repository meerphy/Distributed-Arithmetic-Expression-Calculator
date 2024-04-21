package authorization

import (
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             int64
	Name           string
	Password       string
	OriginPassword string
}

const hmacSampleSecret = "super_secret_signature"

func (u User) ComparePassword(u2 User) error {
	err := compare(u2.Password, u.OriginPassword)
	if err != nil {
		log.Println("auth fail")
		return err
	}

	log.Println("auth success")
	return nil
}

func generate(s string) (string, error) {
	saltedBytes := []byte(s)
	hashedBytes, err := bcrypt.GenerateFromPassword(saltedBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	hash := string(hashedBytes[:])
	return hash, nil
}

func compare(hash string, s string) error {
	incoming := []byte(s)
	existing := []byte(hash)
	return bcrypt.CompareHashAndPassword(existing, incoming)
}

func MakeUser(username, password string) (User, error) {
	hashedPassword, err := generate(password)
	if err != nil {
		return User{}, err
	}
	return User{
		Name:     username,
		Password: hashedPassword,
	}, nil
}

func MakeToken(id int64, login string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":   id,
		"name": login,
		"nbf":  time.Now().Unix(),
		"exp":  time.Now().Add(5 * time.Minute).Unix(),
		"iat":  time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte(hmacSampleSecret))
	if err != nil {
		log.Fatal(err)
	}
	return tokenString
}

func GetTokenValue(token string) User {
	tokenFromString, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Fatal("Error with token")
		}

		return []byte(hmacSampleSecret), nil
	})
	if err != nil {
		log.Fatal(err)
	}
	claims, _ := tokenFromString.Claims.(jwt.MapClaims)
	return User{ID: int64(claims["id"].(float64)), Name: claims["name"].(string)}
}
