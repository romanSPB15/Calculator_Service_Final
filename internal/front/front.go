package front

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/romanSPB15/Calculator_Service_Final/pckg/consts/errors"
	"github.com/romanSPB15/Calculator_Service_Final/pckg/types"
)

var addr = "localhost:8080"

func SetAddr(new string) {
	addr = new
}

var tmpl map[string]*template.Template

var logger *log.Logger

func addFile(path string, name string) error {
	f, err := template.ParseFiles(path)
	if err != nil {
		return err
	}
	tmpl[name] = f
	return nil
}

func walk(root string) error {
	return filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		if info.IsDir() {
			return walk(path)
		}
		return addFile(path, info.Name())
	})
}

var client http.Client

func templateFile(name string) string {
	return fmt.Sprintf(`internal\front\templates\%s`, name)
}

func writeError(w http.ResponseWriter, text string, statusCode int, token string) {
	w.WriteHeader(statusCode)
	if token == "" {
		tmpl["errorWithoutAccount.html"].Execute(w, text)
	} else {
		tmpl["errorWithAccount.html"].Execute(w, text)
	}
}

func executeTemplate(name string, wr io.Writer, data any) {
	t, ok := tmpl[name]
	if !ok {
		logger.Println("[ERROR]:", name, "not found")
		return
	}
	t.Execute(wr, data)
}

var ruErrors = map[string]string{
	errors.ShortPassword:       "Короткий пароль.",
	errors.InternalServerError: "Ошибка сервера",
	errors.UserAlreadyExists:   "Пользователь уже существует",
	errors.UserNotFound:        "Пользователь не существует",
}

func ruError(err string) string {
	ru, ok := ruErrors[err]
	if !ok {
		return "Неизвестная ошибка\n" + err
	}
	return ru
}

func register(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		executeTemplate("registerForm.html", w, nil)
	case http.MethodPost:
		login := r.FormValue("login")
		password := r.FormValue("password")
		resp, err := http.Post("http://"+addr+"/api/v1/register", "application/json", strings.NewReader(fmt.Sprintf(`{"login": "%s", "password": "%s"}`, login, password)))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if resp.StatusCode != http.StatusCreated {
			if resp.StatusCode == http.StatusUnauthorized {
				writeError(w, "Необходимо войти в систему", resp.StatusCode, "")
				return
			}
			writeError(w, ruError(errors.ResponseError(resp)), resp.StatusCode, "")
			return
		}
		var respStruct types.LoginHandlerResponse
		json.NewDecoder(resp.Body).Decode(&respStruct)
		resp.Body.Close()
		executeTemplate("registerOk.html", w, respStruct.AccessToken)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		executeTemplate("loginForm.html", w, nil)
	case http.MethodPost:
		login := r.FormValue("login")
		password := r.FormValue("password")
		resp, err := http.Post("http://"+addr+"/api/v1/login", "application/json", strings.NewReader(fmt.Sprintf(`{"login": "%s", "password": "%s"}`, login, password)))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if resp.StatusCode != http.StatusOK {
			cookie := &http.Cookie{
				Name:     "token",
				SameSite: http.SameSiteStrictMode,

				MaxAge: -1,
			}
			http.SetCookie(w, cookie)
			http.Redirect(w, r, "http://"+addr+"/api/v1/web/register", http.StatusSeeOther)
			return
		}
		respStruct := new(types.LoginHandlerResponse)
		json.NewDecoder(resp.Body).Decode(respStruct)
		resp.Body.Close()
		cookie := &http.Cookie{
			Name:     "token",
			Value:    respStruct.AccessToken,
			SameSite: http.SameSiteStrictMode,

			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		}
		http.SetCookie(w, cookie)
		executeTemplate("loginOk.html", w, respStruct.AccessToken)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("token")
	text := `<p><a href="/api/v1/web/account">Мой Аккаунт</a></p>`
	if err != nil {
		text = `
		<p><a href="/api/v1/web/register">Зарегистрироваться</a></p>
        <p><a href="/api/v1/web/login">Войти</a></p>
		`
	}
	executeTemplate("index.html", w, text)
}

func calculate(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, templateFile("work/calc.html"))
}

func showID(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		http.Redirect(w, r, "http://"+addr+"/web/api/v1/login", http.StatusSeeOther)
		return
	}
	reqStruct := struct {
		Expression string `json:"expression"`
	}{r.FormValue("expression")}
	code, err := json.Marshal(reqStruct)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest(http.MethodPost, "http://"+addr+"/api/v1/calculate", bytes.NewReader(code))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Authorization", "Bearer "+cookie.Value)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		cookie := &http.Cookie{
			Name:     "token",
			SameSite: http.SameSiteStrictMode,

			MaxAge: -1,
		}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "http://"+addr+"/api/v1/web/login", http.StatusSeeOther)
		return
	}
	var m map[string]uint32
	err = json.NewDecoder(resp.Body).Decode(&m)
	if err != nil {
		panic(err)
	}
	data := fmt.Sprintf("ID=%v", m["id"])
	executeTemplate("showid.html", w, data)
}

func expressions(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		http.Redirect(w, r, "http://"+addr+"/api/v1/web/login", http.StatusSeeOther)
		return
	}
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/api/v1/expressions", nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", "Bearer "+cookie.Value)
	resp, err := client.Do(req) // Делаем запрос
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != http.StatusOK {
		cookie := &http.Cookie{
			Name:     "token",
			SameSite: http.SameSiteStrictMode,

			MaxAge: -1,
		}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "http://"+addr+"/api/v1/web/login", http.StatusSeeOther)
		return
	}
	var res types.GetExpressionsHandlerResponse
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&res) // Декодируем тело ответа
	executeTemplate("expressions.html", w, res.Expressions)
}

func expression(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, templateFile("work/expression.html"))
}

func showExpression(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		http.Redirect(w, r, "http://"+addr+"/api/v1/web/login", http.StatusSeeOther)
		return
	}
	url := fmt.Sprintf("http://%s/api/v1/expressions/%s", addr, r.FormValue("id"))
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", "Bearer "+cookie.Value)
	resp, err := client.Do(req) // Делаем запрос
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != http.StatusOK {
		cookie := &http.Cookie{
			Name:     "token",
			SameSite: http.SameSiteStrictMode,

			MaxAge: -1,
		}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "http://"+addr+"/api/v1/web/login", http.StatusSeeOther)
		return
	}
	var res types.GetExpressionHandlerResponse
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&res) // Декодируем тело ответа
	if err != nil {
		panic(err)
	}
	var resExpr struct {
		Status types.Status
		Data   string
		Result string
	}
	resExpr.Data = "Data: " + res.Expression.Data
	resExpr.Status = "Status: " + res.Expression.Status
	resExpr.Result = fmt.Sprintf("Result: %F", res.Expression.Result)
	executeTemplate("showexpr.html", w, resExpr)
}

func account(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		http.Redirect(w, r, "http://"+addr+"/api/v1/web/login", http.StatusSeeOther)
		return
	}
	url := "http://" + addr + "/api/v1/account"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", "Bearer "+cookie.Value)
	resp, err := client.Do(req) // Делаем запрос
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != http.StatusOK {
		cookie := &http.Cookie{
			Name:     "token",
			SameSite: http.SameSiteStrictMode,

			MaxAge: -1,
		}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "http://"+addr+"/api/v1/web/login", http.StatusSeeOther)
		return
	}
	respStruct := new(types.AccountHandlerResponse)
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(respStruct)
	executeTemplate("account.html", w, respStruct)
}

func Handle(mux *http.ServeMux) {
	logger = log.New(os.Stdout, "front ", log.LstdFlags|log.Lshortfile)
	tmpl = make(map[string]*template.Template)
	if err := walk(`internal\front\templates`); err != nil {
		logger.Fatal(err)
	}

	mux.HandleFunc("/api/v1/web", index)
	mux.HandleFunc("/api/v1/web/account/", account)
	mux.HandleFunc("/api/v1/web/account/calculate", calculate)
	mux.HandleFunc("/api/v1/web/register", register)
	mux.HandleFunc("/api/v1/web/login", login)
	mux.HandleFunc("/api/v1/web/account/expressions", expressions)
	mux.HandleFunc("/api/v1/web/account/showid", showID)
	mux.HandleFunc("/api/v1/web/account/expression", expression)
	mux.HandleFunc("/api/v1/web/account/showexpr", showExpression)
}
