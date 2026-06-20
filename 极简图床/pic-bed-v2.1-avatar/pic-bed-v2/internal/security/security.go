package security

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// GetYearMonth 获取当前年月用于分目录存储
func GetYearMonth() (string, string) {
	now := time.Now()
	return fmt.Sprintf("%04d", now.Year()), fmt.Sprintf("%02d", now.Month())
}

// GenerateFileName 生成文件名
// keepOriginalName=true: 原始文件名_随机4位.后缀（保证不重名）
// keepOriginalName=false: 纯随机16位文件名.后缀
func GenerateFileName(originalName, ext string, keepOriginalName bool) string {
	// 生成8字节随机数 = 16位十六进制
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	randomStr := hex.EncodeToString(randomBytes)

	if keepOriginalName && originalName != "" {
		// 提取纯文件名（不含路径和扩展名）
		base := filepath.Base(originalName)
		base = strings.TrimSuffix(base, filepath.Ext(base))
		// 清理不安全字符
		base = sanitizeFileName(base)
		if base != "" {
			return fmt.Sprintf("%s_%s.%s", base, randomStr[:4], ext)
		}
	}
	// 默认：纯随机文件名
	return fmt.Sprintf("%s.%s", randomStr, ext)
}

// sanitizeFileName 清理文件名中的特殊字符和危险字符
func sanitizeFileName(name string) string {
	// 只保留安全字符
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '_' || r == '.' {
			return r
		}
		return '_'
	}, name)
}

// IsPathSafe 路径安全校验，防止路径遍历攻击
func IsPathSafe(baseDir, targetPath string) bool {
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return false
	}
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(absBase, absTarget)
	if err != nil {
		return false
	}
	return !(len(rel) >= 2 && rel[:2] == "..")
}

// GetFileExt 获取文件扩展名（小写，不带点）
func GetFileExt(filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return ""
	}
	return strings.ToLower(strings.TrimPrefix(ext, "."))
}

// ValidateFileType 验证文件类型是否在白名单中（基于扩展名）
func ValidateFileType(filename string, allowedTypes []string) (string, error) {
	ext := GetFileExt(filename)
	if ext == "" {
		return "", fmt.Errorf("无法识别文件扩展名")
	}

	// 检查白名单
	for _, allowed := range allowedTypes {
		if strings.EqualFold(ext, allowed) {
			return ext, nil
		}
	}
	return "", fmt.Errorf("不允许的文件类型: %s", ext)
}
