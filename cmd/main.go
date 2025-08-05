package main

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"idm/inner/common"
	database2 "idm/inner/database"
	"idm/inner/employee"
	"idm/inner/info"
	"idm/inner/role"
	"idm/inner/validator"
	"idm/inner/web"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	var cfg = common.GetConfig(".env")
	db := database2.ConnectDbWithCfg(cfg)
	defer func() {
		if err := db.Close(); err != nil {
			fmt.Printf("Error closing db: %v\n", err)
		}
	}()

	var server = build(db)
	go func() {
		var err = server.App.Listen(":8080")
		if err != nil {
			panic(fmt.Sprintf("http server error: %s", err))
		}
	}()
	var wg = &sync.WaitGroup{}
	wg.Add(1)
	go gracefulShutdown(server, wg)
	wg.Wait()
	fmt.Println("Graceful shutdown complete.")
}

func gracefulShutdown(server *web.Server, wg *sync.WaitGroup) {
	const timeOut = 5 * time.Second
	defer wg.Done()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	defer stop()
	<-ctx.Done()
	fmt.Println("Shutting down gracefully, press Ctrl+C again to force")
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	if err := server.App.ShutdownWithContext(ctx); err != nil {
		fmt.Printf("Server forced to shutdown with error: %v\n", err)
	}
	fmt.Println("Server exiting")
}

func build(database *sqlx.DB) *web.Server {
	var cfg = common.GetConfig(".env")
	var server = web.NewServer()
	var employeeRepo = employee.NewEmployeeRepository(database)
	var vld = validator.New()
	var employeeService = employee.NewService(employeeRepo, vld)
	var employeeController = employee.NewController(server, employeeService)
	employeeController.RegisterRoutes()
	var roleRepo = role.NewRepository(database)
	var roleService = role.NewService(roleRepo)
	var roleController = role.NewController(server, roleService)
	roleController.RegisterRouters()
	var infoController = info.NewController(server, cfg, database)
	infoController.RegisterRoutes()
	return server
}
