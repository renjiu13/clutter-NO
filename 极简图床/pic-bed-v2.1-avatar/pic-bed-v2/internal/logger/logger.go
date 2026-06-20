package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	logFile *os.File
	logger  *log.Logger
	once    sync.Once
)

// Init 初始化日志
func Init(logPath string, enable bool) error {
	if !enable {
		return nil
	}

	var err error
	once.Do(func() {
		// 确保日志目录存在
		logDir := filepath.Dir(logPath)
		if logDir != "." {
			os.MkdirAll(logDir, 0755)
		}

		logFile, err = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return
		}
		logger = log.New(logFile, "", log.LstdFlags)
	})
	return err
}

// Log 记录操作日志
func Log(action, ip, filename, detail string) {
	if logger == nil {
		return
	}
	msg := fmt.Sprintf("[%s] IP:%s File:%s - %s", action, ip, filename, detail)
	logger.Println(msg)
}

// LogUpload 记录上传日志
func LogUpload(ip, filename, url string, size int64) {
	Log("UPLOAD", ip, filename, fmt.Sprintf("URL:%s Size:%d bytes", url, size))
}

// LogDelete 记录删除日志
func LogDelete(ip, filename string) {
	Log("DELETE", ip, filename, "File deleted")
}

// LogAccess 记录访问日志
func LogAccess(ip, filename string) {
	Log("ACCESS", ip, filename, "File accessed")
}

// LogError 记录错误日志
func LogError(ip, filename, err string) {
	Log("ERROR", ip, filename, err)
}

// GetToday 获取当前日期字符串
func GetToday() string {
	return time.Now().Format("2006-01-02")
}
