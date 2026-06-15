package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// ValidateImageType 基于文件魔数校验图片格式，返回标准扩展名
func ValidateImageType(header []byte) (string, error) {
	if len(header) < 12 {
		return "", fmt.Errorf("file header too short")
	}

	// JPEG/JPG
	if bytes.HasPrefix(header, []byte{0xFF, 0xD8, 0xFF}) {
		return "jpg", nil
	}

	// PNG
	if bytes.HasPrefix(header, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
		return "png", nil
	}

	// GIF
	if bytes.HasPrefix(header, []byte{0x47, 0x49, 0x46, 0x38}) {
		return "gif", nil
	}

	// WebP（RIFF头 + WEBP标识）
	if bytes.HasPrefix(header, []byte{0x52, 0x49, 0x46, 0x46}) &&
		bytes.Equal(header[8:12], []byte{0x57, 0x45, 0x42, 0x50}) {
		return "webp", nil
	}

	return "", fmt.Errorf("unsupported image format")
}

// GenerateSafeFileName 生成不可预测的安全文件名
// 格式：纳秒时间戳 + 6位加密随机十六进制 + 扩展名
func GenerateSafeFileName(ext string) string {
	b := make([]byte, 6)
	rand.Read(b)
	return fmt.Sprintf("%d%s.%s",
		time.Now().UnixNano(),
		hex.EncodeToString(b),
		ext,
	)
}

// GetYearMonth 获取当前年月字符串，用于分目录存储
func GetYearMonth() (year string, month string) {
	now := time.Now()
	return fmt.Sprintf("%04d", now.Year()), fmt.Sprintf("%02d", now.Month())
}

// GetContentType 根据扩展名返回正确的HTTP Content-Type
func GetContentType(ext string) string {
	switch ext {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
