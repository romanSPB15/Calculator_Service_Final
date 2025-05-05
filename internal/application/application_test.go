package application_test

// Внимание!!! Тест почистит базу данных для работы без ошибок

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	app "github.com/romanSPB15/Calculator_Service_Final/internal/application"
)

// Чистит данные приложения - Удаляет файл
func ClearDb() {
	os.Remove(app.TestStoragePath)
}

func Register(login, password string) (string, error) {
	resp, err := http.Post("http://localhost:8080/api/v1/register", "application/json", strings.NewReader(fmt.Sprintf(`{"login": "%s", "password": "%s"}`, login, password)))
	if err != nil {
		return "", err
	}

	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status: %s, error: %s", resp.Status, string(b))
	}
	return string(b), err
}

func Login(login, password string) (string, error) {
	resp, err := http.Post("http://localhost:8080/api/v1/login", "application/json", strings.NewReader(fmt.Sprintf(`{"login": "%s", "password": "%s"}`, login, password)))
	if err != nil {
		return "", err
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status: %s, error: %s", resp.Status, string(b))
	}
	return string(b), err
}

func CheckToken(s, login, password string) error {
	token, err := jwt.Parse(s, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid token")
		}
		return []byte(app.SECRETKEY), nil
	})
	if err != nil {
		return fmt.Errorf("parse error")
	}
	login2 := token.Claims.(jwt.MapClaims)["login"].(string)
	password2 := token.Claims.(jwt.MapClaims)["password"].(string)
	if login2 != login {
		return fmt.Errorf("invalid login in token")
	}
	if password2 != password {
		return fmt.Errorf("invalid password in token")
	}
	return nil
}

func AppRun() (stop func(), clear func()) {
	a := app.New()
	go a.Run(true)
	clear = func() {
		err := a.Storage.Clear()
		if err != nil {
			panic(err)
		}
	}
	stop = a.Stop
	return
}

func TestRegisterAndLogin(t *testing.T) {
	stop, clear := AppRun()
	time.Sleep(time.Millisecond * 5)
	clear()

	// Проверим регистрацию
	// Короткий пароль!
	login := "my_login"
	password := "pswd"
	_, err := Register(login, password)
	if err == nil {
		t.Fatal("expected error, but got: nil")
	}

	// Достаточно длинный - не меньше 5 символов
	password = "good_password"
	stringToken, err := Register(login, password)
	if err != nil {
		t.Fatalf("expected no register error, but got: %v", err)
	}

	// Проверка токена
	err = CheckToken(stringToken, login, password)
	if err != nil {
		t.Fatalf("invalid register token: %v", err)
	}

	// Проверим вход
	stringToken, err = Login(login, password)
	if err != nil {
		t.Fatalf("expected no login error, but got: %v", err)
	}

	// Проверка токена
	err = CheckToken(stringToken, login, password)
	if err != nil {
		t.Fatalf("invalid login token: %v", err)
	}
	clear()
	stop()
}
