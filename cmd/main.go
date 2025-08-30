package main

import (
	"context"
	"idm/inner/common"
	database2 "idm/inner/database"
	"idm/inner/employee"
	"idm/inner/info"
	"idm/inner/role"
	"idm/inner/web"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func main() {
	var cfg = common.GetConfig(".env")
	var logger = common.NewLogger(cfg)
	defer func() { _ = logger.Sync() }()
	db := database2.ConnectDbWithCfg(cfg)
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Error closing db", zap.Error(err))
		}
	}()

	var server = build(db, logger)
	go func() {
		var err = server.App.Listen(":8080")
		if err != nil {
			logger.Panic("http server error", zap.Error(err))
		}
	}()
	var wg = &sync.WaitGroup{}
	wg.Add(1)
	go gracefulShutdown(server, wg, logger)
	wg.Wait()
	logger.Info("Graceful shutdown complete.")
}

func gracefulShutdown(server *web.Server, wg *sync.WaitGroup, logger *common.Logger) {
	const timeOut = 5 * time.Second
	defer wg.Done()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	defer stop()
	<-ctx.Done()
	logger.Info("Shutting down gracefully")
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	if err := server.App.ShutdownWithContext(ctx); err != nil {
		logger.Error("Server forced to shutdown with error", zap.Error(err))
	}
	logger.Info("Server exiting")
}

func build(database *sqlx.DB, logger *common.Logger) *web.Server {
	var cfg = common.GetConfig(".env")
	var server = web.NewServer()
	var employeeRepo = employee.NewEmployeeRepository(database)
	var employeeService = employee.NewService(employeeRepo)
	var employeeController = employee.NewController(server, employeeService, logger)
	employeeController.RegisterRoutes()
	var roleRepo = role.NewRepository(database)
	var roleService = role.NewService(roleRepo)
	var roleController = role.NewController(server, roleService, logger)
	roleController.RegisterRouters()
	var infoController = info.NewController(server, cfg, database, logger)
	infoController.RegisterRoutes()
	return server
}
