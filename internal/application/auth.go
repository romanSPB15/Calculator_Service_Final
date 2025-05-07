package application

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/romanSPB15/Calculator_Service_Final/pckg/consts/errors"
	"github.com/romanSPB15/Calculator_Service_Final/pckg/types"
)

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
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()
	req := new(types.RegisterLoginRequest)
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, errors.InvalidBody, http.StatusUnprocessableEntity)
		return
	}
	if len(req.Password) < 5 {
		http.Error(w, errors.ShortPassword, http.StatusUnprocessableEntity)
		return
	}
	_, ok := app.GetUser(req.Login, req.Password)
	if ok {
		http.Error(w, errors.UserAlreadyExists, http.StatusUnprocessableEntity)
		return
	}
	app.AddUser(req.Login, req.Password)
	fmt.Fprint(w, MakeToken(req.Login, req.Password))
}

func (app *Application) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()
	req := new(types.RegisterLoginRequest)
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, errors.InvalidBody, http.StatusUnprocessableEntity)
		return
	}
	_, ok := app.GetUser(req.Login, req.Password)
	if !ok {
		http.Error(w, errors.UserNotFound, http.StatusUnprocessableEntity)
		return
	}
	fmt.Fprint(w, MakeToken(req.Login, req.Password))
}
