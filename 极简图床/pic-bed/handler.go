package main

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

// UploadResponse 统一上传响应结构
type UploadResponse struct {
	Success bool   `json:"success"`
	URL     string `json:"url"`
	Message string `json:"message"`
}

// HandleUpload 处理图片上传
func HandleUpload(w http.ResponseWriter, r *http.Request) {
	// 基础CORS头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		respondJSON(w, false, "", "only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg := GetConfig()
	maxBytes := int64(cfg.MaxSize) * 1024 * 1024

	// 底层限制请求体大小，防止伪造Content-Length打满磁盘
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	// 解析表单：1MB以内放内存，超出自动写临时文件
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		respondJSON(w, false, "", "file exceeds size limit or parse failed", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		respondJSON(w, false, "", "missing 'file' field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 读取文件头做魔数校验
	headerBuf := make([]byte, 512)
	n, err := file.Read(headerBuf)
	if err != nil && err != io.EOF {
		respondJSON(w, false, "", "read file failed", http.StatusInternalServerError)
		return
	}

	// 校验真实文件格式
	ext, err := ValidateImageType(headerBuf[:n])
	if err != nil {
		respondJSON(w, false, "", "only jpg/png/gif/webp supported", http.StatusBadRequest)
		return
	}

	// 重置文件指针到开头，准备完整写入
	if _, err := file.Seek(0, 0); err != nil {
		respondJSON(w, false, "", "file reset failed", http.StatusInternalServerError)
		return
	}

	// 生成存储路径和文件名
	year, month := GetYearMonth()
	fileName := GenerateSafeFileName(ext)

	// 流式写入磁盘
	relativeURL, err := SaveImageStream(file, cfg.StorageDir, year, month, fileName)
	if err != nil {
		respondJSON(w, false, "", "save file failed", http.StatusInternalServerError)
		return
	}

	respondJSON(w, true, relativeURL, "upload success", http.StatusOK)
}

// HandleImagePreview 处理图片直链预览，流式输出，零内存膨胀
func HandleImagePreview(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg := GetConfig()

	// 解析路径：/img/年/月/文件名
	pathSuffix := strings.TrimPrefix(r.URL.Path, "/img/")
	parts := strings.SplitN(pathSuffix, "/", 3)
	if len(parts) != 3 {
		http.Error(w, "invalid image path", http.StatusBadRequest)
		return
	}
	year, month, fileName := parts[0], parts[1], parts[2]

	// 第一层防护：路径段禁止包含特殊字符
	if strings.ContainsAny(year, "./\\") || strings.ContainsAny(month, "./\\") || strings.ContainsAny(fileName, "/\\") {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	// 拼接完整路径
	fullPath := filepath.Join(cfg.StorageDir, year, month, fileName)

	// 第二层防护：绝对路径校验，防止路径遍历
	if !IsPathSafe(cfg.StorageDir, fullPath) {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	// 标准库流式输出：原生支持断点续传、缓存协商、自动Content-Type
	// 全程内核级零拷贝，内存占用恒定在KB级
	http.ServeFile(w, r, fullPath)
}

// 统一JSON响应工具
func respondJSON(w http.ResponseWriter, success bool, url, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(UploadResponse{
		Success: success,
		URL:     url,
		Message: msg,
	})
}
