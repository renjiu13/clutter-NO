package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// 1. 初始化全局配置
	if err := InitConfig(); err != nil {
		log.Fatalf("config init failed: %v", err)
	}
	cfg := GetConfig()

	// 2. 确保存储根目录存在
	if err := os.MkdirAll(cfg.StorageDir, 0755); err != nil {
		log.Fatalf("create storage dir failed: %v", err)
	}

	// 3. 注册路由
	http.HandleFunc("/upload", AuthMiddleware(HandleUpload))
	http.HandleFunc("/img/", HandleImagePreview)

	// 4. 启动服务
	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("=== 极轻量图床启动成功 ===")
	log.Printf("监听端口: %d", cfg.Port)
	log.Printf("存储目录: %s", cfg.StorageDir)
	log.Printf("单文件上限: %d MB", cfg.MaxSize)
	log.Printf("上传接口: POST http://0.0.0.0:%d/upload", cfg.Port)
	log.Printf("预览格式: GET  http://服务器IP:%d/img/年/月/文件名", cfg.Port)
	if cfg.APIKey != "" {
		log.Printf("鉴权状态: 已开启 Bearer Token")
	} else {
		log.Printf("鉴权状态: 未开启（公开上传）")
	}

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server start failed: %v", err)
	}
}
