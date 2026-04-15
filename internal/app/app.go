package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"lanfiletransfertool/internal/access"
	"lanfiletransfertool/internal/checksum"
	"lanfiletransfertool/internal/config"
	"lanfiletransfertool/internal/discovery"
	"lanfiletransfertool/internal/encryption"
	"lanfiletransfertool/internal/environment"
	"lanfiletransfertool/internal/p2p"
	"lanfiletransfertool/internal/performance"
	"lanfiletransfertool/internal/protocol"
	"lanfiletransfertool/internal/resume"
	"lanfiletransfertool/internal/server"
	"lanfiletransfertool/internal/stats"
	"lanfiletransfertool/internal/storage"
	"lanfiletransfertool/internal/token"
	"lanfiletransfertool/internal/transfer"
	"lanfiletransfertool/internal/udp"
	"lanfiletransfertool/internal/userconfig"
	"lanfiletransfertool/internal/websocket"
	"lanfiletransfertool/pkg/constants"
	"lanfiletransfertool/pkg/logger"
	"lanfiletransfertool/pkg/utils"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx              context.Context
	config           *config.Config
	server           *server.Server
	storage          *storage.Storage
	tokenManager     *token.Manager
	accessControl    *access.Control
	transferService  *transfer.Service
	resumeService    *resume.Service
	encryptionSvc    *encryption.Service
	checksumSvc      *checksum.Service
	perfMonitor      *performance.Monitor
	envChecker       *environment.Checker
	userConfig       *userconfig.Manager
	discoverySvc     *discovery.Service
	wsService        *websocket.Service
	udpService       *udp.Service
	p2pService       *p2p.Service
	statsMonitor     *stats.Monitor
	protocolSelector *protocol.Selector
	mu               sync.RWMutex
}

func NewApp() *App {
	return &App{}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	var err error

	// 步骤1: 创建数据目录
	dataDir := filepath.Join(utils.GetExecutableDir(), "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("创建数据目录失败: %v", err)
	}

	// 步骤2: 初始化配置管理器（遵循：初始化 -> 系统配置 -> 用户配置覆盖）
	cfgManager, err := config.NewManager()
	if err != nil {
		logger.Warn("配置管理器初始化失败，使用默认配置: %v", err)
		a.config = config.DefaultConfig()
	} else {
		a.config = cfgManager.GetConfig()
	}

	// 步骤3: 初始化存储
	dbPath := filepath.Join(dataDir, "history.db")
	a.storage, err = storage.NewStorage(dbPath)
	if err != nil {
		log.Fatalf("初始化存储失败: %v", err)
	}

	// 步骤4: 初始化各个服务，使用配置值而非硬编码
	a.tokenManager = token.NewManager(a.config.Security.TokenExpiry, a.config.Security.SecretKey)
	a.accessControl = access.NewControl(a.config.Security.Whitelist, a.config.Security.Blacklist)
	a.transferService = transfer.NewService(&a.config.Transfer, dataDir)
	a.resumeService = resume.NewService(dataDir)
	a.encryptionSvc = encryption.NewService()
	a.checksumSvc = checksum.NewService()
	a.perfMonitor = performance.NewMonitor()
	a.envChecker = environment.NewChecker()
	a.statsMonitor = stats.GetMonitor()

	// 步骤5: 初始化用户配置管理器
	userConfigPath := filepath.Join(dataDir, "user_config.json")
	a.userConfig, err = userconfig.NewManager(userConfigPath)
	if err != nil {
		log.Fatalf("初始化用户配置失败: %v", err)
	}

	// 步骤6: 初始化协议选择器
	a.protocolSelector = protocol.NewSelector(a.config)
	// 加载用户偏好协议
	if userCfg, err := a.userConfig.GetConfig(); err == nil && userCfg.Settings != nil {
		if transferSettings, ok := userCfg.Settings["transfer"].(map[string]interface{}); ok {
			if pref, ok := transferSettings["defaultProtocol"].(string); ok && pref != "" {
				a.protocolSelector.SetPreference(protocol.Protocol(pref))
			}
		}
	}
	logger.Info("协议选择器已初始化，偏好协议: %s", a.protocolSelector.GetPreference())

	// 步骤7: 创建HTTP服务器
	a.server = server.NewServer(a.config, a.storage, a.tokenManager, a.accessControl, a.transferService, a.resumeService, a.encryptionSvc, a.checksumSvc, a.perfMonitor, a.envChecker, a.userConfig)

	// 启动UDP发现服务（使用配置值）
	a.discoverySvc = discovery.NewService(a.config, a.config.Server.Port, "LANFileTransfer")
	if err := a.discoverySvc.Start(); err != nil {
		logger.Warn("UDP发现服务启动失败: %v", err)
	}

	// 启动WebSocket传输服务
	if a.config.WebSocket.Enabled {
		a.wsService = websocket.NewService()
		logger.Info("WebSocket传输服务已创建")
	}

	// 启动UDP传输服务（使用配置值）
	if a.config.UDP.Enabled {
		a.udpService = udp.NewService(a.config, func(fileID string) (string, string, int64, error) {
			info, err := a.transferService.GetFileInfo(fileID)
			if err != nil {
				return "", "", 0, err
			}
			return info.Path, info.Name, info.Size, nil
		})
		if err := a.udpService.Start(); err != nil {
			logger.Warn("UDP传输服务启动失败: %v", err)
		}
	}

	// 启动P2P传输服务（使用配置值）
	if a.config.P2P.Enabled {
		a.p2pService = p2p.NewService(a.config, func(fileID string) (string, string, int64, error) {
			info, err := a.transferService.GetFileInfo(fileID)
			if err != nil {
				return "", "", 0, err
			}
			return info.Path, info.Name, info.Size, nil
		})
		if err := a.p2pService.Start(); err != nil {
			logger.Warn("P2P传输服务启动失败: %v", err)
		}
	}

	go func() {
		if err := a.server.Start(); err != nil {
			logger.Error("服务器启动失败: %v", err)
		}
	}()

	logger.Info("应用启动成功 - LANftt v%s", a.config.App.Version)
}

func (a *App) DomReady(ctx context.Context) {
}

func (a *App) BeforeClose(ctx context.Context) bool {
	logger.Info("应用即将关闭，开始清理资源...")

	if a.discoverySvc != nil {
		a.discoverySvc.Stop()
	}
	if a.wsService != nil {
		logger.Info("WebSocket服务已停止")
	}
	if a.udpService != nil {
		a.udpService.Stop()
	}
	if a.p2pService != nil {
		a.p2pService.Stop()
	}
	if a.server != nil {
		if err := a.server.Stop(); err != nil {
			logger.Error("停止服务器失败: %v", err)
		}
	}
	if a.perfMonitor != nil {
		a.perfMonitor.StopPool()
	}
	if a.storage != nil {
		a.storage.Close()
	}

	logger.Info("资源清理完成，允许关闭窗口")
	return false
}

func (a *App) Shutdown(ctx context.Context) {
	logger.Info("应用已关闭")
}

func (a *App) GetServerInfo() map[string]interface{} {
	return map[string]interface{}{
		"host":    a.config.Server.Host,
		"port":    a.config.Server.Port,
		"running": a.server.IsRunning(),
		"url":     fmt.Sprintf("http://%s:%d", a.getServerIP(), a.config.Server.Port),
	}
}

func (a *App) getServerIP() string {
	selectedIP := a.userConfig.GetSelectedIP()
	if selectedIP != "" {
		return selectedIP
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}

func (a *App) GetAllIPs() ([]map[string]interface{}, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("获取网络接口失败: %w", err)
	}

	var ips []map[string]interface{}
	seen := make(map[string]bool)

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ipStr := ipnet.IP.String()
				if !seen[ipStr] {
					seen[ipStr] = true
					ips = append(ips, map[string]interface{}{
						"ip":    ipStr,
						"mask":  ipnet.Mask.String(),
						"scope": getIPScope(ipStr),
					})
				}
			}
		}
	}

	if len(ips) == 0 {
		ips = append(ips, map[string]interface{}{
			"ip":    "127.0.0.1",
			"mask":  "255.0.0.0",
			"scope": "loopback",
		})
	}

	return ips, nil
}

func getIPScope(ip string) string {
	parts := [4]int{}
	fmt.Sscanf(ip, "%d.%d.%d.%d", &parts[0], &parts[1], &parts[2], &parts[3])

	switch {
	case parts[0] == 10:
		return "private-a"
	case parts[0] == 172 && parts[1] >= 16 && parts[1] <= 31:
		return "private-b"
	case parts[0] == 192 && parts[1] == 168:
		return "private-c"
	case parts[0] == 169 && parts[1] == 254:
		return "link-local"
	case parts[0] >= 224 && parts[0] <= 239:
		return "multicast"
	default:
		return "public"
	}
}

func (a *App) SetSelectedIP(ip string) error {
	return a.userConfig.SetSelectedIP(ip)
}

func (a *App) GetSelectedIP() (string, error) {
	return a.userConfig.GetSelectedIP(), nil
}

func (a *App) SelectFiles(directory bool) ([]map[string]interface{}, error) {
	if a.ctx == nil {
		return nil, fmt.Errorf("应用上下文未初始化")
	}

	var selection string
	var err error

	if directory {
		// 选择文件夹
		selection, err = runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
			Title: "选择文件夹",
		})
		logger.Info("文件夹选择结果：%s, 错误：%v", selection, err)
	} else {
		// 选择文件 - 不使用过滤器，允许选择所有文件
		selection, err = runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
			Title: "选择文件",
		})
		logger.Info("文件选择结果：%s, 错误：%v", selection, err)
	}

	if err != nil {
		logger.Error("文件选择失败：%w", err)
		return nil, fmt.Errorf("打开对话框失败：%w", err)
	}

	// 用户取消选择
	if selection == "" {
		logger.Info("用户取消了文件选择")
		return []map[string]interface{}{}, nil
	}

	logger.Info("选中的文件路径：%s", selection)

	// 获取文件信息
	fileInfo, err := os.Stat(selection)
	if err != nil {
		logger.Error("获取文件信息失败：%w", err)
		return nil, fmt.Errorf("获取文件信息失败：%w", err)
	}

	logger.Info("文件信息：名称=%s, 大小=%d, 是目录=%v", fileInfo.Name(), fileInfo.Size(), fileInfo.IsDir())

	result := []map[string]interface{}{
		{
			"name":     fileInfo.Name(),
			"path":     selection,
			"size":     fileInfo.Size(),
			"is_dir":   fileInfo.IsDir(),
			"mod_time": fileInfo.ModTime().Format(time.RFC3339),
		},
	}

	return result, nil
}

func (a *App) GenerateDownloadLink(filePath string) (map[string]interface{}, error) {
	fileInfo, err := a.transferService.RegisterFile(filePath)
	if err != nil {
		return nil, err
	}

	// 使用新的方法生成包含文件元数据的token
	downloadToken := a.tokenManager.GenerateTokenWithFileInfo(
		fileInfo.ID,
		constants.DefaultTokenExpiryDownload,
		fileInfo.Name,
		fileInfo.Size,
		fileInfo.Path,
	)

	link := fmt.Sprintf("http://%s:%d/download/%s", a.getServerIP(), a.config.Server.Port, downloadToken)

	qrCode, _ := utils.GenerateQRCode(link)

	// // 添加历史记录
	// record := &storage.HistoryRecord{
	// 	FileName: fileInfo.Name,
	// 	FileSize: fileInfo.Size,
	// 	Action:   "upload",
	// 	Status:   "completed",
	// }
	// a.storage.AddHistory(record)

	return map[string]interface{}{
		"token":     downloadToken,
		"link":      link,
		"qr_code":   qrCode,
		"file_id":   fileInfo.ID,
		"file_name": fileInfo.Name,
		"file_size": fileInfo.Size,
	}, nil
}

// GenerateDownloadLinkForFile 根据文件ID生成下载链接
func (a *App) GenerateDownloadLinkForFile(fileID string) (map[string]interface{}, error) {
	// 验证文件是否存在
	fileInfo, err := a.transferService.GetFileInfo(fileID)
	if err != nil {
		return nil, fmt.Errorf("文件不存在: %v", err)
	}

	// 生成包含文件元数据的下载token
	downloadToken := a.tokenManager.GenerateTokenWithFileInfo(
		fileInfo.ID,
		constants.DefaultTokenExpiryDownload,
		fileInfo.Name,
		fileInfo.Size,
		fileInfo.Path,
	)

	return map[string]interface{}{
		"token": downloadToken,
		"link":  fmt.Sprintf("http://%s:%d/download/%s", a.getServerIP(), a.config.Server.Port, downloadToken),
	}, nil
}

// GenerateBatchDownloadLink 批量生成下载链接（多个文件一个链接）
func (a *App) GenerateBatchDownloadLink(filePaths []string) (map[string]interface{}, error) {
	if len(filePaths) == 0 {
		return nil, fmt.Errorf("没有选择文件")
	}

	// 注册所有文件
	var fileInfos []*transfer.FileInfo
	var totalSize int64
	var fileIDs []string
	for _, path := range filePaths {
		fileInfo, err := a.transferService.RegisterFile(path)
		if err != nil {
			return nil, err
		}
		fileInfos = append(fileInfos, fileInfo)
		fileIDs = append(fileIDs, fileInfo.ID)
		totalSize += fileInfo.Size
	}

	// 将文件ID列表编码为逗号分隔的字符串
	batchFileIDs := "batch:" + strings.Join(fileIDs, ",")

	// 构建文件列表JSON用于存储在token中
	var fileList []map[string]interface{}
	for _, info := range fileInfos {
		fileList = append(fileList, map[string]interface{}{
			"file_id":   info.ID,
			"file_name": info.Name,
			"file_size": info.Size,
			"file_path": info.Path,
		})
	}

	// 生成批量token
	batchToken := a.tokenManager.GenerateTokenWithFileInfo(
		batchFileIDs,
		constants.DefaultTokenExpiryDownload,
		fmt.Sprintf("批量文件 (%d个)", len(fileInfos)),
		totalSize,
		"",
	)

	// 构建批量下载链接
	link := fmt.Sprintf("http://%s:%d/download/batch/%s", a.getServerIP(), a.config.Server.Port, batchToken)

	qrCode, _ := utils.GenerateQRCode(link)

	return map[string]interface{}{
		"token":      batchToken,
		"link":       link,
		"qr_code":    qrCode,
		"file_count": len(fileInfos),
		"total_size": totalSize,
		"files":      fileList,
	}, nil
}

func (a *App) GetAvailableFiles() ([]map[string]interface{}, error) {
	files, err := a.transferService.GetAvailableFiles()
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(files))
	for i, f := range files {
		result[i] = map[string]interface{}{
			"id":       f.ID,
			"name":     f.Name,
			"size":     f.Size,
			"type":     f.Type,
			"mod_time": f.ModTime.Format(time.RFC3339),
		}
	}
	return result, nil
}

func (a *App) GetDownloadInfo(token string) (map[string]interface{}, error) {
	info, err := a.tokenManager.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	// 检查是否是批量下载token
	if strings.HasPrefix(info.FileID, "batch:") {
		return a.getBatchDownloadInfo(info.FileID)
	}

	// 优先从token中获取文件信息（支持跨重启解析）
	if info.FileName != "" && info.FileSize > 0 {
		return map[string]interface{}{
			"file_id":   info.FileID,
			"file_name": info.FileName,
			"file_size": info.FileSize,
			"checksum":  "",
			"is_batch":  false,
		}, nil
	}

	// 回退：从内存中获取文件信息
	fileInfo, err := a.transferService.GetFileInfo(info.FileID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"file_id":   fileInfo.ID,
		"file_name": fileInfo.Name,
		"file_size": fileInfo.Size,
		"checksum":  fileInfo.Checksum,
		"is_batch":  false,
	}, nil
}

// getBatchDownloadInfo 获取批量下载信息
func (a *App) getBatchDownloadInfo(batchFileIDs string) (map[string]interface{}, error) {
	// 解析文件ID列表
	idsStr := strings.TrimPrefix(batchFileIDs, "batch:")
	fileIDs := strings.Split(idsStr, ",")

	var fileList []map[string]interface{}
	var totalSize int64
	for _, fileID := range fileIDs {
		fileInfo, err := a.transferService.GetFileInfo(fileID)
		if err != nil {
			continue
		}
		fileList = append(fileList, map[string]interface{}{
			"file_id":   fileInfo.ID,
			"file_name": fileInfo.Name,
			"file_size": fileInfo.Size,
		})
		totalSize += fileInfo.Size
	}

	if len(fileList) == 0 {
		return nil, fmt.Errorf("没有有效的文件")
	}

	return map[string]interface{}{
		"file_name":  fmt.Sprintf("批量文件 (%d个)", len(fileList)),
		"file_size":  totalSize,
		"file_count": len(fileList),
		"files":      fileList,
		"is_batch":   true,
	}, nil
}

func (a *App) DownloadFile(token string, savePath string) (map[string]interface{}, error) {
	info, err := a.tokenManager.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	clientIP := "localhost"
	if !a.accessControl.AllowAccess(clientIP) {
		return nil, fmt.Errorf("访问被拒绝")
	}

	var result *transfer.TransferResult
	var downloadErr error
	var fileName string
	var fileSize int64

	// 优先使用token中的文件路径（支持跨重启解析）
	if info.FilePath != "" && info.FileName != "" {
		fileName = info.FileName
		fileSize = info.FileSize
		result, downloadErr = a.transferService.DownloadFileByPath(info.FilePath, info.FileName, info.FileSize, savePath, nil)
	} else {
		// 回退：从内存中获取文件信息
		fileInfo, err := a.transferService.GetFileInfo(info.FileID)
		if err != nil {
			return nil, err
		}
		fileName = fileInfo.Name
		fileSize = fileInfo.Size
		result, downloadErr = a.transferService.DownloadFile(info.FileID, savePath, nil)
	}

	if downloadErr != nil {
		logger.Error("下载失败: %v", downloadErr)
		// 添加失败历史记录
		record := &storage.HistoryRecord{
			FileName: fileName,
			FileSize: fileSize,
			Action:   "download",
			Status:   "failed",
		}
		a.storage.AddHistory(record)
		return nil, downloadErr
	}

	// 添加成功历史记录
	record := &storage.HistoryRecord{
		FileName: fileName,
		FileSize: fileSize,
		Action:   "download",
		Status:   "completed",
	}
	a.storage.AddHistory(record)

	return map[string]interface{}{
		"received_bytes": result.ReceivedBytes,
		"progress":       100,
		"status":         "completed",
	}, nil
}

func (a *App) BatchDownload(fileIDs []string, savePath string) (map[string]interface{}, error) {
	progress := make(chan map[string]interface{})

	go func() {
		result, err := a.transferService.BatchDownload(fileIDs, savePath, progress)
		if err != nil {
			logger.Error("批量下载失败: %v", err)
			return
		}
		_ = result
	}()

	status := <-progress
	return status, nil
}

func (a *App) CalculateChecksum(filePath string) (string, error) {
	return a.checksumSvc.CalculateSHA256(filePath)
}

func (a *App) VerifyFile(filePath string, expectedChecksum string) (bool, error) {
	return a.checksumSvc.VerifyFile(filePath, expectedChecksum)
}

func (a *App) EncryptData(plainText string, key string) (string, error) {
	return a.encryptionSvc.Encrypt(plainText, key)
}

func (a *App) DecryptData(cipherText string, key string) (string, error) {
	return a.encryptionSvc.Decrypt(cipherText, key)
}

func (a *App) GenerateEncryptionKey() (string, error) {
	return a.encryptionSvc.GenerateKey()
}

func (a *App) GetPerformanceStats() (map[string]interface{}, error) {
	stats := a.perfMonitor.GetStats()
	return map[string]interface{}{
		"cpu_usage":          stats.CPUUsage,
		"memory_usage":       stats.MemoryUsage,
		"network_send_speed": stats.NetworkSendSpeed,
		"network_recv_speed": stats.NetworkRecvSpeed,
		"disk_read_speed":    stats.DiskReadSpeed,
		"disk_write_speed":   stats.DiskWriteSpeed,
		"active_goroutines":  stats.ActiveGoroutines,
		"timestamp":          time.Now().Format(time.RFC3339),
	}, nil
}

func (a *App) InitThreadPool(size int) error {
	return a.perfMonitor.InitPool(size)
}

func (a *App) StopThreadPool() error {
	return a.perfMonitor.StopPool()
}

func (a *App) CheckEnvironment() (map[string]interface{}, error) {
	results, err := a.envChecker.CheckAll()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"firewall":  results.Firewall,
		"network":   results.Network,
		"port":      results.Port,
		"solutions": results.Solutions,
	}, nil
}

func (a *App) GetUserConfig() (map[string]interface{}, error) {
	cfg, err := a.userConfig.GetConfig()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"theme":    cfg.Theme,
		"language": cfg.Language,
		"settings": cfg.Settings,
	}, nil
}

func (a *App) SaveUserConfig(configData map[string]interface{}) error {
	cfg := &userconfig.Config{
		Theme:    configData["theme"].(string),
		Language: configData["language"].(string),
		Settings: configData["settings"].(map[string]interface{}),
	}
	return a.userConfig.SaveConfig(cfg)
}

func (a *App) ResetUserConfig() error {
	return a.userConfig.ResetConfig()
}

func (a *App) GetHistory(limit int) ([]map[string]interface{}, error) {
	records, err := a.storage.GetHistory(limit)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(records))
	for i, r := range records {
		result[i] = map[string]interface{}{
			"id":            r.ID,
			"file_name":     r.FileName,
			"file_size":     r.FileSize,
			"file_path":     r.FilePath,
			"action":        r.Action,
			"status":        r.Status,
			"protocol":      r.Protocol,
			"download_link": r.DownloadLink,
			"duration":      r.Duration,
			"created_at":    r.CreatedAt.Format(time.RFC3339),
			"updated_at":    r.UpdatedAt.Format(time.RFC3339),
		}
	}
	return result, nil
}

func (a *App) ClearHistory() error {
	return a.storage.ClearHistory()
}

func (a *App) DeleteHistory(id int) error {
	return a.storage.DeleteHistory(id)
}

// RegenerateLink 重新生成下载链接
func (a *App) RegenerateLink(historyID int) (map[string]interface{}, error) {
	// 获取历史记录
	record, err := a.storage.GetHistoryByID(historyID)
	if err != nil {
		return nil, fmt.Errorf("获取历史记录失败: %w", err)
	}
	if record == nil {
		return nil, fmt.Errorf("历史记录不存在")
	}

	// 检查文件是否还存在
	if record.FilePath == "" {
		return nil, fmt.Errorf("无法重新生成：没有文件路径信息")
	}

	// 重新注册文件并生成新链接
	fileID := generateFileID()
	expiresAt := time.Now().Add(time.Duration(a.config.Security.TokenExpiry) * time.Hour)

	err = a.storage.RegisterFile(fileID, record.FilePath, record.FileName, record.FileSize, "", "", &expiresAt)
	if err != nil {
		return nil, fmt.Errorf("注册文件失败: %w", err)
	}

	// 生成新的Token和链接
	token := a.tokenManager.GenerateToken(fileID, time.Duration(a.config.Security.TokenExpiry)*time.Hour)

	link := fmt.Sprintf("http://%s:%d/download/%s", a.getSelectedIP(), a.config.Server.Port, token)

	// 更新历史记录
	updates := map[string]interface{}{
		"download_link": link,
		"status":        "completed",
	}
	err = a.storage.UpdateHistory(historyID, updates)
	if err != nil {
		logger.Warn("更新历史记录失败: %v", err)
	}

	return map[string]interface{}{
		"link":      link,
		"file_name": record.FileName,
		"file_size": record.FileSize,
	}, nil
}

func generateFileID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func (a *App) getSelectedIP() string {
	ip := a.userConfig.GetSelectedIP()
	if ip == "" {
		return "localhost"
	}
	return ip
}

func (a *App) DiscoverPeers() ([]map[string]interface{}, error) {
	if a.discoverySvc == nil {
		return nil, fmt.Errorf("发现服务未启动")
	}
	return a.discoverySvc.GetPeersAsMap(), nil
}

func (a *App) SetManualIP(ip string) error {
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("无效的IP地址")
	}
	return a.userConfig.SetSelectedIP(ip)
}

func (a *App) ParseEncryptedToken(token, customKey string) (map[string]interface{}, error) {
	tokenData, err := a.tokenManager.ParseEncryptedToken(token, customKey)
	if err != nil {
		return nil, err
	}

	// 优先从token中获取文件信息（支持跨重启解析）
	if tokenData.FileName != "" && tokenData.FileSize > 0 {
		return map[string]interface{}{
			"file_id":   tokenData.FileID,
			"file_name": tokenData.FileName,
			"file_size": tokenData.FileSize,
			"checksum":  "",
			"expires":   tokenData.Expiry,
		}, nil
	}

	// 回退：从内存中获取文件信息
	fileInfo, err := a.transferService.GetFileInfo(tokenData.FileID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"file_id":   fileInfo.ID,
		"file_name": fileInfo.Name,
		"file_size": fileInfo.Size,
		"checksum":  fileInfo.Checksum,
		"expires":   tokenData.Expiry,
	}, nil
}

func (a *App) GetDownloadInfoWithKey(token, customKey string) (map[string]interface{}, error) {
	info, err := a.tokenManager.ValidateTokenWithKey(token, customKey)
	if err != nil {
		return nil, err
	}

	// 检查是否是批量下载token
	if strings.HasPrefix(info.FileID, "batch:") {
		return a.getBatchDownloadInfo(info.FileID)
	}

	// 优先从token中获取文件信息（支持跨重启解析）
	if info.FileName != "" && info.FileSize > 0 {
		return map[string]interface{}{
			"file_id":   info.FileID,
			"file_name": info.FileName,
			"file_size": info.FileSize,
			"checksum":  "",
			"is_batch":  false,
		}, nil
	}

	// 回退：从内存中获取文件信息
	fileInfo, err := a.transferService.GetFileInfo(info.FileID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"file_id":   fileInfo.ID,
		"file_name": fileInfo.Name,
		"file_size": fileInfo.Size,
		"checksum":  fileInfo.Checksum,
		"is_batch":  false,
	}, nil
}

// GetProtocolStatus 获取传输协议状态
func (a *App) GetProtocolStatus() map[string]interface{} {
	status := map[string]interface{}{
		"http": map[string]interface{}{
			"enabled": true,
			"port":    a.config.Server.Port,
			"status":  "running",
		},
		"websocket": map[string]interface{}{
			"enabled": a.config.WebSocket.Enabled,
			"status":  "stopped",
		},
		"udp": map[string]interface{}{
			"enabled": a.config.UDP.Enabled,
			"port":    a.config.UDP.Port,
			"status":  "stopped",
		},
		"p2p": map[string]interface{}{
			"enabled": a.config.P2P.Enabled,
			"port":    a.config.P2P.Port,
			"status":  "stopped",
		},
	}

	if a.wsService != nil {
		status["websocket"].(map[string]interface{})["status"] = "running"
	}
	if a.udpService != nil {
		status["udp"].(map[string]interface{})["status"] = "running"
	}
	if a.p2pService != nil {
		status["p2p"].(map[string]interface{})["status"] = "running"
	}

	return status
}

// GetAppInfo 获取应用信息
func (a *App) GetAppInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":      a.config.App.Name,
		"shortName": a.config.App.ShortName,
		"version":   a.config.App.Version,
	}
}

// GetSecurityInfo 获取安全信息（包括当前使用的密钥）
func (a *App) GetSecurityInfo() map[string]interface{} {
	secretKey := a.config.Security.SecretKey
	isDefault := false
	if secretKey == "" || secretKey == "lanftt-default-secret-key-for-sharing" {
		isDefault = true
		secretKey = "lanftt-default-secret-key-for-sharing"
	}
	return map[string]interface{}{
		"secret_key": secretKey,
		"is_default": isDefault,
	}
}

// GetProtocolRecommendation 获取协议推荐
func (a *App) GetProtocolRecommendation(fileSize int64) map[string]interface{} {
	recommended := a.protocolSelector.RecommendProtocol(fileSize)
	allProtocols := a.protocolSelector.GetAllProtocols()

	protocols := make([]map[string]interface{}, 0, len(allProtocols))
	for _, p := range allProtocols {
		protocols = append(protocols, map[string]interface{}{
			"type":        string(p.Type),
			"name":        p.Name,
			"description": p.Description,
			"available":   p.Available,
			"priority":    p.Priority,
		})
	}

	return map[string]interface{}{
		"recommended": string(recommended),
		"protocols":   protocols,
		"preference":  string(a.protocolSelector.GetPreference()),
	}
}

// SetProtocolPreference 设置协议偏好
func (a *App) SetProtocolPreference(pref string) error {
	return a.protocolSelector.SetPreference(protocol.Protocol(pref))
}

// SelectProtocol 选择传输协议（用于传输前调用）
func (a *App) SelectProtocol(fileSize int64, peerAvailable bool, userOverride string) string {
	criteria := protocol.SelectionCriteria{
		FileSize:        fileSize,
		PeerAvailable:   peerAvailable,
		NetworkType:     "lan",
		RequireRealtime: false,
	}

	if userOverride != "" && userOverride != "auto" {
		criteria.UserOverride = protocol.Protocol(userOverride)
	}

	selected := a.protocolSelector.Select(criteria)
	return string(selected)
}

// GetLogs 获取所有日志
func (a *App) GetLogs() []map[string]interface{} {
	logs := logger.GetLogs()
	result := make([]map[string]interface{}, len(logs))
	for i, entry := range logs {
		result[i] = map[string]interface{}{
			"level":     entry.Level,
			"message":   entry.Message,
			"timestamp": entry.Timestamp,
		}
	}
	return result
}

// ClearLogs 清除所有日志
func (a *App) ClearLogs() {
	logger.ClearLogs()
}

// LogInfo 记录信息日志（前端调用）
func (a *App) LogInfo(message string) {
	logger.Info("[Frontend] %s", message)
}

// LogWarn 记录警告日志（前端调用）
func (a *App) LogWarn(message string) {
	logger.Warn("[Frontend] %s", message)
}

// LogError 记录错误日志（前端调用）
func (a *App) LogError(message string) {
	logger.Error("[Frontend] %s", message)
}

// LogDebug 记录调试日志（前端调用）
func (a *App) LogDebug(message string) {
	logger.Debug("[Frontend] %s", message)
}

// SaveTextFile 保存文本文件到指定路径
func (a *App) SaveTextFile(filePath string, content string) error {
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		logger.Error("保存文件失败: %v", err)
		return err
	}
	logger.Info("文件已保存: %s", filePath)
	return nil
}

// SelectSaveFile 选择保存文件路径
func (a *App) SelectSaveFile(defaultFilename string) (string, error) {
	result, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "保存文件",
		DefaultFilename: defaultFilename,
	})
	if err != nil {
		return "", err
	}
	return result, nil
}

// ReadTextFile 读取文本文件
func (a *App) ReadTextFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		logger.Error("读取文件失败: %v", err)
		return "", err
	}
	return string(content), nil
}
