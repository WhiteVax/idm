package common

import (
	"os"
	"testing"
)

// TestGetConfigWhenNotFileThenGetVariablesEnvironment - в проекте нет .env  файла (должны получить конфигурацию
// из пременных окружения)
func TestGetConfigWhenNotFileThenGetVariablesEnvironment(t *testing.T) {
	t.Setenv("DB_DRIVER_NAME", "postgres")
	t.Setenv("DB_DSN", "host=127.0.0.1 port=5432")
	t.Setenv("APP_NAME", "idm")
	t.Setenv("APP_VERSION", "0.0.0")
	t.Setenv("SSL_SERT", "sert")
	t.Setenv("SSL_KEY", "Ket")

	rls := GetConfig(os.Getenv(""))
	t.Run("Should read from variable environment", func(t *testing.T) {
		if rls.DbDriverName != "postgres" {
			t.Errorf("Result DbDriverName should be postgress, but got %s.", rls.DbDriverName)
		}
		if rls.Dsn != "host=127.0.0.1 port=5432" {
			t.Errorf("Result Dsn should be host=127.0.0.1 port=5432, but got %s.", rls.Dsn)
		}
	})
}

// TestGetConfigWhenNotFileAndThenEmptyVariablesThenGetEmptyString - в проекте есть .env  файл,
// но в нём нет нужных переменных и в переменных окружения их тоже нет
func TestGetConfigWhenNotFileAndThenEmptyVariablesThenGetEmptyString(t *testing.T) {
	tempDir, err := os.MkdirTemp(".", "testNotArg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempDir)

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic due to missing required environment variables")
		}
	}()

	GetConfig(tempDir)
}

// TestGetConfigWhenFileEmptyThenGetVariablesEnvironment - в проекте есть .env  файл и в нём нет нужных переменных,
// но в переменных окружения они есть
func TestGetConfigWhenFileEmptyThenGetVariablesEnvironment(t *testing.T) {
	t.Setenv("DB_DRIVER_NAME", "postgres")
	t.Setenv("APP_NAME", "idm")
	t.Setenv("DB_DSN", "host=127.0.0.1 port=5432")
	t.Setenv("APP_VERSION", "0.0.0")
	t.Setenv("SSL_SERT", "sert")
	t.Setenv("SSL_KEY", "Ket")
	t.Setenv("KEYCLOAK_JWK_URL", "url")

	tempDir, err := os.MkdirTemp(".", "testNotArg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempDir)
	rsl := GetConfig(tempDir)

	t.Run("Should return values from env variables", func(t *testing.T) {
		if rsl.DbDriverName != "postgres" {
			t.Errorf("DbDriverName should be 'postgres', got %s", rsl.DbDriverName)
		}
	})
}

// TestGetConfigWhenHaveCorrectFileAndVariablesEnvThenGetFile - в проекте есть корректно заполненный .env файл,
// в переменных окружения нет конфликтующих с ним переменных
func TestGetConfigWhenHaveCorrectFileAndVariablesEnvThenGetFile(t *testing.T) {
	t.Setenv("DB_DRIVER_NAME", "oracle")
	t.Setenv("APP_NAME", "idm")
	t.Setenv("DB_DSN", "host=127.0.0.1")
	t.Setenv("APP_VERSION", "0.0.0")
	t.Setenv("SSL_SERT", "sert")
	t.Setenv("SSL_KEY", "Ket")
	t.Setenv("KEYCLOAK_JWK_URL", "url")
	defer os.Unsetenv("DB_DRIV")
	defer os.Unsetenv("DB_D")

	tempFile, err := os.CreateTemp(".", "test.env")
	if err != nil {
		t.Fatal(err)
	}
	_, err = tempFile.WriteString("DB_DRIVER_NAME=postgres\nDB_DSN=host=127.0.0.1 port=5432")
	if err != nil {
		t.Fatal(err)
	}

	tempFile.Close()
	defer os.Remove(tempFile.Name())

	rsl := GetConfig(tempFile.Name())

	t.Run("Should return strings from file", func(t *testing.T) {
		if rsl.DbDriverName != "oracle" {
			t.Errorf("Result DbDriverName should be postgress, but got %s.", rsl.DbDriverName)
		}

		if rsl.Dsn != "host=127.0.0.1" {
			t.Errorf("Result Dsn should be host=127.0.0.1, but got %s.", rsl.Dsn)
		}
	})

}

// TestGetConfigWhenHaveCorrectFileAndVariablesEnvThenGetVariable - в проекте есть .env  файл и в
// нём есть нужные переменные, но в переменных окружения они тоже есть (с другими значениями) -
// должны получить структуру  idm.inner.common.Config, заполненную данными. Нужно проверить, какими значениями она будет заполнена
func TestGetConfigWhenHaveCorrectFileAndVariablesEnvThenGetVariable(t *testing.T) {
	t.Setenv("DB_DRIVER_NAME", "postgres")
	t.Setenv("DB_DSN", "host=127.0.0.1")
	t.Setenv("APP_NAME", "idm")
	t.Setenv("APP_VERSION", "0.0.0")
	t.Setenv("SSL_SERT", "sert")
	t.Setenv("SSL_KEY", "Ket")
	t.Setenv("KEYCLOAK_JWK_URL", "url")

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

	rsl := GetConfig(tempFile.Name())

	t.Run("Should return strings from Env", func(t *testing.T) {
		if rsl.DbDriverName != "postgres" {
			t.Errorf("Result DbDriverName should be postgress, but got %s.", rsl.DbDriverName)
		}
		if rsl.Dsn != "host=127.0.0.1" {
			t.Errorf("Result Dsn should be host=127.0.0.1, but got %s.", rsl.Dsn)
		}
	})
}
