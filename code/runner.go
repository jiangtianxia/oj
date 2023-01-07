package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

func main() {
	// go run codeUser/main.go

	wd, _ := os.Getwd()
	// 输出目录，看看路径对不对
	fmt.Println("工作目录: " + wd)
	cmd := exec.Command("go", "run", wd+"/code/653d53f3-6240-4c36-b1c7-242c2dbb7b73/main.go")
	var out, stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &out

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalln(err)
	}
	io.WriteString(stdinPipe, "23 11\n")

	// 根据测试的输入案例进行运行，拿到输出结果和标准的输出结果进行比对
	if err := cmd.Run(); err != nil {
		log.Fatalln(err, stderr.String())
	}
	fmt.Println(out.String())
}
