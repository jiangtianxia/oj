package main

import (
	"fmt"
	"testing"

	"github.com/dgrijalva/jwt-go"
)

type UserClaims struct {
	Identity string `json:"identity"`
	Name     string `json:"name"`
	jwt.StandardClaims
}

var myKey = []byte("h2wnknlsd")

// 生成token
func TestGenerateToken(t *testing.T) {
	userClaim := &UserClaims{
		Identity:       "user_1",
		Name:           "test",
		StandardClaims: jwt.StandardClaims{},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaim)
	tokenString, err := token.SignedString(myKey)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(tokenString)
}

// 解析token
func TestAnalyseToken(t *testing.T) {
	tokenString := "er"
	UserClaims := new(UserClaims)

	claims, err := jwt.ParseWithClaims(tokenString, UserClaims, func(token *jwt.Token) (interface{}, error) {
		return myKey, nil
	})

	if err != nil {
		t.Fatal(err)
	}

	if claims.Valid {
		fmt.Println(UserClaims)
	}
}
