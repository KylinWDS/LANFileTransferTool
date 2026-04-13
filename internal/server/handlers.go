package server

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"lanfiletransfertool/internal/stats"
	"lanfiletransfertool/internal/storage"
	"lanfiletransfertool/internal/userconfig"
	"lanfiletransfertool/pkg/logger"
	"lanfiletransfertool/pkg/utils"

	"github.com/gin-gonic/gin"
)

func (s *Server) handleSelectFiles(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "请使用客户端文件选择器"})
}

func (s *Server) handleGenerateLink(c *gin.Context) {
	var req struct {
		FilePath string `json:"file_path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件路径不能为空"})
		return
	}

	result, err := s.transferSvc.RegisterFile(req.FilePath)
	if err != nil {
		logger.Error("注册文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册文件失败"})
		return
	}

	downloadToken := s.tokenManager.GenerateToken(result.ID, 0)

	link := generateDownloadURL(downloadToken, s.config.Server.Port)

	qrCode, _ := utils.GenerateQRCode(link)

	c.JSON(http.StatusOK, gin.H{
		"token":     downloadToken,
		"link":      link,
		"qr_code":   qrCode,
		"file_id":   result.ID,
		"file_name": result.Name,
		"file_size": result.Size,
	})
}

func (s *Server) handleGetAvailableFiles(c *gin.Context) {
	files, err := s.transferSvc.GetAvailableFiles()
	if err != nil {
		logger.Error("获取可用文件列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文件列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"files": files})
}

func (s *Server) handleGetDownloadInfo(c *gin.Context) {
	tokenStr := c.Param("token")

	info, err := s.tokenManager.ValidateToken(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效或已过期的Token"})
		return
	}

	// 优先从token中获取文件信息（支持跨重启解析）
	if info.FileName != "" && info.FileSize > 0 {
		c.JSON(http.StatusOK, gin.H{
			"file_id":   info.FileID,
			"file_name": info.FileName,
			"file_size": info.FileSize,
			"checksum":  "",
		})
		return
	}

	// 回退：从内存中获取文件信息
	fileInfo, err := s.transferSvc.GetFileInfo(info.FileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"file_id":   fileInfo.ID,
		"file_name": fileInfo.Name,
		"file_size": fileInfo.Size,
		"checksum":  fileInfo.Checksum,
	})
}

func (s *Server) handleBatchDownload(c *gin.Context) {
	var req struct {
		FileIDs  []string `json:"file_ids" binding:"required"`
		SavePath string   `json:"save_path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	progressChan := make(chan map[string]interface{}, 1)

	go func() {
		result, err := s.transferSvc.BatchDownload(req.FileIDs, req.SavePath, progressChan)
		if err != nil {
			logger.Error("批量下载失败: %v", err)
			progressChan <- map[string]interface{}{
				"status": "failed",
				"error":  err.Error(),
			}
			return
		}
		logger.Info("批量下载完成: %d/%d 文件", result.Completed, result.TotalFiles)
	}()

	status := <-progressChan
	c.JSON(http.StatusOK, status)
}

func (s *Server) handleCalculateChecksum(c *gin.Context) {
	var req struct {
		FilePath string `json:"file_path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件路径不能为空"})
		return
	}

	checksum, err := s.checksumSvc.CalculateSHA256(req.FilePath)
	if err != nil {
		logger.Error("计算校验和失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "计算校验和失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"checksum": checksum})
}

func (s *Server) handleVerifyChecksum(c *gin.Context) {
	var req struct {
		FilePath         string `json:"file_path" binding:"required"`
		ExpectedChecksum string `json:"expected_checksum" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	valid, err := s.checksumSvc.VerifyFile(req.FilePath, req.ExpectedChecksum)
	if err != nil {
		logger.Error("验证文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "验证文件失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"valid": valid})
}

func (s *Server) handleEncrypt(c *gin.Context) {
	var req struct {
		PlainText string `json:"plain_text" binding:"required"`
		Key       string `json:"key" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	cipherText, err := s.encryptionSvc.Encrypt(req.PlainText, req.Key)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"cipher_text": cipherText})
}

func (s *Server) handleDecrypt(c *gin.Context) {
	var req struct {
		CipherText string `json:"cipher_text" binding:"required"`
		Key        string `json:"key" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	plainText, err := s.encryptionSvc.Decrypt(req.CipherText, req.Key)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plain_text": plainText})
}

func (s *Server) handleGenerateKey(c *gin.Context) {
	key, err := s.encryptionSvc.GenerateKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成密钥失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"key": key})
}

func (s *Server) handleGetPerformanceStats(c *gin.Context) {
	stats := s.perfMonitor.GetStats()
	c.JSON(http.StatusOK, stats)
}

func (s *Server) handleInitPool(c *gin.Context) {
	var req struct {
		Size int `json:"size"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || req.Size <= 0 {
		req.Size = 5
	}

	if err := s.perfMonitor.InitPool(req.Size); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "初始化线程池失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "线程池初始化成功"})
}

func (s *Server) handleStopPool(c *gin.Context) {
	if err := s.perfMonitor.StopPool(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "停止线程池失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "线程池已停止"})
}

func (s *Server) handleCheckEnvironment(c *gin.Context) {
	results, err := s.envChecker.CheckAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "环境检测失败"})
		return
	}

	c.JSON(http.StatusOK, results)
}

func (s *Server) handleGetUserConfig(c *gin.Context) {
	cfg, err := s.userConfig.GetConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户配置失败"})
		return
	}

	c.JSON(http.StatusOK, cfg)
}

func (s *Server) handleSaveUserConfig(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的配置数据"})
		return
	}

	cfg := &userconfig.Config{
		Theme:    getStringField(req, "theme", "light"),
		Language: getStringField(req, "language", "zh-CN"),
		Settings: getMapField(req, "settings"),
	}

	if err := s.userConfig.SaveConfig(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存配置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "配置保存成功"})
}

func (s *Server) handleResetUserConfig(c *gin.Context) {
	if err := s.userConfig.ResetConfig(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "重置配置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "配置已重置为默认值"})
}

func (s *Server) handleGetHistory(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 10
	}

	records, err := s.storage.GetHistory(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取历史记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"records": records})
}

func (s *Server) handleClearHistory(c *gin.Context) {
	if err := s.storage.ClearHistory(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "清除历史记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "历史记录已清除"})
}

func (s *Server) handleDownload(c *gin.Context) {
	tokenStr := c.Param("token")

	info, err := s.tokenManager.ValidateToken(tokenStr)
	if err != nil {
		c.String(http.StatusUnauthorized, "无效或已过期的下载链接")
		return
	}

	clientIP := c.ClientIP()
	if !s.accessCtrl.AllowAccess(clientIP) {
		c.String(http.StatusForbidden, "访问被拒绝")
		return
	}

	var filePath, fileName string
	var fileSize int64

	// 优先从token中获取文件信息（支持跨重启解析）
	if info.FilePath != "" && info.FileName != "" {
		filePath = info.FilePath
		fileName = info.FileName
		fileSize = info.FileSize
	} else {
		// 回退：从内存中获取文件信息
		fileInfo, err := s.transferSvc.GetFileInfo(info.FileID)
		if err != nil {
			c.String(http.StatusNotFound, "文件不存在")
			return
		}
		filePath = fileInfo.Path
		fileName = fileInfo.Name
		fileSize = fileInfo.Size
	}

	// 检查是否是文件夹
	stat, err := os.Stat(filePath)
	if err != nil {
		c.String(http.StatusNotFound, "文件不存在")
		return
	}

	if stat.IsDir() {
		// 文件夹打包为 zip 下载
		zipName := filepath.Base(filePath) + ".zip"
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Disposition", "attachment; filename="+zipName)
		c.Header("Content-Type", "application/zip")
		// 禁用缓存，确保流式传输
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")

		// 使用简化的 zip 打包
		if err := createZipFromDir(c.Writer, filePath); err != nil {
			logger.Error("打包文件夹失败: %v", err)
			return
		}
	} else {
		// 普通文件下载
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Disposition", "attachment; filename="+fileName)
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Content-Length", strconv.FormatInt(fileSize, 10))

		// 使用自定义复制以记录统计
		file, err := os.Open(filePath)
		if err != nil {
			c.String(http.StatusInternalServerError, "无法打开文件")
			return
		}
		defer file.Close()

		buffer := make([]byte, 64*1024) // 64KB 缓冲区
		for {
			n, err := file.Read(buffer)
			if n > 0 {
				c.Writer.Write(buffer[:n])
				stats.RecordSend(int64(n))
				stats.RecordDiskRead(int64(n))
			}
			if err != nil {
				break
			}
		}
	}

	// 增加下载次数
	s.storage.IncrementDownloadCount(filePath)

	record := &storage.HistoryRecord{
		FileName: fileName,
		FileSize: fileSize,
		FilePath: filePath,
		Action:   "download",
		Status:   "completed",
	}
	s.storage.AddHistory(record)
}

// createZipFromDir 将文件夹打包为 zip 并写入 writer
func createZipFromDir(w io.Writer, dirPath string) error {
	zw := zip.NewWriter(w)
	defer zw.Close()

	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录本身，只添加文件
		if info.IsDir() {
			return nil
		}

		// 计算相对路径
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}
		// 统一使用正斜杠
		relPath = strings.ReplaceAll(relPath, "\\", "/")

		// 创建 zip 文件条目（使用最简单的 Create 方法）
		wf, err := zw.Create(relPath)
		if err != nil {
			return fmt.Errorf("创建 zip 条目失败: %w", err)
		}

		// 打开源文件
		rf, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("打开文件失败: %w", err)
		}

		// 复制内容
		_, err = io.Copy(wf, rf)
		rf.Close()
		if err != nil {
			return fmt.Errorf("复制文件内容失败: %w", err)
		}

		return nil
	})
}

// handleBatchDownloadLink 处理批量下载链接（将多个文件打包成zip）
func (s *Server) handleBatchDownloadLink(c *gin.Context) {
	tokenStr := c.Param("token")

	// 验证token
	info, err := s.tokenManager.ValidateToken(tokenStr)
	if err != nil {
		c.String(http.StatusUnauthorized, "无效或已过期的下载链接")
		return
	}

	// 检查访问权限
	clientIP := c.ClientIP()
	if !s.accessCtrl.AllowAccess(clientIP) {
		c.String(http.StatusForbidden, "访问被拒绝")
		return
	}

	// 获取文件ID列表
	// 批量token的FileID格式为 "batch:fileID1,fileID2,..."
	var fileIDs []string
	if strings.HasPrefix(info.FileID, "batch:") {
		// 解析批量文件ID
		idsStr := strings.TrimPrefix(info.FileID, "batch:")
		fileIDs = strings.Split(idsStr, ",")
	} else {
		// 单个文件
		fileIDs = []string{info.FileID}
	}

	// 设置响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename=batch_download.zip")
	c.Header("Content-Type", "application/zip")
	// 禁用缓存，确保流式传输
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	// 创建zip写入器
	zw := zip.NewWriter(c.Writer)
	defer zw.Close()

	// 遍历所有文件并添加到zip
	for _, fileID := range fileIDs {
		fileInfo, err := s.transferSvc.GetFileInfo(fileID)
		if err != nil {
			logger.Error("获取文件信息失败: %v", err)
			continue
		}

		filePath := fileInfo.Path
		stat, err := os.Stat(filePath)
		if err != nil {
			logger.Error("文件不存在: %s", filePath)
			continue
		}

		if stat.IsDir() {
			// 文件夹：递归添加所有文件
			err = addDirToZip(zw, filePath, fileInfo.Name)
			if err != nil {
				logger.Error("打包文件夹失败: %v", err)
			}
		} else {
			// 单个文件
			err = addFileToZip(zw, filePath, fileInfo.Name)
			if err != nil {
				logger.Error("添加文件到zip失败: %v", err)
			}
		}

		// 记录历史
		record := &storage.HistoryRecord{
			FileName: fileInfo.Name,
			FileSize: fileInfo.Size,
			FilePath: fileInfo.Path,
			Action:   "download",
			Status:   "completed",
		}
		s.storage.AddHistory(record)
	}
}

// addFileToZip 添加单个文件到zip
func addFileToZip(zw *zip.Writer, filePath, zipName string) error {
	// 创建zip文件条目（使用最简单的 Create 方法）
	w, err := zw.Create(zipName)
	if err != nil {
		return fmt.Errorf("创建zip条目失败: %w", err)
	}

	// 打开文件
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()

	// 复制文件内容
	_, err = io.Copy(w, f)
	if err != nil {
		return fmt.Errorf("复制文件内容失败: %w", err)
	}
	return nil
}

// addDirToZip 添加文件夹内容到zip（不包含文件夹本身，只包含其内容）
func addDirToZip(zw *zip.Writer, dirPath, baseName string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录本身
		if path == dirPath {
			return nil
		}

		// 跳过子目录，只处理文件
		if info.IsDir() {
			return nil
		}

		// 计算相对路径
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		// 构建zip路径：baseName/相对路径
		zipPath := baseName + "/" + strings.ReplaceAll(relPath, "\\", "/")

		// 添加文件
		return addFileToZip(zw, path, zipPath)
	})
}

func generateDownloadURL(token string, port int) string {
	ip, _ := utils.GetLocalIP()
	return fmt.Sprintf("http://%s:%d/download/%s", ip, port, token)
}

func getStringField(m map[string]interface{}, key, defaultValue string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultValue
}

func getMapField(m map[string]interface{}, key string) map[string]interface{} {
	if v, ok := m[key]; ok {
		if mm, ok := v.(map[string]interface{}); ok {
			return mm
		}
	}
	return make(map[string]interface{})
}
