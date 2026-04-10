# LAN File Transfer Tool - 项目实现说明

## 📁 项目结构

```
LANFileTransferTool/
├── main.go                          # 应用入口
├── go.mod                           # Go模块配置
├── config.yaml                      # 应用配置文件
├── internal/                        # Go后端核心代码
│   ├── app/app.go                   # 应用主逻辑
│   ├── config/config.go             # 配置管理
│   ├── server/server.go             # HTTP服务器
│   ├── server/handlers.go           # API处理程序
│   ├── storage/storage.go           # SQLite数据存储
│   ├── token/token.go               # Token管理
│   ├── access/access.go             # 访问控制（IP黑白名单）
│   ├── transfer/transfer.go         # 文件传输核心
│   ├── resume/resume.go             # 断点续传
│   ├── encryption/encryption.go     # AES-256-GCM加密
│   ├── checksum/checksum.go         # SHA256校验
│   ├── performance/performance.go   # 性能监控和线程池
│   ├── environment/environment.go   # 环境检测
│   └── userconfig/userconfig.go     # 用户配置管理
├── pkg/                             # 公共包
│   ├── constants/constants.go       # 常量定义
│   ├── errors/errors.go             # 错误定义
│   ├── logger/logger.go             # 日志系统
│   └── utils/utils.go               # 工具函数
└── frontend/                        # Vue 3前端
    ├── package.json                 # 前端依赖配置
    ├── vite.config.js               # Vite构建配置
    ├── index.html                   # HTML入口
    ├── src/
    │   ├── main.js                  # Vue应用入口
    │   ├── App.vue                  # 根组件
    │   ├── style.css                # 全局样式
    │   ├── i18n/                    # 国际化文件
    │   │   ├── zh-CN.json           # 中文
    │   │   ├── en.json              # 英文
    │   │   └── ru.json              # 俄语
    │   └── components/              # Vue组件
    │       ├── HeaderComponent.vue  # 头部组件
    │       ├── TransferTab.vue      # 文件传输标签页
    │       ├── DownloadTab.vue      # 下载管理标签页
    │       ├── HistoryTab.vue       # 历史记录标签页
    │       ├── EncryptionTab.vue    # 加密工具标签页
    │       ├── PerformanceTab.vue   # 性能监控标签页
    │       ├── EnvironmentTab.vue   # 环境检测标签页
    │       └── SettingsTab.vue      # 设置标签页
```

## ✨ 已实现功能

### 🎯 核心功能

1. **文件传输系统**
   - 文件选择和注册
   - 下载链接生成（支持二维码）
   - 批量文件下载
   - 实时进度显示

2. **安全特性**
   - AES-256-GCM加密/解密
   - SHA256文件完整性校验
   - Token认证机制
   - IP访问控制（黑白名单）

3. **断点续传**
   - 传输状态保存
   - 中断恢复支持
   - 过期清理机制

4. **性能优化**
   - 多线程传输池
   - 自适应分片大小
   - 实时性能监控（CPU、内存、网络）
   - Goroutine数量监控

5. **环境检测**
   - 防火墙状态检测（Windows/macOS/Linux）
   - 网络连接检查
   - 端口占用检测
   - 智能解决方案建议

6. **用户配置**
   - 主题切换（浅色/深色）
   - 多语言支持（中文/英文/俄语）
   - 个性化设置持久化
   - 配置导入/导出

7. **历史记录**
   - 传输记录存储
   - 操作日志查看
   - 一键清除功能

## 🚀 快速开始

### 前置要求

- Go 1.21+
- Node.js 16+
- Wails CLI

### 安装步骤

1. **克隆项目**
```bash
cd /Users/kylin/Downloads/lftt/LANFileTransferTool
```

2. **安装Go依赖**
```bash
go mod tidy
```

3. **安装前端依赖**
```bash
cd frontend
npm install
cd ..
```

4. **运行开发模式**
```bash
wails dev
```

5. **构建生产版本**
```bash
wails build
```

## 🔧 配置说明

配置文件 `config.yaml` 包含以下设置：

- **服务器配置**: 端口、监听地址
- **传输配置**: 分片大小、最大连接数、断点续传开关
- **安全配置**: Token有效期、密钥、IP白名单/黑名单
- **历史记录**: 最大记录数

## 📡 API接口

### 文件传输
- `POST /api/files/select` - 选择文件
- `POST /api/files/generate` - 生成下载链接
- `GET /download/:token` - 下载文件
- `GET /api/download/info/:token` - 获取下载信息
- `GET /api/files/available` - 获取可用文件列表
- `POST /api/batch/download` - 批量下载

### 校验接口
- `POST /api/checksum/calculate` - 计算SHA256
- `POST /api/checksum/verify` - 验证文件完整性

### 加密接口
- `POST /api/encryption/encrypt` - 加密数据
- `POST /api/encryption/decrypt` - 解密数据
- `POST /api/encryption/key` - 生成密钥

### 性能接口
- `GET /api/performance/stats` - 获取性能统计
- `POST /api/performance/pool/init` - 初始化线程池
- `POST /api/performance/pool/stop` - 停止线程池

### 环境检测
- `GET /api/environment/check` - 环境检测

### 用户配置
- `GET /api/user/config` - 获取用户配置
- `POST /api/user/config` - 保存用户配置
- `POST /api/user/config/reset` - 重置配置

### 历史记录
- `GET /api/history` - 获取历史记录
- `DELETE /api/history` - 清除历史记录

## 🎨 技术栈

**后端 (Go)**:
- Wails v2 (桌面应用框架)
- Gin (HTTP框架)
- SQLite (数据存储)
- AES-256-GCM (加密)
- SHA256 (校验)

**前端 (Vue 3)**:
- Vue 3 Composition API
- Vite (构建工具)
- vue-i18n (国际化)
- CSS Variables (主题系统)

## 📝 使用说明

1. **启动应用后，自动启动HTTP服务器**
2. **选择要分享的文件或文件夹**
3. **点击"生成下载链接"获取下载地址和二维码**
4. **接收方通过链接或扫描二维码下载文件**
5. **可在"下载管理"中查看可用文件并进行批量下载**
6. **使用"加密工具"对敏感数据进行加密**
7. **通过"环境检测"排查网络问题**
8. **在"设置"中自定义主题、语言和其他选项**

## 🔒 安全建议

1. 生产环境请修改 `config.yaml` 中的 `secret_key`
2. 配置IP白名单限制访问范围
3. 传输敏感文件时启用加密
4. 下载后验证文件完整性
5. 定期清理历史记录

## 🛠️ 开发说明

### 代码规范
- Go代码遵循官方规范
- Vue代码遵循Vue风格指南
- 所有公开函数添加中文注释

### 项目特点
- 模块化设计，易于扩展
- 完整的错误处理机制
- 详细的日志记录
- 跨平台支持（Windows/macOS/Linux）

## 📄 许可证

MIT License
