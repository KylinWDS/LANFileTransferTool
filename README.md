# LAN File Transfer Tool - 局域网文件传输工具

<div align="center">

**轻量级、跨平台、功能完善的局域网文件传输桌面应用**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Wails](https://img.shields.io/badge/Wails-v2-2196F3?style=flat)](https://wails.io)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey)](https://github.com/yourusername/lanfiletransfertool)

[功能特性](#-功能特性) • [快速开始](#-快速开始) • [使用指南](#-使用指南) • [技术架构](#-技术架构)

</div>

---

## 📖 项目简介

LAN File Transfer Tool 是一个功能完善的局域网文件传输**桌面应用**，使用 Go + Wails + Vue 3 开发，提供原生文件选择体验、多协议传输、断点续传、文件加密等企业级功能。

### 🎯 适用场景

- 🏢 **办公环境**: 团队内部快速共享文件，支持批量传输
- 🏠 **家庭网络**: 家庭成员间文件传输，支持大文件断点续传
- 🎓 **教学场景**: 教师与学生之间资料分发，支持打包下载
- 💼 **临时共享**: 无需搭建服务器的快速文件分享，支持加密传输

---

## ✨ 功能特性

### 🎨 用户体验

- 🌓 **主题系统** - 支持白天/黑夜两种主题，护眼设计
- 🌍 **多语言支持** - 支持中文、英文、俄语三种语言
- 📱 **响应式界面** - 现代化UI设计，流畅操作体验
- 🎯 **一键操作** - 简化操作流程，提升使用效率

### 🚀 核心功能

#### 文件传输
- 📁 **文件选择** - 原生文件选择器，支持文件和文件夹
- 🔗 **链接生成** - 自动生成加密下载链接，支持二维码
- 📥 **客户端下载** - 支持客户端内打开下载链接，无需浏览器
- 📦 **批量下载** - 支持批量选择文件，打包下载
- 🔄 **断点续传** - 支持大文件分片传输，中断后可恢复
- ⚡ **多线程传输** - 多线程并发传输，自适应线程池

#### 安全特性
- 🔒 **文件加密** - AES-256-GCM加密传输
- 🔐 **完整性校验** - SHA256文件完整性验证
- 🛡️ **访问控制** - IP黑白名单，Token验证
- 🔑 **密钥管理** - 自动生成加密密钥

#### 性能优化
- 📊 **性能监控** - 实时监控CPU、内存、网络速度
- 🎯 **自适应优化** - 根据系统资源动态调整
- 💾 **传输统计** - 实时统计传输速度和进度

#### 环境检测
- 🔍 **防火墙检测** - 自动检测防火墙状态
- 🌐 **网络检测** - 检测网络连接和局域网访问
- 🔌 **端口检测** - 检测端口占用情况
- 💡 **解决方案** - 提供智能解决方案建议

#### 其他功能
- 📜 **历史记录** - 保存传输历史，支持一键重发
- ⚙️ **用户配置** - 个性化配置，支持主题、语言设置
- 📈 **传输协议** - 支持TCP、UDP、P2P多种传输方式

---

## 🚀 快速开始

### 方式一：下载预编译版本（推荐）

从 [Releases](https://github.com/yourusername/lanfiletransfertool/releases) 下载对应平台的安装包：

- **Windows**: `lanfiletransfertool-windows-amd64.exe`
- **macOS**: `lanfiletransfertool-darwin-amd64.app` 或 `lanfiletransfertool-darwin-arm64.app`
- **Linux**: `lanfiletransfertool-linux-amd64`

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
git clone https://github.com/yourusername/lanfiletransfertool.git
cd lanfiletransfertool

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

---

## 📚 使用指南

### 基本使用流程

#### 1. 启动应用
- 双击打开应用
- 应用自动启动HTTP服务器
- 界面显示服务器地址和状态

#### 2. 主题和语言设置
- 点击右上角主题按钮切换白天/黑夜模式
- 点击语言选择器切换中文/英文/俄语

#### 3. 文件传输
**发送文件：**
1. 点击"选择文件"或"选择文件夹"按钮
2. 查看文件信息（名称、大小、类型）
3. 点击"生成下载链接"
4. 复制链接或扫描二维码分享给接收方

**接收文件：**
1. 在浏览器中打开下载链接
2. 点击下载按钮
3. 或在客户端内粘贴链接直接下载

#### 4. 批量下载
1. 查看可用文件列表
2. 勾选需要下载的文件
3. 点击"批量下载"或"打包下载"
4. 查看下载进度

#### 5. 文件校验
1. 下载完成后点击"校验文件完整性"
2. 系统自动计算SHA256校验和
3. 显示校验结果，确保文件完整

### 高级功能

#### 性能监控
- 点击"启动线程池"开启多线程传输
- 实时查看CPU、内存使用情况
- 监控传输速度和进度

#### 环境检测
- 点击"开始检测"检查系统环境
- 查看防火墙、网络、端口状态
- 根据建议解决潜在问题

#### 加密传输
- 点击"生成密钥"创建加密密钥
- 输入文本进行加密/解密
- 确保传输数据安全

### 配置说明

配置文件位于 `config.yaml`，首次运行自动创建。

```yaml
# 应用配置
app:
  name: "LAN File Transfer Tool"
  version: "1.0.0"

# 服务器配置
server:
  port: 8080              # 服务端口
  host: "0.0.0.0"        # 监听地址

# 传输配置
transfer:
  max_connections: 10     # 最大连接数
  chunk_size: 1048576     # 分片大小(1MB)
  enable_resume: true     # 启用断点续传

# 安全配置
security:
  token_expiry: 86400    # Token有效期(秒)
  secret_key: "..."      # 加密密钥
  whitelist: []          # IP白名单
  blacklist: []          # IP黑名单

# 历史记录
history:
  max_records: 10        # 最大记录数
```

---

## 🏗️ 技术架构

### 后端技术栈 (Go)

| 模块 | 技术 | 说明 |
|------|------|------|
| **框架** | Gin | HTTP服务器框架 |
| **运行时** | Wails v2 | 桌面应用框架 |
| **数据库** | SQLite | 历史记录存储 |
| **加密** | AES-256-GCM | 文件加密 |
| **校验** | SHA256 | 文件完整性校验 |
| **二维码** | go-qrcode | 二维码生成 |
| **国际化** | i18n | 多语言支持 |

### 前端技术栈 (Vue 3)

| 模块 | 技术 | 说明 |
|------|------|------|
| **框架** | Vue 3 | 渐进式JavaScript框架 |
| **构建** | Vite | 下一代前端构建工具 |
| **状态** | Composition API | 组合式API |
| **主题** | CSS Variables | 动态主题切换 |
| **国际化** | i18n | 多语言支持 |

### 项目结构

```
LANFileTransferTool/
├── cmd/desktop/              # 桌面应用入口
├── internal/                 # Go核心代码
│   ├── config/              # 配置管理
│   ├── server/              # HTTP服务器
│   ├── storage/             # 数据存储
│   ├── token/               # Token管理
│   ├── access/              # 访问控制
│   ├── transfer/            # 文件传输
│   ├── environment/         # 环境检测
│   ├── userconfig/          # 用户配置
│   ├── resume/              # 断点续传
│   ├── performance/         # 性能优化
│   ├── encryption/          # 加密传输
│   ├── checksum/            # 文件校验
│   └── i18n/                # 国际化
├── pkg/                      # 公共包
│   ├── utils/               # 工具函数
│   ├── constants/           # 常量定义
│   ├── logger/              # 日志系统
│   └── errors/              # 错误处理
├── frontend/                 # Vue前端
│   ├── src/
│   │   ├── components/      # Vue组件
│   │   ├── api/            # API服务
│   │   ├── i18n/           # 国际化
│   │   ├── composables/    # 组合式函数
│   │   └── utils/          # 工具函数
│   ├── index.html
│   └── package.json
├── config.yaml              # 配置文件
└── README.md                # 本文档
```

---

## 📊 API接口文档

### 文件传输接口

```
POST /api/files/select          # 选择文件
POST /api/files/generate        # 生成下载链接
GET  /download/:token           # 下载文件
GET  /api/download/info/:token  # 获取下载信息
GET  /api/files/available       # 获取可用文件列表
POST /api/batch/download        # 批量下载
```

### 校验接口

```
POST /api/checksum/calculate    # 计算SHA256
POST /api/checksum/verify       # 验证文件完整性
```

### 配置接口

```
GET  /api/user/config           # 获取用户配置
POST /api/user/config           # 保存用户配置
POST /api/user/config/reset     # 重置配置
```

### 性能接口

```
GET  /api/performance/stats     # 获取性能统计
POST /api/performance/pool/init # 初始化线程池
POST /api/performance/pool/stop # 停止线程池
```

### 加密接口

```
POST /api/encryption/encrypt    # 加密数据
POST /api/encryption/decrypt    # 解密数据
POST /api/encryption/key        # 生成密钥
```

### 环境检测接口

```
GET  /api/environment/check     # 环境检测
```

---

## 🔒 安全建议

1. **修改密钥**: 生产环境务必修改 `secret_key`
2. **设置白名单**: 限制访问IP范围
3. **启用加密**: 传输敏感文件时启用加密
4. **校验文件**: 下载后验证文件完整性
5. **定期清理**: 定期清理历史记录
6. **防火墙**: 配置防火墙规则

---

## 🎯 性能优化建议

1. **多线程传输**: 启动线程池提升传输速度
2. **分片大小**: 根据网络情况调整分片大小
3. **断点续传**: 大文件传输启用断点续传
4. **网络优化**: 使用有线网络提升稳定性
5. **资源监控**: 监控CPU和内存使用情况

---

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

### 开发环境

```bash
# 克隆仓库
git clone https://github.com/yourusername/lanfiletransfertool.git
cd lanfiletransfertool

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

---

## 📝 更新日志

### v2.0.0 (2026-03-25)

#### 新增功能
- ✨ 主题系统 - 支持白天/黑夜两种主题
- ✨ i18n国际化 - 支持中文、英文、俄语
- ✨ 客户端内下载 - 无需浏览器即可下载
- ✨ 批量下载 - 支持批量选择和打包下载
- ✨ 文件校验 - SHA256完整性校验
- ✨ 性能监控 - 实时监控传输性能
- ✨ 环境检测 - 自动检测系统环境
- ✨ 断点续传 - 大文件传输支持断点恢复
- ✨ 文件加密 - AES-256-GCM加密传输
- ✨ 多线程传输 - 自适应线程池优化

#### 优化改进
- 🎨 重构前端代码，模块化组件设计
- 🎨 优化UI界面，提升用户体验
- 🎨 完善国际化支持，前后端统一
- 🎨 添加完整中文注释，提升代码可读性
- 🎨 清理冗余文档，优化项目结构

### v1.0.0 (2024-01-01)

- ✨ 初始版本发布
- ✅ 桌面应用，原生体验
- ✅ 文件和文件夹传输
- ✅ 加密下载链接
- ✅ 历史记录管理
- ✅ IP黑白名单
- ✅ 二维码分享

---

## 📄 许可证

[MIT License](LICENSE)

---

## 💬 联系方式

如有问题或建议，请提交 [Issue](https://github.com/yourusername/lanfiletransfertool/issues)

---

<div align="center">

**⭐ 如果这个项目对你有帮助，请给一个 Star！⭐**

Made with ❤️ by LANFileTransferTool Team

</div>
