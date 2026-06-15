# 极轻量私有图床系统

## ✨ 特性
- 🚀 **极低内存**：闲置 8~12MB，峰值不超过 25MB
- 📦 **单文件部署**：纯静态二进制，零依赖
- 🔧 **全架构支持**：amd64 / arm64 / armv7（玩客云32位）
- 🔒 **安全加固**：魔数校验、路径防护、请求体硬限制
- 📂 **自动分目录**：按年/月自动分类存储
- 🔑 **可选鉴权**：Bearer Token 鉴权
- 🎯 **PicList 兼容**：完美支持 PicList 自定义图床

---

## 📦 架构版本说明

| 文件名 | 架构 | 适用设备 |
|--------|------|----------|
| `pic-bed-linux-amd64` | x86-64 | 常规云服务器、PC |
| `pic-bed-linux-arm64` | ARM64/aarch64 | 树莓派4/5、ARM服务器、飞牛NAS |
| `pic-bed-linux-armv7` | ARMv7 32位 | **玩客云**、树莓派3及更早、旧ARM设备 |

---

## 🚀 快速部署

### 1. 玩客云（ARMv7 32位）
```bash
# 上传 pic-bed-linux-armv7 到 /opt/Pic/
cd /opt/Pic
mv pic-bed-linux-armv7 pic-bed
chmod +x pic-bed

# 首次运行生成配置
./pic-bed
```

### 2. systemd 开机自启
```bash
# 复制服务文件
cp pic-bed.service /etc/systemd/system/

# 重载并启动
systemctl daemon-reload
systemctl enable --now pic-bed

# 查看状态
systemctl status pic-bed

# 查看日志
journalctl -u pic-bed -f
```

---

## 🔧 配置文件 config.json

首次运行自动生成：
```json
{
  "port": 8080,
  "storage_dir": "./data",
  "max_size": 10,
  "api_key": ""
}
```

修改后重启生效：`systemctl restart pic-bed`

---

## 📡 接口使用

**上传：**
```bash
curl -F "file=@test.jpg" http://服务器IP:8080/upload
```

**预览：**
```
http://服务器IP:8080/img/2026/06/文件名.jpg
```
