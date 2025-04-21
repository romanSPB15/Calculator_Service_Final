package application

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/romanSPB15/Calculator_Service/pckg/rpn"
)

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
	id := uuid.New().ID()
	str, has := req["expression"]
	if !has {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	e := Expression{str, WaitStatus, "", 0}
	Expressions[id] = &e
	go func() {
		e.Status = CalculationStatus
		res, err := rpn.Calc(str, Tasks, a.Config.Debug)
		if err != nil {
			e.Error = err.Error()
			e.Status = ErrorStatus
		} else {
			e.Status = "OK"
			e.Error = ""
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
	id := IDExpression(i)
	exp, has := Expressions[id]
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
	var ExpressionsID []ExpressionWithID
	for id, e := range Expressions {
		ExpressionsID = append(ExpressionsID, ExpressionWithID{id, *e})
	}
	data, err := json.Marshal(GetExpressionsHandlerResult{ExpressionsID})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func (a *Application) TaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		defer r.Body.Close()
		var tid rpn.TaskID
		for id, t := range *Tasks.Map() {
			if (*t).Status == WaitStatus {
				tid = rpn.TaskID{Task: *t, ID: id}
				break
			}
		}
		if tid.ID == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		b, err := json.Marshal(GetTaskHandlerResult{tid})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		(*Tasks.Get(tid.ID)).Status = CalculationStatus
		w.Write(b)
	case http.MethodPost:
		b, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		var r AgentResult
		err = json.Unmarshal(b, &r)
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		t := Tasks.Get(r.ID)
		t.Result = r.Result
		t.Done <- struct{}{}
		t.Status = "OK"
		if a.Config.Debug {
			log.Printf("Result Task %d(%.2F) is handle in TasksMap", r.ID, t.Result)
		}
	}
}
