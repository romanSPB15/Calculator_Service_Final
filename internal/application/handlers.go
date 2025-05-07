package application

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/romanSPB15/Calculator_Service_Final/pckg/consts"
	"github.com/romanSPB15/Calculator_Service_Final/pckg/consts/errors"
	"github.com/romanSPB15/Calculator_Service_Final/pckg/rpn"
	"github.com/romanSPB15/Calculator_Service_Final/pckg/types"
)

// Получить, есть ли такой пользователь
func (a *Application) GetUserByRequest(req *http.Request) (*types.User, string, int) {
	str := req.Header.Get("Authorization-Bearer")

	if str == "" {
		return nil, "invalid header", http.StatusUnprocessableEntity
	}
	token, err := jwt.Parse(str, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid token")
		}
		return []byte(SECRETKEY), nil
	})
	if err == jwt.ErrTokenExpired {
		return nil, "token is expired", http.StatusUnauthorized
	}
	if err != nil {
		return nil, "invalid token", http.StatusUnprocessableEntity
	}
	login := token.Claims.(jwt.MapClaims)["login"].(string)
	password := token.Claims.(jwt.MapClaims)["password"].(string)
	u, ok := a.GetUser(login, password)
	if !ok {
		return nil, errors.UserNotFound, http.StatusUnprocessableEntity
	}
	return u, "", 200
}

// Добавление выражения через http://localhost:8080/api/v1/calculate POST.
// Тело: {"expression": "<выражение>"}
func (a *Application) AddExpressionHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req map[string]string
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	u, str, code := a.GetUserByRequest(r)
	if str != "" {
		http.Error(w, str, code)
		return
	}
	id := uuid.New().ID()
	str, has := req["expression"]
	if !has {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	e := types.Expression{
		Data:   str,
		Status: consts.WaitStatus,
		Result: 0,
	}
	u.Expressions[id] = &e // panic: invalid memory address or nil pointer dereference
	a.calcExpr(u.ID, id)
	data, err := json.Marshal(types.CalculateHandlerResponse{
		ID: id,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}

func (a *Application) calcExpr(userID types.UserID, id types.ExpressionID) {
	u, _ := a.GetUserByID(userID)
	e := u.Expressions[id]
	go func() {
		e.Status = consts.CalculationStatus
		res, err := rpn.Calc(e.Data, a.Tasks, a.env)
		if err != nil {
			e.Status = consts.ErrorStatus
		} else {
			e.Status = "OK"
			e.Result = res
		}
	}()
}

// Получение выражения через http://localhost:8080/api/v1/expression/:id GET.
func (a *Application) GetExpressionHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	strid := r.PathValue("id")
	i, err := strconv.Atoi(strid)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	u, str, code := a.GetUserByRequest(r)
	if str != "" {
		http.Error(w, str, code)
		return
	}
	id := types.ExpressionID(i)
	exp, has := u.Expressions[id]
	if !has {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	data, err := json.Marshal(types.GetExpressionHandlerResult{
		Expression: types.ExpressionWithID{
			ID:         id,
			Expression: *exp,
		},
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func (a *Application) GetExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	u, str, code := a.GetUserByRequest(r)
	if str != "" {
		http.Error(w, str, code)
		return
	}
	var expressionsWithID []types.ExpressionWithID
	for id, e := range u.Expressions {
		expressionsWithID = append(expressionsWithID, types.ExpressionWithID{
			ID:         id,
			Expression: *e,
		})
	}
	data, err := json.Marshal(types.GetExpressionsHandlerResult{
		Expressions: expressionsWithID,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}
