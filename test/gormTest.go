package main

import (
	"fmt"
	"oj/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Test() {
	dsn := "root:124578@tcp(127.0.0.1:3306)/oj?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	data := make([]*models.ProblemBasic, 0)
	err = db.Find(&data).Error
	if err != nil {
		panic(err)
	}

	for _, v := range data {
		fmt.Printf("Problem ==> %v \n", v)
	}
	fmt.Println("ok")
}
