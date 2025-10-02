package main

import (
	"context"
	"crypto/tls"
	"idm/docs"
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

// @title IDM API documentation
// @BasePath /api/v1/
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
	docs.SwaggerInfo.Version = cfg.AppVersion
	cer, err := tls.LoadX509KeyPair(cfg.SslSert, cfg.SslKey)
	if err != nil {
		logger.Panic("Failed certificate loading: ", zap.Error(err))
	}
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cer}}
	ln, err := tls.Listen("tcp", ":8080", tlsConfig)
	if err != nil {
		logger.Panic("Failed TLS listen", zap.Error(err))
	}
	ln = common.CustomListener{Listener: ln, Url: "127.0.0.1:8080/swagger/index.html"}
	var server = build(db, logger)
	go func() {
		var err = server.App.Listener(ln)
		if err != nil {
			logger.Panic("Failed TLS listener creating: %s", zap.Error(err))
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
	var employeeHandler = employee.NewHandler(server, employeeService, logger)
	employeeHandler.RegisterRoutes()
	var roleRepo = role.NewRepository(database)
	var roleService = role.NewService(roleRepo)
	var roleHandler = role.NewHandler(server, roleService, logger)
	roleHandler.RegisterRouters()
	var infoHandler = info.NewHandler(server, cfg, database, logger)
	infoHandler.RegisterRoutes()
	return server
}
