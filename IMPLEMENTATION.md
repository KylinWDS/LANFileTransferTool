# LAN-File-Transfer-Tool (LANftt) - 项目实现说明

## 项目信息

- **全称**: LAN-File-Transfer-Tool
- **简称**: LANftt
- **版本**: v0.2.0
- **描述**: 轻量级、跨平台、功能完善的局域网文件传输桌面应用

---

## 一、物理架构

### 1.1 目录结构

```
LANFileTransferTool/
├── main.go                          # 应用入口，Wails初始化
├── go.mod                           # Go模块配置
├── go.sum                           # Go依赖校验
├── config.yaml                      # 系统配置文件
├── wails.json                       # Wails配置文件
│
├── internal/                        # Go后端核心代码（内部包）
│   ├── app/
│   │   └── app.go                   # 应用主逻辑，Wails绑定，服务编排
│   │
│   ├── config/
│   │   ├── config.go                # 配置结构体定义
│   │   └── manager.go               # 配置管理器（加载顺序管理）
│   │
│   ├── server/
│   │   ├── server.go                # HTTP服务器（Gin框架）
│   │   └── handlers.go              # HTTP请求处理器
│   │
│   ├── storage/
│   │   └── storage.go               # SQLite数据存储
│   │
│   ├── token/
│   │   └── token.go                 # Token管理（AES加密）
│   │
│   ├── access/
│   │   └── access.go                # 访问控制（IP黑白名单）
│   │
│   ├── transfer/
│   │   └── transfer.go              # 文件传输核心
│   │
│   ├── resume/
│   │   └── resume.go                # 断点续传
│   │
│   ├── encryption/
│   │   └── encryption.go            # AES-256-GCM加密
│   │
│   ├── checksum/
│   │   └── checksum.go              # SHA256校验
│   │
│   ├── performance/
│   │   ├── performance.go           # 性能监控
│   │   └── pool.go                  # 线程池实现
│   │
│   ├── environment/
│   │   └── environment.go           # 环境检测
│   │
│   ├── userconfig/
│   │   └── userconfig.go            # 用户配置管理
│   │
│   ├── stats/
│   │   └── monitor.go               # 传输统计监控
│   │
│   ├── discovery/
│   │   └── discovery.go             # UDP设备发现服务
│   │
│   ├── websocket/
│   │   └── websocket.go             # WebSocket传输协议
│   │
│   ├── udp/
│   │   └── udp.go                   # UDP传输协议
│   │
│   ├── p2p/
│   │   └── p2p.go                   # P2P传输协议
│   │
│   └── protocol/
│       └── selector.go              # 智能协议选择器
│
├── pkg/                             # 公共包（可被外部引用）
│   ├── constants/
│   │   └── constants.go             # 常量定义
│   │
│   ├── errors/
│   │   └── errors.go                # 错误定义
│   │
│   ├── logger/
│   │   └── logger.go                # 日志系统（支持内存收集）
│   │
│   └── utils/
│       └── utils.go                 # 工具函数
│
└── frontend/                        # Vue 3前端
    ├── package.json                 # 前端依赖配置
    ├── vite.config.js               # Vite构建配置
    ├── index.html                   # HTML入口
    │
    ├── src/
    │   ├── main.js                  # Vue应用入口
    │   ├── App.vue                  # 根组件
    │   ├── style.css                # 全局样式（CSS变量主题）
    │   │
    │   ├── api/
    │   │   └── index.js             # API服务封装
    │   │
    │   ├── i18n/                    # 国际化文件
    │   │   ├── zh-CN.json           # 中文
    │   │   ├── en.json              # 英文
    │   │   └── ru.json              # 俄语
    │   │
    │   └── components/              # Vue组件
    │       ├── TransferTab.vue      # 文件传输标签页
    │       ├── DownloadTab.vue      # 下载管理标签页
    │       ├── HistoryTab.vue       # 历史记录标签页
    │       ├── LogTab.vue           # 日志查看标签页
    │       ├── EncryptionTab.vue    # 加密工具标签页
    │       ├── PerformanceTab.vue   # 性能监控标签页
    │       ├── EnvironmentTab.vue   # 环境检测标签页
    │       └── SettingsTab.vue      # 设置标签页
    │
    └── wailsjs/                     # Wails生成的JS绑定
        ├── go/app/                  # Go函数绑定
        └── runtime/                 # Wails运行时
```

### 1.2 数据存储位置

| 平台 | 配置目录 | 日志目录 | 数据库 |
|------|----------|----------|--------|
| **macOS** | `~/Library/Application Support/LANftt/` | `~/Library/Logs/LANftt/` | `~/Library/Application Support/LANftt/history.db` |
| **Windows** | `%LOCALAPPDATA%/LANftt/` | `%LOCALAPPDATA%/LANftt/logs/` | `%LOCALAPPDATA%/LANftt/history.db` |
| **Linux** | `~/.config/lanftt/` | `~/.config/lanftt/logs/` | `~/.config/lanftt/history.db` |

---

## 二、技术架构

### 2.1 后端技术栈 (Go)

| 模块 | 技术 | 说明 |
|------|------|------|
| **框架** | Wails v2 | 桌面应用框架，Go + WebView |
| **HTTP** | Gin | HTTP服务器框架 |
| **数据库** | SQLite (go-sqlite3) | 历史记录和文件注册表存储 |
| **加密** | AES-256-GCM | Token加密、文件加密 |
| **密钥派生** | scrypt | 密钥派生函数 |
| **校验** | SHA256 | 文件完整性校验 |
| **WebSocket** | gorilla/websocket | 实时通信协议 |
| **UDP** | net/udp | UDP广播发现和传输 |
| **P2P** | net/tcp | 点对点直连传输 |
| **配置** | gopkg.in/yaml.v3 | YAML配置解析 |

### 2.2 前端技术栈 (Vue 3)

| 模块 | 技术 | 说明 |
|------|------|------|
| **框架** | Vue 3 | 渐进式JavaScript框架 |
| **构建** | Vite | 下一代前端构建工具 |
| **状态** | Composition API | 组合式API |
| **主题** | CSS Variables | 动态主题切换（浅色/深色） |
| **国际化** | vue-i18n | 多语言支持 |
| **二维码** | qrcode | 二维码生成 |

### 2.3 架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                        用户界面 (Vue 3)                          │
├─────────┬─────────┬─────────┬─────────┬─────────┬──────────────┤
│Transfer │Download │ History │   Log   │Settings │ Encryption   │
│   Tab   │   Tab   │   Tab   │   Tab   │   Tab   │     Tab      │
└────┬────┴────┬────┴────┬────┴────┬────┴────┬────┴──────┬───────┘
     │         │         │         │         │           │
     └─────────┴─────────┴─────────┴─────────┴───────────┘
                              │
                    ┌─────────▼─────────┐
                    │   Wails Runtime   │
                    │   (JS <-> Go)     │
                    └─────────┬─────────┘
                              │
┌─────────────────────────────▼─────────────────────────────────┐
│                     应用层 (internal/app)                       │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │                    App (app.go)                          │  │
│  │  - 服务编排和生命周期管理                                    │  │
│  │  - Wails函数绑定                                          │  │
│  │  - 前后端通信桥梁                                          │  │
│  └─────────────────────────────────────────────────────────┘  │
└─────────────────────────────┬─────────────────────────────────┘
                              │
┌─────────────────────────────▼─────────────────────────────────┐
│                      服务层 (internal/)                         │
│                                                                │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │   Transfer   │  │    Token     │  │  Discovery   │         │
│  │   Service    │  │   Manager    │  │   Service    │         │
│  │  文件传输     │  │  Token管理    │  │  设备发现     │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
│                                                                │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │  Encryption  │  │   Protocol   │  │ Performance  │         │
│  │   Service    │  │   Selector   │  │   Monitor    │         │
│  │  加密服务     │  │  协议选择器   │  │  性能监控     │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
│                                                                │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │   Storage    │  │   Access     │  │ Environment  │         │
│  │   Service    │  │   Control    │  │   Checker    │         │
│  │  数据存储     │  │  访问控制     │  │  环境检测     │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
└────────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────▼─────────────────────────────────┐
│                      基础设施层 (pkg/)                          │
│                                                                │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │   Logger     │  │   Utils      │  │  Constants   │         │
│  │  日志系统     │  │  工具函数     │  │  常量定义     │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
│                                                                │
│  ┌──────────────┐                                              │
│  │   Errors     │                                              │
│  │  错误定义     │                                              │
│  └──────────────┘                                              │
└────────────────────────────────────────────────────────────────┘
```

---

## 三、业务逻辑架构

### 3.1 核心服务说明

#### 3.1.1 文件传输服务 (TransferService)

**职责**: 文件注册、下载、批量下载

**核心方法**:
- `RegisterFile(filePath)` - 注册文件，返回FileInfo
- `GetFileInfo(id)` - 获取文件信息
- `DownloadFile(id, savePath, progressChan)` - 下载单个文件
- `DownloadFileByPath(filePath, fileName, fileSize, savePath, progressChan)` - 按路径下载（支持跨重启）
- `BatchDownload(fileIDs, savePath, progressChan)` - 批量下载

**数据结构**:
```go
type FileInfo struct {
    ID           string    // 文件唯一标识
    Path         string    // 文件绝对路径
    Name         string    // 文件名
    Size         int64     // 文件大小
    Type         string    // 文件类型
    Checksum     string    // SHA256校验和
    ModTime      time.Time // 修改时间
    RegisteredAt time.Time // 注册时间
}
```

#### 3.1.2 Token管理器 (TokenManager)

**职责**: Token生成、验证、加密

**核心方法**:
- `GenerateToken(fileID, expiry)` - 生成Token
- `GenerateTokenWithFileInfo(fileID, expiry, fileName, fileSize, filePath)` - 生成包含文件元数据的Token
- `ValidateToken(token)` - 验证Token
- `ValidateTokenWithKey(token, customKey)` - 使用自定义密钥验证
- `ParseEncryptedToken(token, customKey)` - 解析加密Token

**Token数据结构**:
```go
type TokenData struct {
    FileID   string `json:"file_id"`    // 文件ID
    Expiry   int64  `json:"expiry"`     // 过期时间戳
    Type     string `json:"type"`       // Token类型
    Checksum string `json:"checksum"`   // 校验和
    FileName string `json:"file_name"`  // 文件名（跨重启支持）
    FileSize int64  `json:"file_size"`  // 文件大小（跨重启支持）
    FilePath string `json:"file_path"`  // 文件路径（跨重启支持）
}
```

**加密方式**: AES-256-GCM

#### 3.1.3 设备发现服务 (DiscoveryService)

**职责**: UDP广播发现局域网设备

**工作流程**:
1. 监听UDP端口（默认37021）
2. 定时广播发现消息（默认5秒间隔）
3. 接收其他设备的广播消息
4. 维护在线设备列表
5. 清理超时设备（默认30秒）

**消息类型**:
- `LANFILETRANSFER_DISCOVERY` - 发现请求
- `LANFILETRANSFER_RESPONSE` - 发现响应

#### 3.1.4 协议选择器 (ProtocolSelector)

**职责**: 根据文件大小和网络条件智能选择传输协议

**选择算法**:
| 文件大小 | 推荐协议 | 原因 |
|----------|----------|------|
| > 100MB | P2P > UDP | 大文件需要高速传输 |
| 10MB - 100MB | UDP > WebSocket | 中等文件平衡速度和可靠性 |
| 1MB - 10MB | WebSocket > HTTP | 小文件注重实时性 |
| < 1MB | HTTP | 小文件使用最稳定协议 |

**协议优先级**:
1. P2P（优先级1）- 点对点直连
2. UDP（优先级2）- 高速传输
3. WebSocket（优先级3）- 实时通信
4. HTTP（优先级4）- 标准协议

#### 3.1.5 加密服务 (EncryptionService)

**职责**: AES-256-GCM加密/解密

**加密流程**:
1. 生成16字节随机盐值
2. 使用scrypt派生32字节密钥
3. AES-256-GCM加密
4. 输出格式: Base64(salt + nonce + ciphertext)

**解密流程**:
1. Base64解码
2. 提取salt（前16字节）
3. scrypt派生密钥
4. AES-256-GCM解密

#### 3.1.6 存储服务 (Storage)

**职责**: SQLite数据持久化

**数据表**:

1. **history** - 历史记录表
```sql
CREATE TABLE history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_name TEXT NOT NULL,
    file_size INTEGER,
    file_path TEXT,
    action TEXT NOT NULL,        -- upload/download
    status TEXT NOT NULL,        -- completed/failed/pending
    protocol TEXT,
    download_link TEXT,
    duration INTEGER,
    created_at DATETIME,
    updated_at DATETIME
);
```

2. **file_registry** - 文件注册表
```sql
CREATE TABLE file_registry (
    id TEXT PRIMARY KEY,
    file_path TEXT NOT NULL,
    file_name TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    checksum TEXT,
    content_type TEXT,
    download_count INTEGER DEFAULT 0,
    created_at DATETIME,
    expires_at DATETIME
);
```

### 3.2 配置系统

#### 3.2.1 配置加载顺序

```
┌─────────────────┐
│  默认配置        │  constants.go 中的默认值
│  (Default)      │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  系统配置        │  config.yaml
│  (System)       │  覆盖默认配置
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  用户配置        │  user_config.json
│  (User)         │  覆盖系统配置
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  运行时配置      │  临时修改
│  (Runtime)      │  最高优先级
└─────────────────┘
```

#### 3.2.2 配置文件结构

**config.yaml**:
```yaml
app:
  name: "LAN-File-Transfer-Tool"
  short_name: "LANftt"
  version: "0.2.0"

server:
  port: 8080              # HTTP服务端口
  host: "0.0.0.0"        # 监听地址

discovery:
  enabled: true           # 启用设备发现
  port: 37021             # 发现服务端口
  broadcast_interval: 5   # 广播间隔(秒)
  peer_timeout: 30        # 节点超时(秒)

websocket:
  enabled: true           # 启用WebSocket
  port_offset: 0          # 端口偏移
  chunk_size: 65536       # 数据块大小(64KB)

udp:
  enabled: true           # 启用UDP
  port: 37022             # UDP端口
  chunk_size: 32768       # 数据块大小(32KB)

p2p:
  enabled: true           # 启用P2P
  port: 37023             # P2P端口
  chunk_size: 65536       # 数据块大小(64KB)

transfer:
  max_connections: 10     # 最大连接数
  chunk_size: 1048576     # 分片大小(1MB)
  enable_resume: true     # 启用断点续传

security:
  token_expiry: 86400     # Token有效期(秒)
  secret_key: "..."       # 加密密钥
  whitelist: []           # IP白名单
  blacklist: []           # IP黑名单

history:
  max_records: 100        # 最大历史记录数

performance:
  pool_size: 10           # 线程池大小
  monitor_interval: 2     # 监控间隔(秒)
```

**user_config.json**:
```json
{
  "theme": "light",
  "language": "zh-CN",
  "selected_ip": "192.168.1.100",
  "settings": {
    "transfer": {
      "defaultProtocol": "auto",
      "chunkSize": 1048576,
      "maxConnections": 10
    },
    "protocols": {
      "websocket": true,
      "udp": true,
      "p2p": true,
      "discovery": true
    }
  }
}
```

---

## 四、API接口

### 4.1 Wails绑定函数（前端调用）

#### 文件传输
| 函数 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| `SelectFiles(directory)` | bool | `[]map[string]interface{}` | 选择文件或文件夹 |
| `GenerateDownloadLink(filePath)` | string | `map[string]interface{}` | 生成下载链接 |
| `GenerateDownloadLinkForFile(fileID)` | string | `map[string]interface{}` | 根据文件ID生成链接 |
| `GenerateBatchDownloadLink(filePaths)` | `[]string` | `map[string]interface{}` | 批量生成链接 |
| `GetAvailableFiles()` | - | `[]map[string]interface{}` | 获取可用文件列表 |
| `GetDownloadInfo(token)` | string | `map[string]interface{}` | 获取下载信息 |
| `DownloadFile(token, savePath)` | string, string | `map[string]interface{}` | 下载文件 |
| `BatchDownload(fileIDs, savePath)` | `[]string`, string | `map[string]interface{}` | 批量下载 |

#### 网络功能
| 函数 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| `GetServerInfo()` | - | `map[string]interface{}` | 获取服务器信息 |
| `GetAllIPs()` | - | `[]map[string]interface{}` | 获取所有IP地址 |
| `SetSelectedIP(ip)` | string | error | 设置选中IP |
| `GetSelectedIP()` | - | string | 获取选中IP |
| `SetManualIP(ip)` | string | error | 设置手动IP |
| `DiscoverPeers()` | - | `[]map[string]interface{}` | 发现设备 |

#### 加密功能
| 函数 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| `EncryptData(plainText, key)` | string, string | string | 加密数据 |
| `DecryptData(cipherText, key)` | string, string | string | 解密数据 |
| `GenerateEncryptionKey()` | - | string | 生成密钥 |
| `ParseEncryptedToken(token, key)` | string, string | `map[string]interface{}` | 解析加密Token |
| `GetDownloadInfoWithKey(token, key)` | string, string | `map[string]interface{}` | 使用密钥获取下载信息 |

#### 校验功能
| 函数 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| `CalculateChecksum(filePath)` | string | string | 计算SHA256 |
| `VerifyFile(filePath, expectedChecksum)` | string, string | bool | 验证文件完整性 |

#### 性能监控
| 函数 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| `GetPerformanceStats()` | - | `map[string]interface{}` | 获取性能统计 |
| `InitThreadPool(size)` | int | error | 初始化线程池 |
| `StopThreadPool()` | - | error | 停止线程池 |

#### 环境检测
| 函数 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| `CheckEnvironment()` | - | `map[string]interface{}` | 环境检测 |

#### 用户配置
| 函数 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| `GetUserConfig()` | - | `map[string]interface{}` | 获取用户配置 |
| `SaveUserConfig(configData)` | `map[string]interface{}` | error | 保存用户配置 |
| `ResetUserConfig()` | - | error | 重置配置 |

#### 历史记录
| 函数 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| `GetHistory(limit)` | int | `[]map[string]interface{}` | 获取历史记录 |
| `ClearHistory()` | - | error | 清除历史记录 |
| `DeleteHistory(id)` | int | error | 删除单条记录 |
| `RegenerateLink(historyID)` | int | `map[string]interface{}` | 重新生成链接 |

#### 日志功能
| 函数 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| `GetLogs()` | - | `[]map[string]interface{}` | 获取所有日志 |
| `ClearLogs()` | - | - | 清除日志 |
| `LogInfo(message)` | string | - | 记录信息日志 |
| `LogWarn(message)` | string | - | 记录警告日志 |
| `LogError(message)` | string | - | 记录错误日志 |
| `LogDebug(message)` | string | - | 记录调试日志 |

#### 文件操作
| 函数 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| `SelectSaveFile(defaultFilename)` | string | string | 选择保存路径 |
| `SaveTextFile(filePath, content)` | string, string | error | 保存文本文件 |
| `ReadTextFile(filePath)` | string | string | 读取文本文件 |

#### 协议管理
| 函数 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| `GetProtocolStatus()` | - | `map[string]interface{}` | 获取协议状态 |
| `GetProtocolRecommendation(fileSize)` | int64 | `map[string]interface{}` | 获取协议推荐 |
| `SetProtocolPreference(pref)` | string | error | 设置协议偏好 |
| `SelectProtocol(fileSize, peerAvailable, userOverride)` | int64, bool, string | string | 选择传输协议 |

#### 应用信息
| 函数 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| `GetAppInfo()` | - | `map[string]interface{}` | 获取应用信息 |
| `GetSecurityInfo()` | - | `map[string]interface{}` | 获取安全信息 |

### 4.2 HTTP API接口

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
POST /api/user/config/reset     # 重置配置

GET  /api/performance/stats     # 获取性能统计
POST /api/performance/pool/init # 初始化线程池
POST /api/performance/pool/stop # 停止线程池

POST /api/encryption/encrypt    # 加密数据
POST /api/encryption/decrypt    # 解密数据
POST /api/encryption/key        # 生成密钥

GET  /api/environment/check     # 环境检测

GET  /api/history               # 获取历史记录
DELETE /api/history             # 清除历史记录

GET  /download/:token           # 下载文件
GET  /download/batch/:token     # 批量下载
```

---

## 五、数据流

### 5.1 文件分享流程

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  选择文件    │───▶│  注册文件    │───▶│  生成Token   │
│  (前端)     │    │(TransferSvc)│    │(TokenMgr)   │
└─────────────┘    └─────────────┘    └──────┬──────┘
                                             │
                   ┌─────────────┐           │
                   │  返回链接    │◀──────────┘
                   │  + 二维码   │
                   └─────────────┘
```

### 5.2 文件下载流程

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  打开链接    │───▶│  验证Token   │───▶│  获取文件信息 │
│  (浏览器)   │    │(TokenMgr)   │    │(TransferSvc)│
└─────────────┘    └─────────────┘    └──────┬──────┘
                                             │
                   ┌─────────────┐           │
                   │  下载文件    │◀──────────┘
                   │  + 记录历史  │
                   └─────────────┘
```

### 5.3 设备发现流程

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  UDP广播    │───▶│  接收广播    │───▶│  更新设备列表 │
│  (定时5秒)  │    │  (监听)     │    │  (内存)     │
└─────────────┘    └─────────────┘    └─────────────┘
       │                                     │
       │         ┌─────────────┐             │
       └────────▶│  清理超时    │◀────────────┘
                  │  (30秒)    │
                  └─────────────┘
```

---

## 六、安全机制

### 6.1 Token安全

- **加密方式**: AES-256-GCM
- **密钥长度**: 32字节
- **Token内容**: 加密的JSON（包含文件ID、过期时间、文件元数据）
- **有效期**: 默认24小时

### 6.2 访问控制

- **IP白名单**: 只允许指定IP访问
- **IP黑名单**: 禁止指定IP访问
- **优先级**: 黑名单 > 白名单

### 6.3 文件校验

- **算法**: SHA256
- **时机**: 文件注册时计算，下载后可验证

---

## 七、性能优化

### 7.1 线程池

- **默认大小**: 10个工作线程
- **最大任务队列**: 100个任务
- **用途**: 并发文件传输

### 7.2 传输优化

- **分片大小**: 1MB（可配置）
- **断点续传**: 支持
- **多协议**: 根据文件大小自动选择

### 7.3 内存优化

- **日志内存限制**: 最多1000条
- **历史记录限制**: 最多100条
- **流式传输**: 大文件分块读写

---

## 八、错误处理

### 8.1 错误类型

```go
var (
    ErrFileNotFound    = errors.New("文件不存在")
    ErrTokenExpired    = errors.New("Token已过期")
    ErrInvalidToken    = errors.New("无效的Token")
    ErrAccessDenied    = errors.New("访问被拒绝")
    ErrTransferFailed  = errors.New("传输失败")
)
```

### 8.2 日志级别

- **DEBUG**: 调试信息
- **INFO**: 正常信息
- **WARN**: 警告信息
- **ERROR**: 错误信息

---

## 九、国际化

支持三种语言：
- **中文** (zh-CN)
- **英文** (en)
- **俄语** (ru)

翻译文件位于 `frontend/src/i18n/` 目录。

---

## 十、构建与部署

### 10.1 开发环境

```bash
# 安装依赖
go mod tidy
cd frontend && npm install

# 开发模式运行
wails dev
```

### 10.2 生产构建

```bash
# 构建当前平台
wails build

# 构建指定平台
wails build -platform darwin/amd64
wails build -platform darwin/arm64
wails build -platform windows/amd64
wails build -platform linux/amd64
```

### 10.3 构建产物

| 平台 | 产物 |
|------|------|
| macOS | `LAN-File-Transfer-Tool.app` |
| Windows | `LAN-File-Transfer-Tool.exe` |
| Linux | `LAN-File-Transfer-Tool` |

---

## 十一、许可证

MIT License
