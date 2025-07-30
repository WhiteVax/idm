package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"idm/inner/common"
	database2 "idm/inner/database"
	"idm/inner/employee"
	"idm/inner/info"
	"idm/inner/role"
	"idm/inner/validator"
	"idm/inner/web"
)

func main() {
	var cfg = common.GetConfig(".env")
	db := database2.ConnectDbWithCfg(cfg)
	defer func() {
		if err := db.Close; err != nil {
			fmt.Println("Error closing db: %v", err)
		}
	}()
	var server = build(db)
	var err = server.App.Listen(":8080")
	if err != nil {
		panic(fmt.Sprintf("http server error: %s", err))
	}
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
