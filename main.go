package main

import (
	"log"
	"net/http"
	"os"

	"effihub/config"
	"effihub/handlers"
	"effihub/models"

	"github.com/joho/godotenv"
)

func main() {
	// 尝试加载 .env 文件（如果存在）
	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) {
			log.Printf("加载 .env 文件时出错: %v", err)
		}
	}

	// 初始化数据库
	if err := config.InitDB(); err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	defer config.DB.Close()

	// 自动建表
	if err := models.InitTables(); err != nil {
		log.Fatal("建表失败:", err)
	}

	// 路由
	mux := http.NewServeMux()

	// API 路由
	mux.HandleFunc("/api/auth/login", handlers.LoginHandler)
	mux.HandleFunc("/api/auth/upload-config", handlers.UploadConfigHandler)
	mux.HandleFunc("/api/categories", handlers.CategoriesHandler)
	mux.HandleFunc("/api/categories/", handlers.CategoryHandler)
	mux.HandleFunc("/api/links", handlers.LinksHandler)
	mux.HandleFunc("/api/links/", handlers.LinkHandler)

	// 静态文件
	mux.Handle("/", http.FileServer(http.Dir("./static/")))

	// CORS 中间件
	handler := corsMiddleware(mux)

	log.Println("服务启动在 http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
