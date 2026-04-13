# LAN-File-Transfer-Tool (LANftt) - 项目实现说明

## 项目信息

- **全称**: LAN-File-Transfer-Tool
- **简称**: LANftt
- **版本**: v0.2.0
- **描述**: 轻量级、跨平台、功能完善的局域网文件传输桌面应用

## 📁 项目结构

```
LANFileTransferTool/
├── main.go                          # 应用入口
├── go.mod                           # Go模块配置
├── config.yaml                      # 应用配置文件
├── wails.json                       # Wails配置文件
├── internal/                        # Go后端核心代码
│   ├── app/app.go                   # 应用主逻辑，Wails绑定
│   ├── config/config.go             # 配置管理
│   ├── config/manager.go            # 配置管理器（加载顺序管理）
│   ├── server/server.go             # HTTP服务器
│   ├── server/handlers.go           # HTTP请求处理器
│   ├── storage/storage.go           # SQLite数据存储
│   ├── token/token.go               # Token管理（AES加密）
│   ├── access/access.go             # 访问控制（IP黑白名单）
│   ├── transfer/transfer.go         # 文件传输核心
│   ├── resume/resume.go             # 断点续传
│   ├── encryption/encryption.go     # AES-256-GCM加密
│   ├── checksum/checksum.go         # SHA256校验
│   ├── performance/performance.go   # 性能监控
│   ├── performance/pool.go          # 线程池实现
│   ├── environment/environment.go   # 环境检测
│   ├── userconfig/userconfig.go     # 用户配置管理
│   ├── stats/monitor.go             # 传输统计监控
│   ├── discovery/discovery.go       # UDP设备发现服务
│   ├── websocket/websocket.go       # WebSocket传输协议
│   ├── udp/udp.go                   # UDP传输协议
│   ├── p2p/p2p.go                   # P2P传输协议
│   └── protocol/selector.go         # 智能协议选择器
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
    │   ├── api/index.js             # API服务
    │   ├── i18n/                    # 国际化文件
    │   │   ├── zh-CN.json           # 中文
    │   │   ├── en.json              # 英文
    │   │   └── ru.json              # 俄语
    │   └── components/              # Vue组件
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
   - 文件选择和注册（支持文件和文件夹）
   - 下载链接生成（支持二维码）
   - 批量文件下载
   - 实时进度显示（百分比、速度、剩余时间）
   - 多协议支持（HTTP/WebSocket/UDP/P2P）
   - 智能协议选择（根据文件大小自动选择最佳协议）

2. **网络功能**
   - IP选择器（显示所有可用IP）
   - 手动IP输入（跨网段连接）
   - UDP广播设备发现
   - 自动发现局域网设备

3. **安全特性**
   - AES-256-GCM加密/解密
   - SHA256文件完整性校验
   - 加密Token认证机制
   - IP访问控制（黑白名单）
   - 自定义密钥解析

4. **性能监控**
   - 实时网速监控（发送/接收）
   - 磁盘读写速度监控
   - CPU/内存使用率监控
   - 线程池管理（启动/停止/配置）

5. **环境检测**
   - 防火墙状态检测
   - 网络连接检查
   - 端口占用检测
   - 智能解决方案建议

6. **用户配置**
   - 主题切换（浅色/深色）
   - 多语言支持（中文/英文/俄语）
   - 个性化设置持久化
   - 配置加载顺序（初始化→系统配置→用户配置覆盖）

7. **历史记录**
   - 传输记录存储（含文件路径、协议、时长、链接）
   - 操作日志查看
   - 一键清除功能
   - 单条删除功能
   - 链接重新生成功能
   - 链接复制功能

## 🔧 配置系统

### 配置加载顺序

1. **初始化默认配置**
2. **加载系统配置**（config.yaml）
3. **加载用户配置**（user_config.json）覆盖系统配置
4. **应用临时配置**（运行时动态调整）

### 配置文件

```yaml
# 应用配置
app:
  name: "LAN-File-Transfer-Tool"
  short_name: "LANftt"
  version: "0.2.0"

# HTTP服务器配置
server:
  port: 8080              # 服务端口
  host: "0.0.0.0"        # 监听地址

# UDP发现服务配置
discovery:
  enabled: true           # 是否启用
  port: 37021             # 发现服务端口
  broadcast_interval: 5   # 广播间隔(秒)
  peer_timeout: 30        # 节点超时(秒)

# WebSocket传输配置
websocket:
  enabled: true           # 是否启用
  port_offset: 0          # 端口偏移
  chunk_size: 65536       # 数据块大小

# UDP传输配置
udp:
  enabled: true           # 是否启用
  port: 37022             # UDP端口
  chunk_size: 32768       # 数据块大小

# P2P传输配置
p2p:
  enabled: true           # 是否启用
  port: 37023             # P2P端口
  chunk_size: 65536       # 数据块大小

# 文件传输配置
transfer:
  max_connections: 10     # 最大连接数
  chunk_size: 1048576     # 分片大小(1MB)
  enable_resume: true     # 启用断点续传
  default_protocol: "auto" # 默认协议（auto/http/websocket/udp/p2p）

# 安全配置
security:
  token_expiry: 86400     # Token有效期(秒)
  secret_key: "..."       # 加密密钥
  whitelist: []           # IP白名单
  blacklist: []           # IP黑名单

# 历史记录配置
history:
  max_records: 100        # 最大记录数

# 性能配置
performance:
  pool_size: 10           # 线程池大小
  monitor_interval: 2     # 监控间隔(秒)
```

## 📡 API接口

### 文件传输
- `POST /api/files/select` - 选择文件
- `POST /api/files/generate` - 生成下载链接
- `GET /download/:token` - 下载文件
- `GET /api/download/info/:token` - 获取下载信息
- `GET /api/files/available` - 获取可用文件列表
- `POST /api/batch/download` - 批量下载

### 协议选择
- `GetProtocolRecommendation(fileSize)` - 获取协议推荐
- `SetProtocolPreference(pref)` - 设置协议偏好
- `SelectProtocol(fileSize, peerAvailable, userOverride)` - 选择传输协议

### 校验接口
- `POST /api/checksum/calculate` - 计算SHA256
- `POST /api/checksum/verify` - 验证文件完整性

### 加密接口
- `POST /api/encryption/encrypt` - 加密数据
- `POST /api/encryption/decrypt` - 解密数据
- `POST /api/encryption/key` - 生成密钥

### 性能接口
- `GetPerformanceStats()` - 获取性能统计
- `InitThreadPool(size)` - 初始化线程池
- `StopThreadPool()` - 停止线程池

### 环境检测
- `CheckEnvironment()` - 环境检测

### 用户配置
- `GetUserConfig()` - 获取用户配置
- `SaveUserConfig(config)` - 保存用户配置
- `ResetUserConfig()` - 重置配置

### 历史记录
- `GetHistory(limit)` - 获取历史记录
- `ClearHistory()` - 清除历史记录
- `DeleteHistory(id)` - 删除单条记录
- `RegenerateLink(historyID)` - 重新生成下载链接

## 🎨 技术栈

**后端 (Go)**:
- Wails v2 (桌面应用框架)
- Gin (HTTP框架)
- SQLite (数据存储)
- AES-256-GCM (加密)
- SHA256 (校验)
- gorilla/websocket (WebSocket)

**前端 (Vue 3)**:
- Vue 3 Composition API
- Vite (构建工具)
- vue-i18n (国际化)
- CSS Variables (主题系统)
- qrcode (二维码生成)

## 🚀 快速开始

### 前置要求

- Go 1.21+
- Node.js 16+
- Wails CLI

### 安装步骤

```bash
# 安装Go依赖
go mod tidy

# 安装前端依赖
cd frontend && npm install && cd ..

# 运行开发模式
wails dev

# 构建生产版本
wails build
```

## 📝 使用说明

1. **启动应用后，自动启动HTTP服务器**
2. **选择要分享的文件或文件夹**
3. **系统自动选择最佳传输协议**
4. **点击"生成下载链接"获取下载地址和二维码**
5. **接收方通过链接或扫描二维码下载文件**
6. **可在"下载管理"中查看可用文件并进行批量下载**
7. **使用"加密工具"对敏感数据进行加密**
8. **通过"环境检测"排查网络问题**
9. **在"设置"中自定义主题、语言和其他选项**
10. **在历史记录中查看传输详情、重新生成链接**

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
- 智能协议选择算法
- 完善的进度显示系统

## 📄 许可证

MIT License
