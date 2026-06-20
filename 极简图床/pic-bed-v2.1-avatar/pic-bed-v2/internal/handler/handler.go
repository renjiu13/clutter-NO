package handler

import (
	"encoding/json"
	"fmt"
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

// HandleHome 首页欢迎页面（Jubilee风格）
func HandleHome(w http.ResponseWriter, r *http.Request) {
	// 只处理根路径
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	cfg := config.Get()
	host := r.Host
	avatarURL := cfg.HomeAvatarURL
	
	tmpl := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Welcome to awang! :)</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            min-height: 100vh;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            padding: 40px 20px;
            background: #fff;
        }

        .container {
            text-align: center;
            max-width: 500px;
        }

        h1 {
            font-size: 2rem;
            font-weight: 700;
            color: #000;
            margin-bottom: 20px;
        }

        .avatar {
            width: 300px;
            height: 300px;
            margin: 0 auto 20px;
            border-radius: 20px;
            overflow: hidden;
            display: flex;
            align-items: center;
            justify-content: center;
            position: relative;
        }

        .avatar.default {
            background: linear-gradient(135deg, #ff6b9d 0%, #ffa3c4 100%);
            font-size: 120px;
        }

        .avatar.default::before {
            content: "✨";
            position: absolute;
            top: 20px;
            left: 30px;
            font-size: 40px;
            animation: twinkle 2s ease-in-out infinite;
        }

        .avatar.default::after {
            content: "⭐";
            position: absolute;
            top: 50px;
            right: 40px;
            font-size: 25px;
            animation: twinkle 2s ease-in-out infinite 0.5s;
        }

        .avatar img {
            width: 100%;
            height: 100%;
            object-fit: cover;
        }

        .avatar .face {
            font-size: 150px;
            z-index: 1;
        }

        @keyframes twinkle {
            0%, 100% { opacity: 1; transform: scale(1); }
            50% { opacity: 0.5; transform: scale(0.8); }
        }

        p {
            font-size: 1.1rem;
            color: #333;
            margin-bottom: 10px;
            line-height: 1.6;
        }

        .ip {
            font-size: 1rem;
            color: #666;
            font-family: monospace;
            margin-top: 10px;
        }

        .footer {
            margin-top: 40px;
            font-size: 0.85rem;
            color: #999;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Welcome to awang! :)</h1>
        
        {{if .AvatarURL}}
        <div class="avatar">
            <img src="{{.AvatarURL}}" alt="avatar">
        </div>
        {{else}}
        <div class="avatar default">
            <span class="face">😎</span>
        </div>
        {{end}}

        <p>I serve photos. You'll need a URL.</p>
        <p></p>
        <p class="ip">{{.Host}}</p>
    </div>

    <div class="footer">
        Pic Bed · 极轻量私有图床
    </div>
</body>
</html>`

	t, _ := template.New("home").Parse(tmpl)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t.Execute(w, map[string]interface{}{
		"Host":      host,
		"AvatarURL": avatarURL,
	})
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

	// HTML 页面 - 美化版
	files, _ := storage.ListFiles(cfg.StorageDir)
	tmpl := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>🖼️ Pic Bed - 图片管理</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 40px 20px;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
        }

        /* 头部 */
        .header {
            text-align: center;
            color: white;
            margin-bottom: 40px;
        }

        .header h1 {
            font-size: 2.5rem;
            font-weight: 700;
            margin-bottom: 10px;
            text-shadow: 0 2px 10px rgba(0,0,0,0.2);
        }

        .header p {
            font-size: 1.1rem;
            opacity: 0.9;
        }

        /* 统计卡片 */
        .stats {
            display: flex;
            justify-content: center;
            gap: 30px;
            margin-bottom: 40px;
            flex-wrap: wrap;
        }

        .stat-card {
            background: rgba(255,255,255,0.15);
            backdrop-filter: blur(10px);
            border-radius: 16px;
            padding: 20px 40px;
            color: white;
            text-align: center;
            border: 1px solid rgba(255,255,255,0.2);
        }

        .stat-card .number {
            font-size: 2rem;
            font-weight: 700;
        }

        .stat-card .label {
            font-size: 0.9rem;
            opacity: 0.8;
            margin-top: 5px;
        }

        /* 图片网格 */
        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
            gap: 20px;
        }

        .card {
            background: white;
            border-radius: 16px;
            overflow: hidden;
            box-shadow: 0 4px 15px rgba(0,0,0,0.1);
            transition: all 0.3s ease;
            cursor: pointer;
        }

        .card:hover {
            transform: translateY(-8px);
            box-shadow: 0 12px 30px rgba(0,0,0,0.2);
        }

        .card .img-wrapper {
            width: 100%;
            height: 180px;
            overflow: hidden;
            background: #f5f5f5;
        }

        .card img {
            width: 100%;
            height: 100%;
            object-fit: cover;
            transition: transform 0.3s ease;
        }

        .card:hover img {
            transform: scale(1.05);
        }

        .card .info {
            padding: 15px;
        }

        .card .filename {
            font-size: 0.85rem;
            color: #333;
            font-weight: 500;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
            margin-bottom: 8px;
        }

        .card .meta {
            display: flex;
            justify-content: space-between;
            font-size: 0.75rem;
            color: #999;
        }

        .card .actions {
            display: flex;
            gap: 8px;
            margin-top: 12px;
        }

        .btn {
            flex: 1;
            padding: 8px 12px;
            border: none;
            border-radius: 8px;
            font-size: 0.8rem;
            cursor: pointer;
            transition: all 0.2s ease;
            font-weight: 500;
        }

        .btn-copy {
            background: #667eea;
            color: white;
        }

        .btn-copy:hover {
            background: #5a6fd6;
        }

        .btn-copy:active {
            transform: scale(0.95);
        }

        /* 空状态 */
        .empty {
            text-align: center;
            color: white;
            padding: 80px 20px;
        }

        .empty .icon {
            font-size: 4rem;
            margin-bottom: 20px;
        }

        .empty p {
            font-size: 1.2rem;
            opacity: 0.9;
        }

        /* 复制成功提示 */
        .toast {
            position: fixed;
            bottom: 30px;
            left: 50%;
            transform: translateX(-50%) translateY(100px);
            background: rgba(0,0,0,0.8);
            color: white;
            padding: 12px 24px;
            border-radius: 30px;
            font-size: 0.9rem;
            opacity: 0;
            transition: all 0.3s ease;
            z-index: 1000;
        }

        .toast.show {
            transform: translateX(-50%) translateY(0);
            opacity: 1;
        }

        /* 页脚 */
        .footer {
            text-align: center;
            color: rgba(255,255,255,0.6);
            margin-top: 60px;
            font-size: 0.85rem;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🖼️ Welcome to Pic Bed! :)</h1>
            <p>I serve photos. Your personal image hosting service.</p>
        </div>

        <div class="stats">
            <div class="stat-card">
                <div class="number">{{.Count}}</div>
                <div class="label">张图片</div>
            </div>
            <div class="stat-card">
                <div class="number">{{.TotalSize}}</div>
                <div class="label">总大小</div>
            </div>
        </div>

        {{if .Files}}
        <div class="grid">
            {{range .Files}}
            <div class="card" onclick="copyUrl('{{.Path}}')">
                <div class="img-wrapper">
                    <img src="{{.Path}}" alt="{{.Name}}" loading="lazy">
                </div>
                <div class="info">
                    <div class="filename" title="{{.Name}}">{{.Name}}</div>
                    <div class="meta">
                        <span>{{.SizeStr}}</span>
                        <span>{{.TimeStr}}</span>
                    </div>
                    <div class="actions">
                        <button class="btn btn-copy" onclick="event.stopPropagation(); copyUrl('{{.Path}}')">
                            📋 复制链接
                        </button>
                    </div>
                </div>
            </div>
            {{end}}
        </div>
        {{else}}
        <div class="empty">
            <div class="icon">📷</div>
            <p>还没有图片，快去上传第一张吧！</p>
        </div>
        {{end}}

        <div class="footer">
            Pic Bed · 极轻量私有图床 · Powered by Go
        </div>
    </div>

    <div class="toast" id="toast">✅ 链接已复制到剪贴板</div>

    <script>
        function copyUrl(path) {
            const url = window.location.origin + path;
            navigator.clipboard.writeText(url).then(() => {
                showToast();
            }).catch(() => {
                // 降级方案
                const input = document.createElement('input');
                input.value = url;
                document.body.appendChild(input);
                input.select();
                document.execCommand('copy');
                document.body.removeChild(input);
                showToast();
            });
        }

        function showToast() {
            const toast = document.getElementById('toast');
            toast.classList.add('show');
            setTimeout(() => {
                toast.classList.remove('show');
            }, 2000);
        }
    </script>
</body>
</html>`
	// 预处理文件数据（格式化大小和时间）
	type fileView struct {
		Name    string
		Path    string
		SizeStr string
		TimeStr string
	}
	
	var totalSize int64
	var fileViews []fileView
	for _, f := range files {
		totalSize += f.Size
		fileViews = append(fileViews, fileView{
			Name:    f.Name,
			Path:    f.Path,
			SizeStr: formatSize(f.Size),
			TimeStr: f.ModTime.Format("2006-01-02 15:04"),
		})
	}

	t, _ := template.New("list").Parse(tmpl)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t.Execute(w, map[string]interface{}{
		"Files":     fileViews,
		"Count":     len(files),
		"TotalSize": formatSize(totalSize),
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

// formatSize 格式化文件大小
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
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
