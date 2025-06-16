package tests

import (
	"idm/inner/common"
	"os"
	"testing"
)

// В проекте нет .env  файла (должны получить конфигурацию из пременных окружения)
func TestGetConfigWhenNotFileThenGetVariablesEnvironment(t *testing.T) {
	os.Setenv("DB_DRIVER_NAME", "postgres")
	os.Setenv("DB_DSN", "host=127.0.0.1 port=5432")

	defer os.Unsetenv("DB_DRIVER_NAME")
	defer os.Unsetenv("DB_DSN")
	rls := common.GetConfig(os.Getenv(""))
	t.Run("Should read from variable environment", func(t *testing.T) {
		if rls.DbDriverName != "postgres" {
			t.Errorf("Result DbDriverName should be postgress, but got %s.", rls.DbDriverName)
		}
		if rls.Dsn != "host=127.0.0.1 port=5432" {
			t.Errorf("Result Dsn should be host=127.0.0.1 port=5432, but got %s.", rls.Dsn)
		}
	})
}

// В проекте есть .env  файл, но в нём нет нужных переменных и в переменных окружения их тоже нет
func TestGetConfigWhenNotFileAndThenEmptyVariablesThenGetEmptyString(t *testing.T) {
	tempDir, err := os.MkdirTemp(".", "testNotArg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempDir)
	rsl := common.GetConfig(tempDir)
	t.Run("Should return empty string", func(t *testing.T) {
		if rsl.DbDriverName != "" {
			t.Error("Not empty.")
		}
	})
}

// В проекте есть .env  файл и в нём нет нужных переменных, но в переменных окружения они есть
func TestGetConfigWhenFileEmptyThenGetVariablesEnvironment(t *testing.T) {
	os.Setenv("DB_DRIVER_NAME", "postgres")
	defer os.Unsetenv("DB_DRIVER_NAME")

	tempDir, err := os.MkdirTemp(".", "testNotArg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempDir)

	rsl := common.GetConfig(tempDir)

	t.Run("Should return empty string", func(t *testing.T) {
		if rsl.DbDriverName != "postgres" {
			t.Errorf("Result DbDriverName should be empty, but got %s.", rsl.DbDriverName)
		}
	})
}

// В проекте есть корректно заполненный .env файл, в переменных окружения нет конфликтующих с ним переменных
func TestGetConfigWhenHaveCorrectFileAndVariablesEnvThenGetFile(t *testing.T) {

	os.Setenv("DB_DRIV", "oracle")
	os.Setenv("DB_D", "dsn")
	defer os.Unsetenv("DB_DRIV")
	defer os.Unsetenv("DB_D")

	tempFile, err := os.CreateTemp(".", "test.env")
	if err != nil {
		t.Fatal(err)
	}
	_, err = tempFile.WriteString("DB_DRIVER_NAME=postgres\nDB_DSN=host=127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}

	tempFile.Close()
	defer os.Remove(tempFile.Name())

	rsl := common.GetConfig(tempFile.Name())

	t.Run("Should return strings from file", func(t *testing.T) {
		if rsl.DbDriverName != "postgres" {
			t.Errorf("Result DbDriverName should be postgress, but got %s.", rsl.DbDriverName)
		}

		if rsl.Dsn != "host=127.0.0.1" {
			t.Errorf("Result Dsn should be host=127.0.0.1, but got %s.", rsl.Dsn)
		}
	})

}

// в проекте есть .env  файл и в нём есть нужные переменные, но в переменных окружения они тоже есть (с другими значениями)
// - должны получить структуру  idm.inner.common.Config, заполненную данными. Нужно проверить, какими значениями она будет заполнена
func TestGetConfigWhenHaveCorrectFileAndVariablesEnvThenGetVariable(t *testing.T) {
	os.Setenv("DB_DRIVER_NAME", "postgres")
	os.Setenv("DB_DSN", "host=127.0.0.1")
	defer os.Unsetenv("DB_DSN")
	defer os.Unsetenv("DB_DRIVER_NAME")

	tempFile, err := os.CreateTemp(".", "test.env")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString("DB_DRIVER_NAME=oracle\nDB_DSN=host=130.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	tempFile.Close()

	rsl := common.GetConfig(tempFile.Name())

	t.Run("Should return strings from Env", func(t *testing.T) {
		if rsl.DbDriverName != "postgres" {
			t.Errorf("Result DbDriverName should be postgress, but got %s.", rsl.DbDriverName)
		}
		if rsl.Dsn != "host=127.0.0.1" {
			t.Errorf("Result Dsn should be host=127.0.0.1, but got %s.", rsl.Dsn)
		}
	})
}
