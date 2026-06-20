package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// SaveFile 流式保存文件
func SaveFile(reader io.Reader, baseDir, year, month, fileName string) (string, error) {
	targetDir := filepath.Join(baseDir, year, month)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("create dir failed: %w", err)
	}

	fullPath := filepath.Join(targetDir, fileName)
	outFile, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("create file failed: %w", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, reader); err != nil {
		os.Remove(fullPath)
		return "", fmt.Errorf("write file failed: %w", err)
	}

	return fmt.Sprintf("/img/%s/%s/%s", year, month, fileName), nil
}

// DeleteFile 删除文件
func DeleteFile(baseDir, year, month, fileName string) error {
	filePath := filepath.Join(baseDir, year, month, fileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found")
	}
	return os.Remove(filePath)
}

// ListFiles 列出所有图片文件
func ListFiles(baseDir string) ([]FileInfo, error) {
	var files []FileInfo

	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(baseDir, path)
		files = append(files, FileInfo{
			Name:    info.Name(),
			Path:    "/img/" + filepath.ToSlash(rel),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
		return nil
	})

	return files, err
}

// FileInfo 文件信息结构
type FileInfo struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
}

// CleanOldFiles 自动清理超过指定小时数的文件
func CleanOldFiles(baseDir string, hours int) (int, int64, error) {
	cutoff := time.Now().Add(-time.Duration(hours) * time.Hour)
	deletedCount := 0
	deletedSize := int64(0)

	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if info.ModTime().Before(cutoff) {
			size := info.Size()
			if err := os.Remove(path); err == nil {
				deletedCount++
				deletedSize += size
			}
		}
		return nil
	})

	return deletedCount, deletedSize, err
}

// StartAutoClean 启动自动清理协程
func StartAutoClean(baseDir string, hours int) {
	if hours <= 0 {
		return
	}
	go func() {
		ticker := time.NewTicker(1 * time.Hour) // 每小时检查一次
		defer ticker.Stop()
		for range ticker.C {
			count, size, _ := CleanOldFiles(baseDir, hours)
			if count > 0 {
				fmt.Printf("[AutoClean] Deleted %d files, freed %d bytes\n", count, size)
			}
		}
	}()
}
