package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/pic-bed/pic-bed/internal/config"
	"github.com/pic-bed/pic-bed/internal/handler"
	"github.com/pic-bed/pic-bed/internal/logger"
	"github.com/pic-bed/pic-bed/internal/storage"
)

func main() {
	// 加载配置文件，不存在则自动生成默认配置
	if err := config.InitConfig(); err != nil {
		panic("配置文件加载失败: " + err.Error())
	}

	cfg := config.Get()

	// 初始化日志系统（可通过配置开关控制）
	if cfg.EnableLog {
		if err := logger.Init(cfg.LogFile, true); err != nil {
			fmt.Printf("警告: 日志初始化失败: %v\n", err)
		}
	}

	// 启动自动过期清理协程（可通过配置开关控制）
	if cfg.EnableAutoClean && cfg.AutoCleanHours > 0 {
		storage.StartAutoClean(cfg.StorageDir, cfg.AutoCleanHours)
		fmt.Printf("[自动清理] 已启用，将清理超过 %d 小时的文件\n", cfg.AutoCleanHours)
	}

	// 注册HTTP路由
	// 首页欢迎页面
	http.HandleFunc("/", handler.HandleHome)
	// 上传接口，带可选Bearer Token鉴权
	http.HandleFunc("/upload", handler.AuthMiddleware(handler.HandleUpload))
	// 统一处理图片请求（GET预览 / DELETE删除）
	// 解决Go标准库同一路径不能多次注册的问题
	http.HandleFunc("/img/", handler.HandleImage)
	// Web文件列表管理页面（可开关控制）
	http.HandleFunc("/list", handler.HandleFileList)

	// 启动成功提示
	fmt.Println("=== 极轻量图床启动成功 ===")
	fmt.Printf("监听端口: %d\n", cfg.Port)
	fmt.Printf("存储目录: %s\n", cfg.StorageDir)
	fmt.Printf("单文件上限: %d MB\n", cfg.MaxSize)
	fmt.Printf("上传接口: POST http://0.0.0.0:%d/upload\n", cfg.Port)
	fmt.Printf("预览格式: GET  http://服务器IP:%d/img/年/月/文件名\n", cfg.Port)

	// 功能开关状态显示
	if cfg.EnableFileList {
		fmt.Printf("文件列表: http://服务器IP:%d/list\n", cfg.Port)
	}

	if cfg.APIKey != "" {
		fmt.Println("鉴权状态: 已开启 Bearer Token 鉴权")
	} else {
		fmt.Println("鉴权状态: 未开启（公开上传）")
	}

	fmt.Printf("功能开关: 日志=%v 删除=%v 列表=%v 自动清理=%v\n",
		cfg.EnableLog, cfg.EnableDelete, cfg.EnableFileList, cfg.EnableAutoClean)

	// 启动HTTP服务，设置超时防止卡住
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		ReadTimeout:  time.Duration(cfg.Timeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Timeout) * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		panic("服务启动失败: " + err.Error())
	}
}
