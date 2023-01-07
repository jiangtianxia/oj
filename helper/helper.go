package helper

import (
	"crypto/md5"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/smtp"
	"os"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jordan-wright/email"
	uuid "github.com/satori/go.uuid"
)

// 生成md5
func GetMd5(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

type UserClaims struct {
	Identity string `json:"identity"`
	Name     string `json:"name"`
	IsAdmin  int    `json:"is_admin"`
	jwt.StandardClaims
}

var myKey = []byte("h2wnknlsd")

// 生成token
func GenerateToken(identity, name string, isAdmin int) (string, error) {
	userClaim := &UserClaims{
		Identity:       identity,
		Name:           name,
		IsAdmin:        isAdmin,
		StandardClaims: jwt.StandardClaims{},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaim)
	tokenString, err := token.SignedString(myKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// 解析token
func AnalyseToken(tokenString string) (*UserClaims, error) {
	UserClaims := new(UserClaims)

	claims, err := jwt.ParseWithClaims(tokenString, UserClaims, func(token *jwt.Token) (interface{}, error) {
		return myKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !claims.Valid {
		return nil, fmt.Errorf("analyse Token Error:%v", err)
	}
	return UserClaims, nil
}

// 发送验证码
func SendCode(toUserEmail, code string) error {
	e := email.NewEmail()

	mailUserName := "jiang2381385276@163.com" //邮箱账号
	mailPassword := "OSJXVUTKLANNJZIP"        //邮箱授权码
	Subject := "验证码"                          //发送的主题

	e.From = "Get <jiang2381385276@163.com>"
	e.To = []string{toUserEmail}
	e.Subject = Subject
	e.HTML = []byte("本次请求的验证码为：" + code + ", 5分钟内有效")
	err := e.SendWithTLS("smtp.163.com:465", smtp.PlainAuth("", mailUserName, mailPassword, "smtp.163.com"),
		&tls.Config{InsecureSkipVerify: true, ServerName: "smtp.163.com"})
	return err
}

// 生成唯一码
func GetUUID() string {
	return uuid.NewV4().String()
}

// 生成验证码
func GetRand() string {
	rand.Seed(time.Now().UnixNano())
	s := ""

	for i := 0; i < 6; i++ {
		s += strconv.Itoa(rand.Intn(10))
	}
	return s
}

// 保存代码
func CodeSave(code []byte) (string, error) {
	dirName := "code/" + GetUUID()
	path := dirName + "/main.go"
	err := os.Mkdir(dirName, 0777)
	if err != nil {
		return "", err
	}

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	f.Write(code)
	defer f.Close()
	return path, nil
}
