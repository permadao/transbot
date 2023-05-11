package main

import (
	"fmt"
	"net/http"
)

func main() {
	// 静态文件的目录
	fs := http.FileServer(http.Dir("./"))

	// 处理静态文件
	http.Handle("/", fs)

	port := "8081"

	fmt.Println("Server is running on http://localhost:" + port)
	http.ListenAndServe(":"+port, nil)
}
