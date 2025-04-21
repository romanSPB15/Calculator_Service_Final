package application

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/romanSPB15/Calculator_Service/internal/web"
	"github.com/romanSPB15/Calculator_Service/pckg/dir"
	"github.com/romanSPB15/Calculator_Service/pckg/rpn"
)

const (
	WaitStatus        = "Wait"
	OKStatus          = "OK"
	CalculationStatus = "Calculation"
	ErrorStatus       = "Error"
)

// Выражение
type Expression struct {
	Data   string  `json:"data"`
	Error  string  `json:"error"`
	Status string  `json:"status"`
	Result float64 `json:"result"`
}

// Выражение с ID
type ExpressionWithID struct {
	ID IDExpression `json:"id"`
	Expression
}

// ID выражения
type IDExpression = uint32

type GetExpressionHandlerResult struct {
	Expression ExpressionWithID `json:"expression"`
}

type AddHandlerResult struct {
	ID uint32 `json:"id"`
}

type GetExpressionsHandlerResult struct {
	Expressions []ExpressionWithID `json:"expressions"`
}

type GetTaskHandlerResult struct {
	Task rpn.TaskID `json:"task"`
}

type AgentResult struct {
	ID     rpn.IDTask `json:"id"`
	Result float64    `json:"result"`
}

// Выражения
var Expressions = make(map[IDExpression]*Expression)

// Задачи
var Tasks = rpn.NewConcurrentTaskMap()

// Приложение
type Application struct {
	// Агент
	Config       *config
	Agent        http.Client
	NumGoroutine int
	Router       *mux.Router
}

func New() *Application {
	return &Application{
		Router: mux.NewRouter(),
		Config: newConfig(),
	}
}

type runError struct {
	error
	object string
}

const IP = "localhost:8080"

// Запуск всей системы
func (app *Application) RunServer() {
	rpn.InitEnv(dir.EnvFile()) // Иницилизация переменных из среды
	/* ListenAndServe() закончится только с ошибкой, как и runAgent() */
	startServer := make(chan struct{}, 1) // Канал запуска оркестратора
	err := make(chan runError)            // Канал с ошибкой
	go func() {
		if app.Config.Debug {
			log.Println("Orkestrator Runned")
		}
		startServer <- struct{}{}
		errIntrf := http.ListenAndServe(IP, nil)
		err <- runError{
			errIntrf,
			"Orkestrator",
		}
	}()
	// Создаём новый mux.Router
	/* Инициализация обработчиков роутера */
	app.Router.HandleFunc("/api/v1/calculate", app.AddExpressionHandler)
	app.Router.HandleFunc("/api/v1/expressions/{id}", app.GetExpressionHandler)
	app.Router.HandleFunc("/api/v1/expressions", app.GetExpressionsHandler)
	app.Router.HandleFunc("/api/v1/internal/task", app.TaskHandler)
	if app.Config.Web {
		web.HandleToRouter(app.Router)
	}
	http.Handle("/", app.Router)

	go func() {
		<-startServer // Ждём, когда запустится оркестратор
		errIntrf := app.runAgent()
		err <- runError{
			errIntrf,
			"Agent",
		}
	}()
	time.Sleep(100 * time.Millisecond)
	for {
		var cmd string
		fmt.Scan(&cmd)
		switch cmd {
		case "exit":
			os.Exit(0)
		case "help":
			fmt.Print("\r\n")
			fmt.Println("Calculator_Service - это сервис для вычисления арифметических выражений.")
			fmt.Println("\tКоманды")
			fmt.Println("help - помощь")
			fmt.Println("exit - выход из приложения")
			fmt.Println("Более подробную информацию можно на найти в README репозитория: https://github.com/romanSPB15/Calculator_Service")
			fmt.Println("\t\tRomanSPB15")
		}
	}
}
