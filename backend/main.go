package main

import (
	"openbridge/backend/internal/config"
	"openbridge/backend/internal/domain/entity"
	"openbridge/backend/internal/handler"
	"openbridge/backend/internal/middleware"
	"openbridge/backend/internal/pkg/logger"
	"openbridge/backend/internal/repository"
	"openbridge/backend/internal/tool"
	"openbridge/backend/internal/usecase"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {

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

	// TODO: 创建数据库所在目录
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
	err = db.AutoMigrate(&entity.ProviderAccount{})
	if err != nil {
		logger.L().Fatal("db migrate failed", zap.Error(err))
	}

	err = db.AutoMigrate(&entity.QuotaSnapshot{})
	if err != nil {
		logger.L().Fatal("db migrate failed", zap.Error(err))
	}

	err = db.AutoMigrate(&entity.MountPoint{})
	if err != nil {
		logger.L().Fatal("db migrate failed", zap.Error(err))
	}

	err = db.AutoMigrate(&entity.DownloadTask{})
	if err != nil {
		logger.L().Fatal("db migrate failed", zap.Error(err))
	}

	quotaRepo := repository.NewQuotaRepository(db)

	providerRegistry := tool.NewRegistry()

	mountRepo := repository.NewMountRepository(db)
	providerRepo := repository.NewProviderRepository(db)
	downloadRepo := repository.NewDownloadRepository(db)

	mountUsecase := usecase.NewMountUseCase(mountRepo, providerRepo, quotaRepo, providerRegistry)
	mountHandler := handler.NewMountHandler(mountUsecase)

	providerUsecase := usecase.NewProviderUseCase(providerRepo, providerRegistry, mountRepo)
	providerHandler := handler.NewProviderHandler(providerUsecase)

	userUsecase := usecase.NewUserUseCase(&allConfig, db)
	// 创建 AdminChecker（用于验证用户是否为 OpenList 管理员）
	adminChecker := middleware.NewAdminChecker(allConfig.OpenList.BaseURL)
	// 如果配置了初始 token，设置到 AdminChecker
	if allConfig.OpenList.Token != "" {
		adminChecker.SetToken(allConfig.OpenList.Token)
	}
	userHandler := handler.NewUserHandler(userUsecase, adminChecker)

	storageUsecase := usecase.NewStorageUseCase(&allConfig)
	storageHandler := handler.NewStorageHandler(storageUsecase)

	aria2Client := tool.NewAria2Client(allConfig.Aria2.RPCURL, allConfig.Aria2.Secret)
	downloadUsecase := usecase.NewDownloadUseCase(storageUsecase, downloadRepo, aria2Client, &allConfig)
	downloadHandler := handler.NewDownloadHandler(downloadUsecase, storageUsecase)

	// Gin引擎设置
	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.AccessLog())
	r.Use(gin.Recovery())

	// 注册 Provider 相关路由
	providerGroup := r.Group("/api/v1/provider")
	{
		providerGroup.POST("", adminChecker.Middleware(), providerHandler.RegisterProvider)
		providerGroup.DELETE("", adminChecker.Middleware(), providerHandler.DeleteProvider)
		providerGroup.PUT("/", adminChecker.Middleware(), providerHandler.UpdateProvider)
		providerGroup.GET("/info", providerHandler.GetProvider)
		providerGroup.GET("/list", providerHandler.ListProvider)
	}

	// 注册 Mount 相关路由
	mountGroup := r.Group("/api/v1/mount")
	{
		mountGroup.POST("", adminChecker.Middleware(), mountHandler.CreateMount)
		mountGroup.GET("", mountHandler.ListMounts)
		mountGroup.PUT("/:id", adminChecker.Middleware(), mountHandler.UpdateMount)
		mountGroup.DELETE("/:id", adminChecker.Middleware(), mountHandler.DeleteMount)
		mountGroup.GET("/:id/quota", mountHandler.GetMountQuota)
		mountGroup.POST("/:id/quota/sync", mountHandler.SyncMountQuota)
	}

	// 注册 User 相关路由
	userGroup := r.Group("/api/v1/user")
	{
		userGroup.POST("/login", userHandler.UserLogin)
		userGroup.DELETE("/reset", adminChecker.Middleware(), userHandler.Reset)
		userGroup.GET("/info", userHandler.GetUserInfo)
	}

	// 注册 Storage 相关路由
	storageGroup := r.Group("/api/v1/storage")
	{
		storageGroup.GET("/drivers", storageHandler.GetDrivers)
		storageGroup.GET("/driverInfo", storageHandler.GetDriverInfo)
		storageGroup.GET("/files", storageHandler.GetFiles)
		storageGroup.GET("/file", storageHandler.GetFileInfo)
	}

	// 注册 Download 相关路由
	downloadGroup := r.Group("/api/v1/download")
	{
		downloadGroup.POST("/resolve", downloadHandler.ResolveDirectLink)
		downloadGroup.POST("/tasks", downloadHandler.CreateTask)
		downloadGroup.GET("/tasks/:id", downloadHandler.GetTask)
		downloadGroup.GET("/aria2-status", downloadHandler.GetAria2Status)
		downloadGroup.POST("/tasks/:id/retry", downloadHandler.RetryTask)
		downloadGroup.POST("/tasks/:id/open", downloadHandler.OpenFile)
			downloadGroup.POST("/tasks/:id/open-location", downloadHandler.OpenFileLocation)
	}

	if err := r.Run(":" + allConfig.App.Port); err != nil {
		logger.L().Fatal("http server run failed", zap.Error(err))
	}
}
