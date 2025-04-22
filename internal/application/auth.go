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
	Login, Password string
	Expressions     map[IDExpression]*Expression
}
type AuthRequest struct {
	Password string `json:"password"`
	Login    string `json:"login"`
}

func init() {
	SECRETKEY, _ = os.LookupEnv("SECRETKEY")
}

var SECRETKEY = "romanSPB15" // по умолчанию

// Получить, есть ли такой пользователь
func (s *GRPCServer) GetUserByToken(req map[string]string) (*User, string) {
	str, has := req["token"]
	if !has {
		return nil, "token not found in url"
	}
	token, err := jwt.Parse(str, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid token")
		}

		return []byte(SECRETKEY), nil
	})
	if err != nil {
		return nil, err.Error()
	}
	login := token.Claims.(jwt.MapClaims)["login"]
	password := token.Claims.(jwt.MapClaims)["password"]
	u, ok := s.app.GetUser(login.(string), password.(string))
	if !ok {
		return nil, "user not found"
	}
	return u, ""
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
	app.logger.Printf("%s was registered", req.Login)
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
