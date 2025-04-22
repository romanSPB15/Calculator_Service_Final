package application

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

type User struct {
	Login, Password string
	Expressions     map[IDExpression]*Expression
}

type AuthServer struct {
	app *Application
	http.Server
}

type AuthRequest struct {
	Password string `json:"password"`
	Login    string `json:"login"`
}

func MakeToken(login, password string) string {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login":    login,
		"password": password,
		"nbf":      now.Unix(),
		"exp":      now.Add(time.Hour * 240).Unix(),
		"iat":      now.Unix(),
	})

	tokenString, err := token.SignedString([]byte(hmacSampleSecret))
	if err != nil {
		panic(err)
	}
	return tokenString
}

func NewAuthServer(app *Application) *AuthServer {
	s := &AuthServer{app: app}
	s.Addr = "localhost:8181"
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/register", func(w http.ResponseWriter, r *http.Request) {
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
		app.logger.Printf("%s was registered", req.Login)
	})
	router.HandleFunc("/api/v1/login", func(w http.ResponseWriter, r *http.Request) {
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
			http.Error(w, "user not exists", http.StatusUnprocessableEntity)
			return
		}
		fmt.Fprint(w, MakeToken(req.Login, req.Password))
	})
	s.Handler = router
	return s
}
