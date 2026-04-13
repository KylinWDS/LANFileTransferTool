package server

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	"lanfiletransfertool/internal/access"
	"lanfiletransfertool/internal/checksum"
	"lanfiletransfertool/internal/config"
	"lanfiletransfertool/internal/encryption"
	"lanfiletransfertool/internal/environment"
	"lanfiletransfertool/internal/performance"
	"lanfiletransfertool/internal/resume"
	"lanfiletransfertool/internal/storage"
	"lanfiletransfertool/internal/token"
	"lanfiletransfertool/internal/transfer"
	"lanfiletransfertool/internal/userconfig"
	"lanfiletransfertool/pkg/logger"

	"github.com/gin-gonic/gin"
)

type Server struct {
	config        *config.Config
	httpServer    *http.Server
	router        *gin.Engine
	storage       *storage.Storage
	tokenManager  *token.Manager
	accessCtrl    *access.Control
	transferSvc   *transfer.Service
	resumeSvc     *resume.Service
	encryptionSvc *encryption.Service
	checksumSvc   *checksum.Service
	perfMonitor   *performance.Monitor
	envChecker    *environment.Checker
	userConfig    *userconfig.Manager
	mu            sync.RWMutex
	running       bool
}

func NewServer(
	cfg *config.Config,
	store *storage.Storage,
	tm *token.Manager,
	ac *access.Control,
	ts *transfer.Service,
	rs *resume.Service,
	es *encryption.Service,
	cs *checksum.Service,
	pm *performance.Monitor,
	ec *environment.Checker,
	uc *userconfig.Manager,
) *Server {

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	s := &Server{
		config:        cfg,
		router:        router,
		storage:       store,
		tokenManager:  tm,
		accessCtrl:    ac,
		transferSvc:   ts,
		resumeSvc:     rs,
		encryptionSvc: es,
		checksumSvc:   cs,
		perfMonitor:   pm,
		envChecker:    ec,
		userConfig:    uc,
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	api := s.router.Group("/api")
	{
		files := api.Group("/files")
		{
			files.POST("/select", s.handleSelectFiles)
			files.POST("/generate", s.handleGenerateLink)
			files.GET("/available", s.handleGetAvailableFiles)
		}

		download := api.Group("/download")
		{
			download.GET("/info/:token", s.handleGetDownloadInfo)
			download.POST("/batch", s.handleBatchDownload)
		}

		api.POST("/checksum/calculate", s.handleCalculateChecksum)
		api.POST("/checksum/verify", s.handleVerifyChecksum)

		configRoutes := api.Group("/user/config")
		{
			configRoutes.GET("", s.handleGetUserConfig)
			configRoutes.POST("", s.handleSaveUserConfig)
			configRoutes.POST("/reset", s.handleResetUserConfig)
		}

		perf := api.Group("/performance")
		{
			perf.GET("/stats", s.handleGetPerformanceStats)
			perf.POST("/pool/init", s.handleInitPool)
			perf.POST("/pool/stop", s.handleStopPool)
		}

		enc := api.Group("/encryption")
		{
			enc.POST("/encrypt", s.handleEncrypt)
			enc.POST("/decrypt", s.handleDecrypt)
			enc.POST("/key", s.handleGenerateKey)
		}

		api.GET("/environment/check", s.handleCheckEnvironment)

		history := api.Group("/history")
		{
			history.GET("", s.handleGetHistory)
			history.DELETE("", s.handleClearHistory)
		}
	}

	s.router.GET("/download/:token", s.handleDownload)
	s.router.GET("/download/batch/:token", s.handleBatchDownloadLink)
}

func (s *Server) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("服务器已在运行中")
	}

	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	s.running = true
	s.mu.Unlock()

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
		return fmt.Errorf("监听端口失败: %w", err)
	}

	logger.Info("HTTP服务器启动成功: %s", addr)

	go func() {
		if err := s.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			logger.Error("服务器错误: %v", err)
		}
	}()

	return nil
}

func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running || s.httpServer == nil {
		return nil
	}

	if err := s.httpServer.Close(); err != nil {
		logger.Error("停止服务器失败: %v", err)
		return err
	}

	s.running = false
	logger.Info("HTTP服务器已停止")
	return nil
}

func (s *Server) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}
