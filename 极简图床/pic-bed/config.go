package main

import (
	"encoding/json"
	"os"
)

// Config 全局配置结构体
type Config struct {
	Port       int    `json:"port"`
	StorageDir string `json:"storage_dir"`
	MaxSize    int    `json:"max_size"` // 单位：MB
	APIKey     string `json:"api_key"`  // 留空则关闭鉴权
}

const configFileName = "config.json"

var globalCfg Config

// 默认配置
var defaultConfig = Config{
	Port:       8080,
	StorageDir: "./data",
	MaxSize:    10,
	APIKey:     "",
}

// InitConfig 初始化配置：首次运行自动生成配置文件，全局加载一次
func InitConfig() error {
	// 配置文件不存在则创建默认配置
	if _, err := os.Stat(configFileName); os.IsNotExist(err) {
		data, err := json.MarshalIndent(defaultConfig, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(configFileName, data, 0644); err != nil {
			return err
		}
		globalCfg = defaultConfig
		return nil
	}

	// 读取并解析配置
	data, err := os.ReadFile(configFileName)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &globalCfg)
}

// GetConfig 获取全局配置
func GetConfig() Config {
	return globalCfg
}
