package web

/*
import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/romanSPB15/Calculator_Service_Final/pckg/dir"
)

var client http.Client

type Expression struct {
	Data   string `json:"data"`
	Status string
	Result float64
}

func index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, dir.GetTemplateFile("index.html"))
}

func calculate(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, dir.GetTemplateFile("calc.html"))
}

func showID(w http.ResponseWriter, r *http.Request) {
	req := struct {
		Expression string `json:"expression"`
	}{r.FormValue("expression")}
	code, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	resp, err := client.Post("http://localhost:8080/api/v1/calculate", "application/json", bytes.NewReader(code))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var m map[string]uint32
	err = json.NewDecoder(resp.Body).Decode(&m)
	if err != nil {
		panic(err)
	}
	tmpl, err := template.ParseFiles(dir.GetTemplateFile("showid.html"))
	if err != nil {
		panic(err)
	}
	tmpl.Execute(w, fmt.Sprintf("ID=%v", m["id"]))
}

type aAPIGetExpressionsResult struct { // Ответ API при получении списка выражений
	Expressions []Expression `json:"expressions"`
}

func expressions(w http.ResponseWriter, r *http.Request) {
	resp, err := client.Get("http://localhost:8080/api/v1/expressions") // Делаем запрос
	if err != nil {
		panic(err)
	}
	var res aAPIGetExpressionsResult
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&res) // Декодируем тело ответа
	if err != nil {
		panic(err)
	}
	tmpl, err := template.ParseFiles(dir.GetTemplateFile("expressions.html"))
	if err != nil {
		panic(err)
	}
	tmpl.Execute(w, res.Expressions)
}

func expression(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, dir.GetTemplateFile("expression.html"))
}

type aAPIGetExpressionResult struct { // Ответ API при получении списка выражений
	Expression Expression `json:"expression"`
}

func showExpression(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("http://localhost:8080/api/v1/expressions/%s", r.FormValue("id"))
	resp, err := client.Get(url) // Делаем запрос
	if err != nil {
		panic(err)
	}
	if resp.StatusCode == 404 {
		http.ServeFile(w, r, dir.GetTemplateFile("notfoundexpr.html"))
		return
	}
	var res aAPIGetExpressionResult
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&res) // Декодируем тело ответа
	if err != nil {
		panic(err)
	}
	tmpl, err := template.ParseFiles(dir.GetTemplateFile("showexpr.html"))
	if err != nil {
		panic(err)
	}
	res.Expression.Status = "Status: " + res.Expression.Status
	var resExpr struct {
		Status string
		Data   string
		Result string
	}
	resExpr.Data = "Data: " + res.Expression.Data
	resExpr.Status = res.Expression.Status
	resExpr.Result = fmt.Sprintf("Result: %F", res.Expression.Result)
	err = tmpl.Execute(w, resExpr)
	if err != nil {
		panic(err)
	}
}

func Run() {
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/web", index)
	router.HandleFunc("/api/v1/web/calculate", calculate)
	router.HandleFunc("/api/v1/web/expressions", expressions)
	router.HandleFunc("/api/v1/web/showid", showID)
	router.HandleFunc("/api/v1/web/expression", expression)
	router.HandleFunc("/api/v1/web/showexpr", showExpression)
	http.Handle("/", router)
	app.loggerFatalf("falied to listening port 8181: %v", http.ListenAndServe(":8181", nil))
}*/
