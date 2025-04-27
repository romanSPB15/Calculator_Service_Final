package application

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/romanSPB15/Calculator_Service_Final/pckg/rpn"
)

// Получить, есть ли такой пользователь
func (a *Application) GetUserByRequest(req *http.Request) (*User, string) {
	str := req.URL.Query().Get("token")
	if str == "" {
		return nil, "token not found in request body"
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
	u, ok := a.GetUser(login.(string), password.(string))
	if !ok {
		return nil, "user not found"
	}
	return u, ""
}

// Добавление выражения через http://localhost:8080/api/v1/calculate POST.
// Тело: {"expression": "<выражение>"}
func (a *Application) AddExpressionHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	var req map[string]string
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	u, str := a.GetUserByRequest(r)
	if str != "" {
		http.Error(w, str, http.StatusUnprocessableEntity)
		return
	}
	id := uuid.New().ID()
	str, has := req["expression"]
	if !has {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	e := Expression{str, WaitStatus, 0}
	u.Expressions[id] = &e
	go func() {
		e.Status = CalculationStatus
		res, err := rpn.Calc(str, a.Tasks, a.Config.Debug, a.logger)
		if err != nil {
			e.Status = ErrorStatus
		} else {
			e.Status = "OK"
			e.Result = res
		}
	}()
	data, err := json.Marshal(AddHandlerResult{id})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}

// Получение выражения через http://localhost:8080/api/v1/expression/id GET.
func (a *Application) GetExpressionHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	vars := mux.Vars(r)
	strid := vars["id"]
	i, err := strconv.Atoi(strid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	u, str := a.GetUserByRequest(r)
	if str != "" {
		http.Error(w, str, http.StatusUnprocessableEntity)
		return
	}
	id := IDExpression(i)
	exp, has := u.Expressions[id]
	if !has {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	data, err := json.Marshal(GetExpressionHandlerResult{ExpressionWithID{id, *exp}})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func (a *Application) GetExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	u, str := a.GetUserByRequest(r)
	if str != "" {
		http.Error(w, str, http.StatusUnprocessableEntity)
		return
	}
	var ExpressionsID []ExpressionWithID
	for id, e := range u.Expressions {
		ExpressionsID = append(ExpressionsID, ExpressionWithID{id, *e})
	}
	data, err := json.Marshal(GetExpressionsHandlerResult{ExpressionsID})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}
