package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// SaveImageStream 流式保存图片到年月目录，全程内存占用恒定
func SaveImageStream(reader io.Reader, baseDir, year, month, fileName string) (string, error) {
	// 创建年月二级目录
	targetDir := filepath.Join(baseDir, year, month)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("create dir failed: %w", err)
	}

	fullPath := filepath.Join(targetDir, fileName)

	// 创建目标文件
	outFile, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("create file failed: %w", err)
	}
	defer outFile.Close()

	// 流式拷贝，内存占用恒定（默认32KB缓冲区）
	if _, err := io.Copy(outFile, reader); err != nil {
		os.Remove(fullPath) // 写入失败清理不完整文件
		return "", fmt.Errorf("write file failed: %w", err)
	}

	// 返回访问相对路径
	return fmt.Sprintf("/img/%s/%s/%s", year, month, fileName), nil
}

// IsPathSafe 校验文件路径是否在存储目录内，防止路径遍历攻击
func IsPathSafe(baseDir, targetPath string) bool {
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return false
	}
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return false
	}

	// 确保目标路径以存储根目录为前缀
	rel, err := filepath.Rel(absBase, absTarget)
	if err != nil {
		return false
	}

	// 相对路径不能跳出根目录
	if len(rel) >= 2 && rel[:2] == ".." {
		return false
	}
	return true
}
