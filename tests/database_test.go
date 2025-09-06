package tests

import (
	"idm/inner/common"
	database2 "idm/inner/database"
	"os"
	"testing"
)

// TestConnectionWhenConnectionSuccessful - Проверка через Пинг соединения
func TestConnectionWhenConnectionSuccessful(t *testing.T) {
	os.Setenv("DB_DRIVER_NAME", "postgres")
	os.Setenv("DB_DSN", "host=127.0.0.1 port=5440 user=postgres password=test_postgres dbname=postgres sslmode=disable")
	os.Setenv("APP_NAME", "idm")
	os.Setenv("APP_VERSION", "0.0.0")
	cfg := common.GetConfig("")
	defer os.Unsetenv("DB_DRIVER_NAME")
	defer os.Unsetenv("DB_DSN")

	database := database2.ConnectDbWithCfg(cfg)

	t.Run("Database Connection NotNil", func(t *testing.T) {
		if err := database.Ping(); err != nil {
			t.Error("Database Connection Not Present")
		}
	})
}

func TestConnectionWhenWrongPassword(t *testing.T) {

	os.Setenv("DB_DRIVER_NAME", "postgres")
	os.Setenv("DB_DSN", "host=127.0.0.1 port=5440 user=postgres password=password dbname=postgres sslmode=disable")
	defer os.Unsetenv("DB_DRIVER_NAME")
	defer os.Unsetenv("DB_DSN")

	defer func() {
		if r := recover(); r == nil {
			t.Error("Database Connection Not Present")
		}
	}()

	cfg := common.GetConfig("")
	_ = database2.ConnectDbWithCfg(cfg).Ping()
}
