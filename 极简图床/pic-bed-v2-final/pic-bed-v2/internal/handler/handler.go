package handler

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/pic-bed/pic-bed/internal/config"
	"github.com/pic-bed/pic-bed/internal/logger"
	"github.com/pic-bed/pic-bed/internal/security"
	"github.com/pic-bed/pic-bed/internal/storage"
)

// UploadResponse 上传响应
type UploadResponse struct {
	Success bool   `json:"success"`
	URL     string `json:"url"`
	Message string `json:"message"`
}

// ListResponse 文件列表响应
type ListResponse struct {
	Success bool               `json:"success"`
	Files   []storage.FileInfo `json:"files"`
	Count   int                `json:"count"`
}

// getClientIP 获取客户端IP
func getClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = strings.Split(r.RemoteAddr, ":")[0]
	}
	return ip
}

// HandleUpload 处理上传
func HandleUpload(w http.ResponseWriter, r *http.Request) {
	cfg := config.Get()
	ip := getClientIP(r)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		respondJSON(w, false, "", "only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	maxBytes := int64(cfg.MaxSize) * 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	if err := r.ParseMultipartForm(1 << 20); err != nil {
		logger.LogError(ip, "", "parse form failed: "+err.Error())
		respondJSON(w, false, "", "file too large or parse failed", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		logger.LogError(ip, "", "get file failed: "+err.Error())
		respondJSON(w, false, "", "missing 'file' field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 基于扩展名验证文件类型（移除魔数验证）
	ext, err := security.ValidateFileType(fileHeader.Filename, cfg.AllowedTypes)
	if err != nil {
		logger.LogError(ip, fileHeader.Filename, "invalid format: "+err.Error())
		respondJSON(w, false, "", err.Error(), http.StatusBadRequest)
		return
	}

	year, month := security.GetYearMonth()
	fileName := security.GenerateFileName(fileHeader.Filename, ext, cfg.KeepOriginalName)

	relativeURL, err := storage.SaveFile(file, cfg.StorageDir, year, month, fileName)
	if err != nil {
		logger.LogError(ip, fileName, "save failed: "+err.Error())
		respondJSON(w, false, "", "save failed", http.StatusInternalServerError)
		return
	}

	logger.LogUpload(ip, fileName, relativeURL, fileHeader.Size)
	respondJSON(w, true, relativeURL, "upload success", http.StatusOK)
}

// HandleImage 统一处理图片相关请求（GET预览 / DELETE删除）
// 解决Go标准库同一路径不能多次注册的问题
func HandleImage(w http.ResponseWriter, r *http.Request) {
	cfg := config.Get()
	ip := getClientIP(r)
	w.Header().Set("Access-Control-Allow-Origin", "*")

	pathSuffix := strings.TrimPrefix(r.URL.Path, "/img/")
	parts := strings.SplitN(pathSuffix, "/", 3)
	if len(parts) != 3 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	year, month, fileName := parts[0], parts[1], parts[2]

	// 路径安全校验
	if strings.ContainsAny(year, "./\\") || strings.ContainsAny(month, "./\\") || strings.ContainsAny(fileName, "/\\") {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	fullPath := filepath.Join(cfg.StorageDir, year, month, fileName)
	if !security.IsPathSafe(cfg.StorageDir, fullPath) {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	// 根据HTTP方法分发处理
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		// GET/HEAD: 预览图片
		logger.LogAccess(ip, fileName)
		http.ServeFile(w, r, fullPath)

	case http.MethodDelete:
		// DELETE: 删除图片（需开启开关）
		w.Header().Set("Content-Type", "application/json")
		if !cfg.EnableDelete {
			respondJSON(w, false, "", "delete disabled", http.StatusForbidden)
			return
		}
		if err := storage.DeleteFile(cfg.StorageDir, year, month, fileName); err != nil {
			logger.LogError(ip, fileName, "delete failed: "+err.Error())
			respondJSON(w, false, "", err.Error(), http.StatusNotFound)
			return
		}
		logger.LogDelete(ip, fileName)
		respondJSON(w, true, "", "delete success", http.StatusOK)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleFileList 文件列表页面
func HandleFileList(w http.ResponseWriter, r *http.Request) {
	cfg := config.Get()

	if !cfg.EnableFileList {
		http.Error(w, "file list disabled", http.StatusForbidden)
		return
	}

	if r.URL.Query().Get("format") == "json" {
		files, err := storage.ListFiles(cfg.StorageDir)
		if err != nil {
			respondJSON(w, false, nil, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ListResponse{
			Success: true,
			Files:   files,
			Count:   len(files),
		})
		return
	}

	// HTML 页面
	files, _ := storage.ListFiles(cfg.StorageDir)
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>图片管理</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 1200px; margin: 0 auto; padding: 20px; }
        .grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 15px; }
        .item { border: 1px solid #eee; border-radius: 8px; padding: 10px; }
        .item img { width: 100%; height: 150px; object-fit: cover; border-radius: 4px; }
        .info { font-size: 12px; color: #666; margin-top: 8px; }
    </style>
</head>
<body>
    <h1>图片管理 ({{.Count}} 张)</h1>
    <div class="grid">
        {{range .Files}}
        <div class="item">
            <img src="{{.Path}}" alt="{{.Name}}">
            <div class="info">{{.Name}}<br>{{.Size}} bytes</div>
        </div>
        {{end}}
    </div>
</body>
</html>`
	t, _ := template.New("list").Parse(tmpl)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t.Execute(w, map[string]interface{}{
		"Files": files,
		"Count": len(files),
	})
}

// AuthMiddleware 鉴权中间件
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := config.Get()
		if cfg.APIKey == "" {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondJSON(w, false, "", "missing Authorization", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" || parts[1] != cfg.APIKey {
			respondJSON(w, false, "", "invalid API key", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func respondJSON(w http.ResponseWriter, success bool, url interface{}, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": success,
		"url":     url,
		"message": msg,
	})
}
