// Тесты калькулятора.
// Внимание! Тест почистит базу данных для работы без ошибок
package application_test

// Тест приложения

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	app "github.com/romanSPB15/Calculator_Service_Final/internal/application"
	"github.com/romanSPB15/Calculator_Service_Final/internal/storage"
	"github.com/romanSPB15/Calculator_Service_Final/pckg/consts"
	"github.com/romanSPB15/Calculator_Service_Final/pckg/types"
)

// Тесты выполняются параллельно
var listenMutex = &sync.Mutex{}
var testClient = &http.Client{
	Timeout: time.Second * 3,
}

// Регистрация. Возвращает ошибку
func Register(login, password string) error {
	resp, err := testClient.Post("http://localhost:8080/api/v1/register", "application/json", strings.NewReader(fmt.Sprintf(`{"login": "%s", "password": "%s"}`, login, password)))
	if err != nil {
		return err
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("status: %s, error: %s", resp.Status, string(b))
	}
	return nil
}

// Вход. Возвращает токен и ошибку
func Login(login, password string) (string, error) {
	resp, err := testClient.Post("http://localhost:8080/api/v1/login", "application/json", strings.NewReader(fmt.Sprintf(`{"login": "%s", "password": "%s"}`, login, password)))
	if err != nil {
		return "", err
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status: %s, error: %s", resp.Status, string(b))
	}
	respStruct := types.LoginResponse{}
	json.Unmarshal(b, &respStruct)
	return respStruct.AccessToken, err
}

var sk string

func init() {
	godotenv.Load("../../config/.env")
	var has bool
	sk, has = os.LookupEnv("SECRETKEY")
	if !has {
		panic("SECRETKEY not found")
	}
}

// Проверить токен
func CheckToken(t, login, password string) error {
	t2, err := Login(login, password)
	if err != nil {
		return err
	}

	token, err := jwt.Parse(t, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid token")
		}
		return []byte(sk), nil
	})
	if err != nil {
		return fmt.Errorf("parse error t: %v", err)
	}
	id := token.Claims.(jwt.MapClaims)["id"].(types.UserID)

	token2, err := jwt.Parse(t2, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid token")
		}
		return []byte(sk), nil
	})
	if err != nil {
		return fmt.Errorf("parse error two: %v", err)
	}
	id2 := token2.Claims.(jwt.MapClaims)["id"].(types.UserID)

	if id != id2 {
		return fmt.Errorf("invalid login in token")
	}
	return nil
}

// Запуск приложения. Возвращает функцию для его остановки и очистки хранилища
func AppRun() (stop func(), clear func(), run chan struct{}) {
	listenMutex.Lock()
	a := app.New()
	a.Init(true)
	run = make(chan struct{}, 1)
	go func() {
		run <- struct{}{}
		a.Start()
	}()
	clear = func() {
		st, err := storage.Open(consts.TestStoragePath)
		if err != nil {
			panic(err)
		}
		defer st.Close()
		st.Clear()
	}
	stop = func() {
		a.Stop()
		listenMutex.Unlock()
	}
	return
}

// Тест регистрации и входа
func TestRegisterAndLogin(t *testing.T) {
	// Запускаем приложение
	stop, clear, run := AppRun()
	clear()
	<-run

	// Проверим регистрацию

	// Короткий пароль
	login := "my_login"
	password := "pswd"
	err := Register(login, password)
	if err == nil {
		t.Fatal("expected error, but got: nil")
	}

	// Достаточно длинный - не меньше 5 символов
	password = "good_password"
	err = Register(login, password)
	if err != nil {
		t.Fatalf("Expected no register error, but got: %v", err)
	}

	// Проверим вход

	// Неправильные данные
	_, err = Login("invalid", "invalid")
	if err == nil {
		t.Fatalf("Expected login error, but got: nil")
	}

	// Правильные данные
	token, err := Login(login, password)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	// Проверка токена
	err = CheckToken(token, login, password)
	if err != nil {
		t.Fatalf("Invalid register token: %v", err)
	}
	stop()
	clear()
}

// Вычислить выражение. Возвращает ID
func Calculate(expr, token string, t *testing.T) types.ExpressionID {
	url := "http://localhost:8080/api/v1/calculate"
	body := strings.NewReader(fmt.Sprintf(`{"expression": "%s"}`, expr))
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		t.Fatal("falied to make request: ", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := testClient.Do(req)
	if err != nil {
		t.Fatal("Falied to send request: ", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("Falied to calculate: %s: %s", resp.Status, string(b))
	}
	cResp := new(types.CalculateHandlerResponse)
	err = json.NewDecoder(resp.Body).Decode(cResp)
	if err != nil {
		t.Fatalf("Falied to calculate: invalid response body: %v", err)
	}
	return cResp.ID
}

// Вычислить выражение. Возвращает ID
func GetExpression(token string, id types.ExpressionID, t *testing.T) types.ExpressionWithID {
	url := fmt.Sprintf("http://localhost:8080/api/v1/expressions/%d", id)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatal("Falied to make request: ", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := testClient.Do(req)
	if err != nil {
		t.Fatal("Falied to send request: ", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("Falied to get expression: %s: %s", resp.Status, string(b))
	}
	cResp := new(types.GetExpressionHandlerResponse)
	err = json.NewDecoder(resp.Body).Decode(cResp)
	if err != nil {
		t.Fatalf("Falied to calculate: invalid response body: %v", err)
	}
	return cResp.Expression
}

// Тест работы
func TestWork(t *testing.T) {
	stop, clear, run := AppRun()
	clear()
	<-run
	time.Sleep(time.Millisecond * 20)

	login := "my_login"
	password := "good_password"
	err := Register(login, password)
	if err != nil {
		t.Fatalf("expected no register error, but got: %v", err)
	}
	token, err := Login(login, password)
	if err != nil {
		t.Fatalf("expected no login error, but got: %v", err)
	}

	testcases := []struct {
		Name           string
		Expr           string
		ExpectedError  bool
		ExpectedResult float64
	}{
		{
			Name:           "simple 1",
			Expr:           "2.1+2.9",
			ExpectedResult: 5,
		},
		{
			Name:           "simple 2",
			Expr:           "3/2*100",
			ExpectedResult: 150,
		},
		{
			Name:           "simple 3",
			Expr:           "1.2345*10000",
			ExpectedResult: 12345,
		},
		{
			Name:           "long 1",
			Expr:           "5/2.5*100,1-20", // для проверки и с точкой и с запятой
			ExpectedResult: 180.2,
		},
		{
			Name:           "long 2",
			Expr:           "5/2*(100*2)",
			ExpectedResult: 500,
		},
		{
			Name:           "long 3",
			Expr:           "10-9+(130-8)",
			ExpectedResult: 123,
		},
		{
			Name:          "double sign",
			Expr:          "5//2",
			ExpectedError: true,
		},
		{
			Name:          "one bracket",
			Expr:          "10-2*(8-4",
			ExpectedError: true,
		},
		{
			Name:          "first sign",
			Expr:          "/10*2",
			ExpectedError: true,
		},
		{
			Name:          "end sign",
			Expr:          "2+2-",
			ExpectedError: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			id := Calculate(tc.Expr, token, t)
			var ready types.ExpressionWithID
			start := time.Now()
			deadline := start.Add(time.Second)
			for {
				if time.Now().After(deadline) {
					t.Fatal("Calculate deadline exited")
				}
				expr := GetExpression(token, id, t)
				if expr.Status == consts.OKStatus || expr.Status == consts.ErrorStatus {
					ready = expr
					break
				}
			}

			if ready.Status == consts.ErrorStatus { // Ошибка
				if !tc.ExpectedError {
					t.Fatal("Expected no error, but got ErrorStatus")
				}
				return
			}
			if tc.ExpectedError {
				t.Fatal("Expected error, but got OK")
			}
			if ready.Data != tc.Expr {
				t.Fatalf("Invalid readed expression data: expected: %s, but got: %s", tc.Expr, ready.Data)
			}
			if ready.Result != tc.ExpectedResult {
				t.Fatalf("Invalid readed expression result: expected: %f, but got: %f", tc.ExpectedResult, ready.Result)
			}
		})
	}
	time.Sleep(time.Millisecond * 20)
	stop()
	clear()
}
