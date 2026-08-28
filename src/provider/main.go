package main

import (
	"log"
	"net/http"
	"os"

	"github.com/quanttide/qtcloud-connect/provider/api"
	"github.com/quanttide/qtcloud-connect/provider/storage"
)

func main() {
	// 获取数据库路径
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/qtcloud-connect.db"
	}

	// 初始化存储
	store, err := storage.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// 创建路由器
	router := api.NewRouter(store)

	// 获取端口
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	// 启动服务器
	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
