package main

import (
	"crypto/tls"
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
)

func SendEmailTest() {
	e := email.NewEmail()

	mailUserName := "jiang2381385276@163.com" //邮箱账号
	mailPassword := "OSJXVUTKLANNJZIP"        //邮箱授权码
	code := "12345678"                        //发送的验证码
	Subject := "验证码发送测试"                      //发送的主题

	e.From = "Get <jiang2381385276@163.com>"
	e.To = []string{"2381385276@qq.com"}
	e.Subject = Subject
	e.HTML = []byte("你的验证码为：<h1>" + code + "</h1>")
	err := e.SendWithTLS("smtp.163.com:465", smtp.PlainAuth("", mailUserName, mailPassword, "smtp.163.com"),
		&tls.Config{InsecureSkipVerify: true, ServerName: "smtp.163.com"})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("ok")
}
