package main

import (
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"openbridge/backend/internal/config"
	"openbridge/backend/internal/domain/entity"
	"openbridge/backend/internal/handler"
	"openbridge/backend/internal/middleware"
	"openbridge/backend/internal/pkg/logger"
	"openbridge/backend/internal/repository"
	"openbridge/backend/internal/tool"
	"openbridge/backend/internal/usecase"
	webui "openbridge/backend/web"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] != "server" {
		println("usage: openbridge.exe server")
		os.Exit(1)
	}

	runServer()
}

func runServer() {
	// 读取配置
	allConfig := config.ReadConfig()
	if err := logger.Init(allConfig.Log.Level, allConfig.Log.Format); err != nil {
		panic(err)
	}
	defer logger.Sync()

	logger.L().Info("service starting",
		zap.String("app", allConfig.App.Name),
		zap.String("env", allConfig.App.Env),
		zap.String("port", allConfig.App.Port),
	)

	// 创建数据库目录
	dbDir := filepath.Dir(allConfig.DB.Path)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		logger.L().Fatal("failed to create db directory", zap.Error(err), zap.String("dir", dbDir))
	}

	// 数据库连接
	db, err := gorm.Open(sqlite.Open(allConfig.DB.Path), &gorm.Config{})
	if err != nil {
		logger.L().Fatal("db connect failed", zap.Error(err), zap.String("db_path", allConfig.DB.Path))
	}

	// 自动迁移数据库表
	if err := db.AutoMigrate(&entity.ProviderAccount{}); err != nil {
		logger.L().Fatal("db migrate failed", zap.Error(err))
	}
	if err := db.AutoMigrate(&entity.QuotaSnapshot{}); err != nil {
		logger.L().Fatal("db migrate failed", zap.Error(err))
	}
	if err := db.AutoMigrate(&entity.MountPoint{}); err != nil {
		logger.L().Fatal("db migrate failed", zap.Error(err))
	}
	if err := db.AutoMigrate(&entity.DownloadTask{}); err != nil {
		logger.L().Fatal("db migrate failed", zap.Error(err))
	}
	if err := db.AutoMigrate(&entity.RcloneProfile{}); err != nil {
		logger.L().Fatal("db migrate failed", zap.Error(err))
	}

	quotaRepo := repository.NewQuotaRepository(db)
	providerRegistry := tool.NewRegistry()
	mountRepo := repository.NewMountRepository(db)
	providerRepo := repository.NewProviderRepository(db)
	downloadRepo := repository.NewDownloadRepository(db)
	rcloneProfileRepo := repository.NewRcloneProfileRepository(db)

	mountUsecase := usecase.NewMountUseCase(mountRepo, providerRepo, quotaRepo, providerRegistry, &allConfig)
	mountHandler := handler.NewMountHandler(mountUsecase)
	webDAVProxyHandler := handler.NewWebDAVProxyHandler(mountUsecase, &allConfig)

	providerUsecase := usecase.NewProviderUseCase(providerRepo, providerRegistry, mountRepo, &allConfig)
	providerHandler := handler.NewProviderHandler(providerUsecase)

	userUsecase := usecase.NewUserUseCase(&allConfig, db)
	adminChecker := middleware.NewAdminChecker(allConfig.OpenList.BaseURL)
	if allConfig.OpenList.Token != "" {
		adminChecker.SetToken(allConfig.OpenList.Token)
	}
	userHandler := handler.NewUserHandler(userUsecase, adminChecker)

	storageUsecase := usecase.NewStorageUseCase(&allConfig, userUsecase)
	storageHandler := handler.NewStorageHandler(storageUsecase)
	runtimeController := tool.NewRuntimeController()
	runtimeController.RegisterBeforeStop(storageUsecase.FlushFileTreeCache)
	systemUsecase := usecase.NewSystemUseCase(&allConfig, runtimeController)
	systemHandler := handler.NewSystemHandler(systemUsecase)

	aria2Client := tool.NewAria2Client(allConfig.Aria2.RPCURL, allConfig.Aria2.Secret)
	tool.StartAria2IfNeeded(allConfig, aria2Client)
	downloadUsecase := usecase.NewDownloadUseCase(storageUsecase, downloadRepo, aria2Client, &allConfig)
	downloadHandler := handler.NewDownloadHandler(downloadUsecase, storageUsecase)
	settingsUsecase := usecase.NewSettingsUseCase(&allConfig, aria2Client, storageUsecase)
	settingsHandler := handler.NewSettingsHandler(settingsUsecase, adminChecker)
	rcloneUsecase := usecase.NewRcloneUseCase(&allConfig, rcloneProfileRepo, mountRepo, providerRepo)
	mountUsecase.SetProfileSyncer(rcloneUsecase)
	rcloneHandler := handler.NewRcloneHandler(rcloneUsecase, adminChecker)

	// Gin 引擎设置
	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.AccessLog())
	r.Use(gin.Recovery())

	// -------------------- API 路由 --------------------
	// Provider
	providerGroup := r.Group("/api/v1/provider")
	{
		providerGroup.POST("", adminChecker.Middleware(), providerHandler.RegisterProvider)
		providerGroup.DELETE("", adminChecker.Middleware(), providerHandler.DeleteProvider)
		providerGroup.PUT("/", adminChecker.Middleware(), providerHandler.UpdateProvider)
		providerGroup.GET("/info", providerHandler.GetProvider)
		providerGroup.GET("/list", providerHandler.ListProvider)
	}

	// Mount
	mountGroup := r.Group("/api/v1/mount")
	{
		mountGroup.POST("", adminChecker.Middleware(), mountHandler.CreateMount)
		mountGroup.GET("", mountHandler.ListMounts)
		mountGroup.PUT("/:id", adminChecker.Middleware(), mountHandler.UpdateMount)
		mountGroup.DELETE("/:id", adminChecker.Middleware(), mountHandler.DeleteMount)
		mountGroup.GET("/:id/quota", mountHandler.GetMountQuota)
		mountGroup.POST("/:id/quota/sync", mountHandler.SyncMountQuota)
	}

	// User
	userGroup := r.Group("/api/v1/user")
	{
		userGroup.POST("/login", userHandler.UserLogin)
		userGroup.DELETE("/reset", adminChecker.Middleware(), userHandler.Reset)
		userGroup.GET("/info", userHandler.GetUserInfo)
		userGroup.GET("/session-status", userHandler.GetSessionStatus)
	}

	// Storage
	storageGroup := r.Group("/api/v1/storage")
	{
		storageGroup.GET("/drivers", storageHandler.GetDrivers)
		storageGroup.GET("/driverInfo", storageHandler.GetDriverInfo)
		storageGroup.GET("/files", storageHandler.GetFiles)
		storageGroup.GET("/file", storageHandler.GetFileInfo)
	}

	systemGroup := r.Group("/api/v1/system")
	{
		systemGroup.POST("/pick-path", systemHandler.PickLocalPath)
		systemGroup.GET("/metrics", systemHandler.GetSystemMetrics)
		systemGroup.POST("/restart", adminChecker.Middleware(), systemHandler.RestartApplication)
		systemGroup.POST("/exit", adminChecker.Middleware(), systemHandler.ExitApplication)
	}

	// Download
	downloadGroup := r.Group("/api/v1/download")
	{
		downloadGroup.POST("/resolve", downloadHandler.ResolveDirectLink)
		downloadGroup.GET("/direct", downloadHandler.DownloadDirect)
		downloadGroup.HEAD("/direct", downloadHandler.DownloadDirect)
		downloadGroup.GET("/folder-zip", downloadHandler.DownloadFolderZip)
		downloadGroup.POST("/tasks", downloadHandler.CreateTask)
		downloadGroup.GET("/tasks/:id", downloadHandler.GetTask)
		downloadGroup.GET("/aria2-status", downloadHandler.GetAria2Status)
		downloadGroup.POST("/tasks/:id/retry", downloadHandler.RetryTask)
		downloadGroup.POST("/tasks/:id/open", downloadHandler.OpenFile)
		downloadGroup.POST("/tasks/:id/open-location", downloadHandler.OpenFileLocation)
	}

	settingsGroup := r.Group("/api/v1/settings")
	{
		settingsGroup.GET("", settingsHandler.GetSettings)
		settingsGroup.PUT("/app", adminChecker.Middleware(), settingsHandler.UpdateApp)
		settingsGroup.PUT("/openlist", adminChecker.Middleware(), settingsHandler.UpdateOpenList)
		settingsGroup.PUT("/aria2", adminChecker.Middleware(), settingsHandler.UpdateAria2)
		settingsGroup.PUT("/rclone", adminChecker.Middleware(), settingsHandler.UpdateRclone)
		settingsGroup.PUT("/session", adminChecker.Middleware(), settingsHandler.UpdateSession)
		settingsGroup.PUT("/filetree", adminChecker.Middleware(), settingsHandler.UpdateFileTree)
	}

	rcloneGroup := r.Group("/api/v1/rclone")
	{
		rcloneGroup.GET("/profiles", rcloneHandler.ListProfiles)
		rcloneGroup.POST("/profiles", adminChecker.Middleware(), rcloneHandler.CreateProfile)
		rcloneGroup.PUT("/profiles/:id", adminChecker.Middleware(), rcloneHandler.UpdateProfile)
		rcloneGroup.DELETE("/profiles/:id", adminChecker.Middleware(), rcloneHandler.DeleteProfile)
		rcloneGroup.POST("/profiles/:id/apply", adminChecker.Middleware(), rcloneHandler.ApplyProfile)
		rcloneGroup.POST("/profiles/:id/mount", adminChecker.Middleware(), rcloneHandler.MountProfile)
		rcloneGroup.POST("/profiles/:id/unmount", adminChecker.Middleware(), rcloneHandler.UnmountProfile)
	}

	for _, method := range []string{
		http.MethodGet,
		http.MethodHead,
		http.MethodOptions,
		http.MethodPut,
		http.MethodDelete,
		"PROPFIND",
		"PROPPATCH",
		"MKCOL",
		"COPY",
		"MOVE",
		"LOCK",
		"UNLOCK",
	} {
		r.Handle(method, "/api/v1/webdav/mounts/:id", webDAVProxyHandler.ProxyMount)
		r.Handle(method, "/api/v1/webdav/mounts/:id/*filepath", webDAVProxyHandler.ProxyMount)
	}

	// -------------------- 前端静态文件 --------------------
	distFS, err := fs.Sub(webui.Dist, "dist")
	if err != nil {
		logger.L().Fatal("failed to load embedded frontend", zap.Error(err))
	}

	// 静态资源 /assets
	r.GET("/assets/*filepath", func(c *gin.Context) {
		path := strings.TrimPrefix(c.Param("filepath"), "/")
		c.FileFromFS("assets/"+path, http.FS(distFS))
	})

	// favicon
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.FileFromFS("favicon.ico", http.FS(distFS))
	})

	// 首页
	serveIndex := func(c *gin.Context) {
		f, err := distFS.Open("index.html")
		if err != nil {
			c.String(500, "index.html not found")
			return
		}
		defer f.Close()

		c.Header("Content-Type", "text/html; charset=utf-8")
		c.Status(200)
		_, _ = io.Copy(c.Writer, f)
	}

	r.GET("/", serveIndex)

	// SPA 回退
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(404, gin.H{"code": 404, "msg": "Not Found"})
			return
		}
		serveIndex(c)
	})

	// -------------------- 启动 --------------------
	addr := ":" + allConfig.App.Port
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logger.L().Fatal("http listen failed", zap.Error(err), zap.String("addr", addr))
	}

	if allConfig.App.AutoOpenBrowser {
		tool.OpenBrowserAfterStartup("http://localhost:" + allConfig.App.Port)
	}

	server := &http.Server{Handler: r}
	runtimeController.BindServer(server)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	go func() {
		<-sigCh
		runtimeController.RequestExit()
	}()

	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		logger.L().Fatal("http server run failed", zap.Error(err))
	}

	if runtimeController.Action() == tool.RuntimeActionRestart {
		if err := tool.RestartCurrentProcess(os.Args[1:]); err != nil {
			logger.L().Error("restart current process failed", zap.Error(err))
			return
		}
		logger.L().Info("restart current process requested")
	}
}
