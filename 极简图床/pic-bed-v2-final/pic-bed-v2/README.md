# Pic-Bed - 极轻量私有图床

专为低内存设备设计的单文件私有图床系统，完美支持玩客云等ARM32设备。

## ✨ 核心特性

- 🚀 **极低内存占用**：闲置 8~15MB，峰值不超过 25MB
- 📦 **单文件部署**：纯静态编译，零依赖，下载即运行
- 🔧 **全架构支持**：amd64 / arm64 / armv7（玩客云32位）
- 🔒 **安全加固**：魔数校验、路径防护、请求体硬限制
- 🎯 **PicList 兼容**：完美支持 PicList 自定义图床
- ⚙️ **8大功能开关**：灵活配置，按需开启

## 🎛️ 完整功能开关

| 功能 | 配置项 | 默认值 | 说明 |
|------|--------|--------|------|
| 📝 操作日志 | `enable_log` | ✅ 开启 | 记录上传/删除/访问/错误，用于审计追踪 |
| 🗑️ 删除接口 | `enable_delete` | ✅ 开启 | `DELETE /img/{path}` 删除图片 |
| 📋 文件列表 | `enable_file_list` | ✅ 开启 | Web管理页面 `/list` 查看所有图片 |
| 🧹 自动清理 | `enable_auto_clean` | ❌ 关闭 | 定期删除超过N天的文件 |
| 🗜️ 图片压缩 | `enable_compress` | ❌ 关闭 | 自动压缩上传的图片（预留接口） |
| 📄 原始文件名 | `keep_original_name` | ❌ 关闭 | 保留图片原始文件名+随机后缀 |
| ⏱️ 请求超时 | `timeout` | 30秒 | 上传/下载超时时间，防止卡顿 |
| 📁 文件类型 | `allowed_types` | 5种 | 白名单配置，灵活控制允许格式 |

## 📦 架构版本

| 文件名 | 架构 | 适用设备 | 大小 |
|--------|------|----------|------|
| `pic-bed-linux-amd64` | x86-64 | 常规云服务器、PC | ~6.7MB |
| `pic-bed-linux-arm64` | ARM64/aarch64 | 树莓派4/5、ARM服务器 | ~6.5MB |
| `pic-bed-linux-armv7` | ARMv7 32位 | **玩客云**、树莓派3及更早 | ~6.6MB |

## 🚀 快速部署

### 1. 玩客云（ARMv7 32位）
```bash
mkdir -p /opt/pic-bed && cd /opt/pic-bed

# 下载玩客云专用版本
wget https://github.com/your-repo/releases/download/v2.0/pic-bed-linux-armv7 -O pic-bed
chmod +x pic-bed

# 首次运行自动生成配置
./pic-bed
```

### 2. systemd 开机自启
```bash
# 复制服务文件
cp examples/pic-bed.service /etc/systemd/system/

# 启动并设置开机自启
systemctl daemon-reload
systemctl enable --now pic-bed

# 查看状态
systemctl status pic-bed
journalctl -u pic-bed -f
```

## 🔧 配置说明

首次运行自动生成 `config.json`：
```json
{
  "port": 8080,
  "storage_dir": "./data",
  "max_size": 10,
  "api_key": "",
  "timeout": 30,
  "enable_log": true,
  "enable_delete": true,
  "enable_file_list": true,
  "enable_auto_clean": false,
  "enable_compress": false,
  "keep_original_name": false,
  "allowed_types": ["jpg", "jpeg", "png", "gif", "webp"],
  "auto_clean_days": 30,
  "log_file": "./pic-bed.log"
}
```

修改后重启生效：`systemctl restart pic-bed`

## 📡 API 接口

### 上传图片
```bash
curl -F "file=@test.jpg" http://服务器IP:8080/upload
```
返回：
```json
{"success":true,"url":"/img/2026/06/171846123456789abcdef.jpg","message":"upload success"}
```

### 预览图片
```
GET http://服务器IP:8080/img/2026/06/文件名.jpg
```

### 删除图片（需开启enable_delete）
```bash
curl -X DELETE http://服务器IP:8080/img/2026/06/文件名.jpg
```

### 文件列表（需开启enable_file_list）
```
http://服务器IP:8080/list          # HTML管理页面
http://服务器IP:8080/list?format=json  # JSON格式
```

## 🎯 PicList 配置

| 配置项 | 值 |
|--------|-----|
| 接口网址 | `http://服务器IP:8080/upload` |
| 请求方法 | POST |
| 表单参数名 | `file` |
| 请求头 | `{}` |
| 请求体 | `{}` |
| 自定义前缀 | `http://服务器IP:8080` |
| 网站路径 | 留空 |
| 返回数据URL路径 | `url` |

> ⚠️ 注意：PicList 部分版本不支持 `$.url` 写法，直接填 `url` 即可

## 🔐 开启鉴权

1. 在 `config.json` 中设置 `api_key` 为自定义密钥
2. PicList 请求头改为：`{"Authorization": "Bearer 你的密钥"}`

## 🧹 自动清理

1. 设置 `enable_auto_clean: true`
2. 设置 `auto_clean_days: 30`（清理超过30天的文件）
3. 重启服务，每天凌晨自动执行清理

## 📝 日志格式

日志文件 `pic-bed.log` 格式：
```
2026/06/15 22:00:00 [UPLOAD] IP:192.168.1.100 File:xxx.jpg - URL:/img/2026/06/xxx.jpg Size:123456 bytes
2026/06/15 22:00:01 [DELETE] IP:192.168.1.100 File:xxx.jpg - File deleted
2026/06/15 22:00:02 [ACCESS] IP:192.168.1.100 File:xxx.jpg - File accessed
```

## 🏗️ 项目结构

```
pic-bed-v2/
├── cmd/
│   └── pic-bed/          # 主程序入口
├── internal/
│   ├── config/           # 配置管理
│   ├── logger/           # 日志记录
│   ├── security/         # 安全校验
│   ├── storage/          # 存储与清理
│   └── handler/          # HTTP处理器
├── bin/                  # 编译产物（三架构）
├── examples/             # 示例配置
│   ├── pic-bed.service
│   └── config.example.json
├── docs/                 # 文档
├── build.sh              # 编译脚本
├── go.mod
└── README.md
```

## 🔨 自行编译

```bash
chmod +x build.sh
./build.sh
```

## 📊 性能指标

- 闲置内存：8~12MB
- 峰值内存：15~25MB
- 二进制体积：~6.5MB
- 并发：Go原生协程，支持高并发
- 依赖：纯静态，零外部依赖

## 📄 License

MIT
