package main

import (
	"fmt"
	"net/http"
)

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// 如果是 OPTIONS 请求，则直接返回
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// 静态文件的目录
	fs := http.FileServer(http.Dir("./"))

	// 处理静态文件
	http.Handle("/", fs)

	port := "8081"

	fmt.Println("Server is running on http://0.0.0.0:" + port)
	http.ListenAndServe("0.0.0.0:"+port, CORS(http.DefaultServeMux))
}
