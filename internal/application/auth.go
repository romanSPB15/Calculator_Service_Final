package application

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/romanSPB15/Calculator_Service_Final/pckg/consts/errors"
	"github.com/romanSPB15/Calculator_Service_Final/pckg/types"
)

func (app *Application) MakeToken(id types.UserID) string {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  id,
		"nbf": now.Unix(),
		"exp": now.Add(time.Hour * 24).Unix(),
		"iat": now.Unix(),
	})
	tokenString, err := token.SignedString([]byte(app.env.SECRETKEY))
	if err != nil {
		panic(err)
	}
	return tokenString
}

func makeLoginResponse(token string, w http.ResponseWriter) {
	b, err := json.Marshal(types.LoginResponse{AccessToken: token})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func writeError(w http.ResponseWriter, text string, statusCode int) {
	w.WriteHeader(statusCode)
	fmt.Fprint(w, text)
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
		writeError(w, errors.InvalidBody, http.StatusUnprocessableEntity)
		return
	}
	if len(req.Password) < 5 {
		writeError(w, errors.ShortPassword, http.StatusUnprocessableEntity)
		return
	}
	_, ok := app.GetUser(req.Login, req.Password)
	if ok {
		writeError(w, errors.UserAlreadyExists, http.StatusUnprocessableEntity)
		return
	}
	_, err = app.AddUser(req.Login, req.Password)
	if err != nil {
		writeError(w, errors.InternalServerError, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
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
	u, ok := app.GetUser(req.Login, req.Password)
	if !ok {
		http.Error(w, errors.UserNotFound, http.StatusUnauthorized)
		return
	}
	makeLoginResponse(app.MakeToken(u.ID), w)
}
