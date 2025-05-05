package application

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	ID              uint32
	Login, Password string
	Expressions     map[IDExpression]*Expression
}

type AuthRequest struct {
	Password string `json:"password"`
	Login    string `json:"login"`
}

func init() {
	key, has := os.LookupEnv("SECRETKEY")
	if has {
		SECRETKEY = key
	}
}

var SECRETKEY = "romanSPB15" // romanSPB15 - по умолчанию

func MakeToken(login, password string) string {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login":    login,
		"password": password,
		"nbf":      now.Unix(),
		"exp":      now.Add(time.Minute * 15).Unix(),
		"iat":      now.Unix(),
	})

	tokenString, err := token.SignedString([]byte(SECRETKEY))
	if err != nil {
		panic(err)
	}
	return tokenString
}

func (app *Application) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusUnprocessableEntity)
		return
	}
	defer r.Body.Close()
	req := new(AuthRequest)
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, "invalid body", http.StatusUnprocessableEntity)
		return
	}
	if len(req.Password) < 5 {
		http.Error(w, "short password", http.StatusUnprocessableEntity)
		return
	}
	_, ok := app.GetUser(req.Login, req.Password)
	if ok {
		http.Error(w, "user already exists", http.StatusUnprocessableEntity)
		return
	}
	app.AddUser(req.Login, req.Password)
	fmt.Fprint(w, MakeToken(req.Login, req.Password))
}

func (app *Application) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusUnprocessableEntity)
		return
	}
	defer r.Body.Close()
	req := new(AuthRequest)
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, "invalid body", http.StatusUnprocessableEntity)
		return
	}
	_, ok := app.GetUser(req.Login, req.Password)
	if !ok {
		http.Error(w, "invalid login or password", http.StatusUnprocessableEntity)
		return
	}
	fmt.Fprint(w, MakeToken(req.Login, req.Password))
}
