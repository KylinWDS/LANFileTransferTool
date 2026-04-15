# LAN-File-Transfer-Tool (LANftt)

<div align="center">

**轻量级、跨平台、功能完善的局域网文件传输桌面应用**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat\&logo=go)](https://golang.org)
[![Wails](https://img.shields.io/badge/Wails-v2-2196F3?style=flat)](https://wails.io)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey)](https://github.com/KylinWDS/LANFileTransferTool)

[功能特性](#-功能特性) • [快速开始](#-快速开始) • [使用指南](#-使用指南) • [技术架构](#-技术架构)

</div>

***

## 📖 项目简介

**LAN-File-Transfer-Tool**（简称 **LANftt**）是一个功能完善的局域网文件传输**桌面应用**，使用 Go + Wails v2 + Vue 3 开发，提供原生文件选择体验、多协议传输、断点续传、文件加密等企业级功能。

### 🎯 适用场景

| 场景          | 说明                    |
| ----------- | --------------------- |
| 🏢 **办公环境** | 团队内部快速共享文件，支持批量传输     |
| 🏠 **家庭网络** | 家庭成员间文件传输，支持大文件断点续传   |
| 🎓 **教学场景** | 教师与学生之间资料分发，支持打包下载    |
| 💼 **临时共享** | 无需搭建服务器的快速文件分享，支持加密传输 |

***

## ✨ 功能特性

### 🎨 用户体验

| 特性           | 说明               |
| ------------ | ---------------- |
| 🌓 **主题系统**  | 支持浅色/深色两种主题，护眼设计 |
| 🌍 **多语言支持** | 支持中文、英文、俄语三种语言   |
| 📱 **响应式界面** | 现代化UI设计，流畅操作体验   |
| 🎯 **一键操作**  | 简化操作流程，提升使用效率    |
| 📝 **日志系统**  | 完整的操作日志记录，支持导出   |

### 🚀 核心功能

#### 文件传输

| 功能           | 说明                 |
| ------------ | ------------------ |
| 📁 **文件选择**  | 原生文件选择器，支持文件和文件夹   |
| 🔗 **链接生成**  | 自动生成加密下载链接，支持二维码   |
| 📥 **客户端下载** | 支持客户端内打开下载链接，无需浏览器 |
| 📦 **批量下载**  | 支持批量选择文件，打包下载      |
| 🔄 **断点续传**  | 支持大文件分片传输，中断后可恢复   |
| ⚡ **多线程传输**  | 多线程并发传输，自适应线程池     |

#### 多协议支持

| 协议               | 特点             | 适用场景        |
| ---------------- | -------------- | ----------- |
| 🌐 **HTTP**      | 标准HTTP协议，兼容性最好 | 小文件、兼容性优先   |
| 🔌 **WebSocket** | 实时双向通信，适合大文件   | 中等文件、实时性要求高 |
| 📡 **UDP**       | 高速传输，适合局域网     | 大文件、速度优先    |
| 🔗 **P2P**       | 点对点直连，无需服务器中转  | 超大文件、局域网环境  |

#### 网络功能

| 功能            | 说明              |
| ------------- | --------------- |
| 📋 **IP选择器**  | 自动检测并显示所有可用IP地址 |
| ✏️ **手动IP输入** | 支持跨网段连接         |
| 🔍 **设备发现**   | UDP广播自动发现局域网设备  |
| 📊 **连接状态**   | 实时显示设备连接状态      |

#### 安全特性

| 特性             | 说明                    |
| -------------- | --------------------- |
| 🔒 **文件加密**    | AES-256-GCM加密传输       |
| 🔐 **完整性校验**   | SHA256文件完整性验证         |
| 🛡️ **访问控制**   | IP黑白名单，Token验证        |
| 🔑 **密钥管理**    | 自动生成加密密钥，支持自定义密钥解析    |
| 🎫 **加密Token** | Token自包含文件元数据，支持跨重启解析 |

#### 性能优化

| 功能           | 说明              |
| ------------ | --------------- |
| 📊 **性能监控**  | 实时监控CPU、内存、网络速度 |
| 💾 **磁盘监控**  | 实时监控磁盘读写速度      |
| 🎯 **自适应优化** | 根据系统资源动态调整      |
| 📈 **传输统计**  | 实时统计传输速度和进度     |

#### 环境检测

| 检测项          | 说明           |
| ------------ | ------------ |
| 🔥 **防火墙检测** | 自动检测防火墙状态    |
| 🌐 **网络检测**  | 检测网络连接和局域网访问 |
| 🔌 **端口检测**  | 检测端口占用情况     |
| 💡 **解决方案**  | 提供智能解决方案建议   |

#### 其他功能

| 功能          | 说明                   |
| ----------- | -------------------- |
| 📜 **历史记录** | 保存传输历史，支持一键重发、链接重新生成 |
| ⚙️ **用户配置** | 个性化配置，支持主题、语言设置      |
| 📝 **日志查看** | 完整的操作日志，支持筛选和导出      |

***

## 🚀 快速开始

### 方式一：下载预编译版本（推荐）

从 [Releases](https://github.com/KylinWDS/LANFileTransferTool/releases) 下载对应平台的安装包：

| 平台                        | 文件                         |
| ------------------------- | -------------------------- |
| **Windows**               | `LANftt-windows-amd64.exe` |
| **macOS (Intel)**         | `LANftt-darwin-amd64.app`  |
| **macOS (Apple Silicon)** | `LANftt-darwin-arm64.app`  |
| **Linux**                 | `LANftt-linux-amd64`       |

### 方式二：从源码构建

#### 1. 安装依赖

```bash
# 安装 Go (1.21+)
# macOS
brew install go

# Linux
sudo apt install golang-go

# 安装 Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 安装前端依赖
cd frontend
npm install
```

#### 2. 运行开发版

```bash
# 克隆仓库
git clone https://github.com/KylinWDS/LANFileTransferTool.git
cd LANFileTransferTool

# 开发模式运行
wails dev
```

#### 3. 构建生产版

```bash
# 构建当前平台版本
wails build

# 构建所有平台版本
wails build -platform darwin/amd64
wails build -platform darwin/arm64
wails build -platform windows/amd64
wails build -platform linux/amd64
```

***

## 📚 使用指南

### 基本使用流程

#### 1. 启动应用

- 双击打开应用
- 应用自动启动HTTP服务器
- 界面显示服务器地址和状态

#### 2. 主题和语言设置

- 点击右上角主题按钮切换浅色/深色模式
- 点击语言选择器切换中文/英文/俄语

#### 3. 文件传输

**发送文件：**

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  选择文件    │───▶│  选择IP     │───▶│  生成链接    │
│  或文件夹   │    │  地址       │    │  + 二维码   │
└─────────────┘    └─────────────┘    └─────────────┘
```

**接收文件：**

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  打开链接    │───▶│  查看文件    │───▶│  下载文件    │
│  或扫码     │    │  信息       │    │  到本地     │
└─────────────┘    └─────────────┘    └─────────────┘
```

#### 4. 设备发现

1. 查看"发现的设备"列表
2. 点击设备可快速连接
3. 支持跨网段手动输入IP

#### 5. 批量下载

1. 查看可用文件列表
2. 勾选需要下载的文件
3. 点击"批量下载"
4. 查看下载进度

#### 6. 文件校验

1. 下载完成后点击"校验文件完整性"
2. 系统自动计算SHA256校验和
3. 显示校验结果，确保文件完整

### 高级功能

#### 性能监控

- 点击"启动线程池"开启多线程传输
- 实时查看CPU、内存使用情况
- 监控网络发送/接收速度
- 监控磁盘读写速度

#### 环境检测

- 点击"开始检测"检查系统环境
- 查看防火墙、网络、端口状态
- 根据建议解决潜在问题

#### 加密传输

- 点击"生成密钥"创建加密密钥
- 输入文本进行加密/解密
- 在下载管理中使用自定义密钥解析加密链接

#### 日志管理

- 查看所有操作日志
- 按级别筛选日志
- 导出日志到文件
- 清除日志记录

### 配置说明

配置文件位于用户数据目录：

| 平台          | 配置目录                                    |
| ----------- | --------------------------------------- |
| **macOS**   | `~/Library/Application Support/LANftt/` |
| **Windows** | `%LOCALAPPDATA%/LANftt/`                |
| **Linux**   | `~/.config/lanftt/`                     |

**config.yaml** 配置项：

```yaml
app:
  name: "LAN-File-Transfer-Tool"
  short_name: "LANftt"
  version: "0.2.0"

server:
  port: 8080              # 服务端口
  host: "0.0.0.0"        # 监听地址

discovery:
  enabled: true           # 启用设备发现
  port: 37021             # 发现端口
  broadcast_interval: 5   # 广播间隔(秒)
  peer_timeout: 30        # 超时时间(秒)

websocket:
  enabled: true
  chunk_size: 65536

udp:
  enabled: true
  port: 37022
  chunk_size: 32768

p2p:
  enabled: true
  port: 37023
  chunk_size: 65536

transfer:
  max_connections: 10     # 最大连接数
  chunk_size: 1048576     # 分片大小(1MB)
  enable_resume: true     # 启用断点续传

security:
  token_expiry: 86400    # Token有效期(秒)
  secret_key: "..."      # 加密密钥
  whitelist: []          # IP白名单
  blacklist: []          # IP黑名单

history:
  max_records: 100       # 最大记录数

performance:
  pool_size: 10          # 线程池大小
  monitor_interval: 2    # 监控间隔(秒)
```

***

## 🏗️ 技术架构

### 后端技术栈 (Go)

| 模块            | 技术                | 说明        |
| ------------- | ----------------- | --------- |
| **框架**        | Wails v2          | 桌面应用框架    |
| **HTTP**      | Gin               | HTTP服务器框架 |
| **数据库**       | SQLite            | 历史记录存储    |
| **加密**        | AES-256-GCM       | 文件加密      |
| **校验**        | SHA256            | 文件完整性校验   |
| **WebSocket** | gorilla/websocket | 实时通信      |
| **UDP**       | net/udp           | UDP传输协议   |
| **P2P**       | net/tcp           | P2P传输协议   |

### 前端技术栈 (Vue 3)

| 模块      | 技术              | 说明              |
| ------- | --------------- | --------------- |
| **框架**  | Vue 3           | 渐进式JavaScript框架 |
| **构建**  | Vite            | 下一代前端构建工具       |
| **状态**  | Composition API | 组合式API          |
| **主题**  | CSS Variables   | 动态主题切换          |
| **国际化** | vue-i18n        | 多语言支持           |
| **二维码** | qrcode          | 二维码生成           |

### 项目结构

```
LANFileTransferTool/
├── main.go                    # 应用入口
├── config.yaml                # 配置文件
├── wails.json                 # Wails配置
├── internal/                  # Go核心代码
│   ├── app/                   # 应用主逻辑
│   ├── config/                # 配置管理
│   ├── server/                # HTTP服务器
│   ├── storage/               # 数据存储
│   ├── token/                 # Token管理
│   ├── access/                # 访问控制
│   ├── transfer/              # 文件传输
│   ├── environment/           # 环境检测
│   ├── userconfig/            # 用户配置
│   ├── resume/                # 断点续传
│   ├── performance/           # 性能优化
│   ├── encryption/            # 加密传输
│   ├── checksum/              # 文件校验
│   ├── stats/                 # 传输统计
│   ├── discovery/             # 设备发现
│   ├── websocket/             # WebSocket协议
│   ├── udp/                   # UDP协议
│   ├── p2p/                   # P2P协议
│   └── protocol/              # 协议选择器
├── pkg/                       # 公共包
│   ├── utils/                 # 工具函数
│   ├── constants/             # 常量定义
│   ├── logger/                # 日志系统
│   └── errors/                # 错误处理
├── frontend/                  # Vue前端
│   ├── src/
│   │   ├── components/        # Vue组件
│   │   ├── api/               # API服务
│   │   ├── i18n/              # 国际化
│   │   └── utils/             # 工具函数
│   ├── index.html
│   └── package.json
└── README.md                  # 本文档
```

详细技术文档请参阅 [IMPLEMENTATION.md](./IMPLEMENTATION.md)

***

## 📊 API接口文档

### Wails绑定函数

#### 文件传输

| 函数                                     | 说明       |
| -------------------------------------- | -------- |
| `SelectFiles(directory)`               | 选择文件或文件夹 |
| `GenerateDownloadLink(filePath)`       | 生成下载链接   |
| `GenerateBatchDownloadLink(filePaths)` | 批量生成链接   |
| `GetAvailableFiles()`                  | 获取可用文件列表 |
| `GetDownloadInfo(token)`               | 获取下载信息   |
| `DownloadFile(token, savePath)`        | 下载文件     |
| `BatchDownload(fileIDs, savePath)`     | 批量下载     |

#### 网络功能

| 函数                  | 说明       |
| ------------------- | -------- |
| `GetServerInfo()`   | 获取服务器信息  |
| `GetAllIPs()`       | 获取所有IP地址 |
| `SetSelectedIP(ip)` | 设置选中IP   |
| `DiscoverPeers()`   | 发现设备     |

#### 加密功能

| 函数                             | 说明   |
| ------------------------------ | ---- |
| `EncryptData(plainText, key)`  | 加密数据 |
| `DecryptData(cipherText, key)` | 解密数据 |
| `GenerateEncryptionKey()`      | 生成密钥 |

#### 历史记录

| 函数                          | 说明     |
| --------------------------- | ------ |
| `GetHistory(limit)`         | 获取历史记录 |
| `ClearHistory()`            | 清除历史记录 |
| `DeleteHistory(id)`         | 删除单条记录 |
| `RegenerateLink(historyID)` | 重新生成链接 |

### HTTP API

```
POST /api/files/select          # 选择文件
POST /api/files/generate        # 生成下载链接
GET  /api/files/available       # 获取可用文件列表
GET  /api/download/info/:token  # 获取下载信息
POST /api/batch/download        # 批量下载
POST /api/checksum/calculate    # 计算SHA256
POST /api/checksum/verify       # 验证文件完整性
GET  /api/user/config           # 获取用户配置
POST /api/user/config           # 保存用户配置
GET  /api/performance/stats     # 获取性能统计
POST /api/encryption/encrypt    # 加密数据
POST /api/encryption/decrypt    # 解密数据
GET  /api/environment/check     # 环境检测
GET  /api/history               # 获取历史记录
GET  /download/:token           # 下载文件
```

***

## 🔒 安全建议

| 建议        | 说明                    |
| --------- | --------------------- |
| **修改密钥**  | 生产环境务必修改 `secret_key` |
| **设置白名单** | 限制访问IP范围              |
| **启用加密**  | 传输敏感文件时启用加密           |
| **校验文件**  | 下载后验证文件完整性            |
| **定期清理**  | 定期清理历史记录              |
| **防火墙**   | 配置防火墙规则               |

***

## 🎯 性能优化建议

| 建议        | 说明           |
| --------- | ------------ |
| **多线程传输** | 启动线程池提升传输速度  |
| **分片大小**  | 根据网络情况调整分片大小 |
| **断点续传**  | 大文件传输启用断点续传  |
| **网络优化**  | 使用有线网络提升稳定性  |
| **资源监控**  | 监控CPU和内存使用情况 |

***

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

### 开发环境

```bash
# 克隆仓库
git clone https://github.com/KylinWDS/LANFileTransferTool.git
cd LANFileTransferTool

# 安装依赖
go mod download
cd frontend && npm install

# 运行开发版
wails dev

# 构建生产版
wails build
```

### 代码规范

- Go代码遵循官方规范，添加完整中文注释
- Vue代码遵循Vue风格指南，添加完整中文注释
- 提交前运行 `go fmt` 和 `npm run lint`

***

## 📝 更新日志

### v0.2.0 (2026-06)

#### 新增功能

- ✨ 多协议支持 - HTTP/WebSocket/UDP/P2P
- ✨ IP选择器 - 自动检测所有可用IP
- ✨ 手动IP输入 - 支持跨网段连接
- ✨ 设备发现 - UDP广播自动发现局域网设备
- ✨ 网速监控 - 实时监控发送/接收速度
- ✨ 磁盘监控 - 实时监控磁盘读写速度
- ✨ 加密Token - AES加密自包含Token
- ✨ 自定义密钥 - 支持自定义密钥解析链接
- ✨ 日志系统 - 完整的操作日志记录

#### 优化改进

- 🎨 统一项目名称为 LAN-File-Transfer-Tool (LANftt)
- 🎨 完善国际化支持，三种语言全覆盖
- 🎨 优化配置文件结构，支持更多配置项
- 🎨 添加完整中文注释，提升代码可读性
- 🎨 更新文档，完善项目说明
- 🎨 优化UI主题系统，支持浅色/深色模式
- 🎨 改进数据存储位置，使用用户可写目录

### v0.1.0 (2026-04)

- ✨ 初始版本发布
- ✅ 桌面应用，原生体验
- ✅ 文件和文件夹传输
- ✅ 加密下载链接
- ✅ 历史记录管理
- ✅ IP黑白名单
- ✅ 二维码分享

***

## 📄 许可证

[MIT License](LICENSE)

***

## 💬 联系方式

如有问题或建议，请提交 [Issue](https://github.com/KylinWDS/LANFileTransferTool/issues)

***

<div align="center">

**⭐ 如果这个项目对你有帮助，请给一个 Star！⭐**

Made with ❤️ by LANftt Team

</div>
