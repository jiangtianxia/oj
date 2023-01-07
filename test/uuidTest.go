package main

import (
	"fmt"

	uuid "github.com/satori/go.uuid"
)

func TestGenerateUUID() {
	s := uuid.NewV4().String()
	fmt.Println(s)
}
func main() {
	TestGenerateUUID()
}
