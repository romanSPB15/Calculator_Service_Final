package application

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	//"github.com/romanSPB15/Calculator_Service_Final/internal/web"
	"github.com/google/uuid"
	"github.com/romanSPB15/Calculator_Service_Final/pckg/dir"
	"github.com/romanSPB15/Calculator_Service_Final/pckg/env"
	"github.com/romanSPB15/Calculator_Service_Final/pckg/rpn"
	pb "github.com/romanSPB15/Calculator_Service_Final/proto"
	"google.golang.org/grpc"
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

// Приложение
type Application struct {
	Config       *config
	NumGoroutine int
	workerId     int
	grpcServer   *grpc.Server
	calcServer   pb.CalculatorServiceServer
	Users        []*User
	Tasks        *rpn.ConcurrentTaskMap
	logger       *log.Logger
	storage      *Storage
	server       *http.Server
	grpcListener net.Listener
	env          *env.List     // Переменные среды
	agentStop    chan struct{} // Канал остановки агента
}

func New() *Application {
	app := &Application{
		Config:       newConfig(),
		NumGoroutine: 0,
		workerId:     0,
		Tasks:        rpn.NewConcurrentTaskMap(),
		grpcServer:   grpc.NewServer(),
		logger:       log.New(os.Stdout, "app: ", log.Ltime),
	}
	app.calcServer = app.NewServer()
	app.env = env.NewList()
	return app
}

func (a *Application) LoadData() error {
	usersList, err := a.storage.SelectAllUsers()
	if err != nil {
		return err
	}
	a.Users = usersList
	for _, u := range usersList {
		expressionsList, err := a.storage.SelectExpressionsForUser(u)
		if err != nil {
			return err
		}
		u.Expressions = make(map[IDExpression]*Expression)
		for _, v := range expressionsList {
			u.Expressions[v.ID] = &v.Expression
		}
	}
	return nil
}

func (a *Application) SaveData() error {
	err := a.storage.Clear()
	if err != nil {
		return err
	}
	for _, u := range a.Users {
		err = a.storage.InsertUser(u)
		if err != nil {
			return err
		}
		for id, expr := range u.Expressions {
			err = a.storage.InsertExpression(&ExpressionWithID{Expression: *expr, ID: id}, u)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *Application) GetUser(login, password string) (u *User, ok bool) {
	for _, v := range a.Users {
		if v.Password == password && v.Login == login {
			u = v
			ok = true
			return
		}
	}
	return
}

func (a *Application) AddUser(login, password string) {
	u := &User{
		Login:       login,
		Password:    password,
		Expressions: make(map[IDExpression]*Expression),
		ID:          uuid.New().ID(),
	}
	a.Users = append(a.Users, u)
}

const GRPC_PORT = 8081

func (app *Application) Init() error {
	var err error
	app.storage, err = OpenStorage(AppStoragePath)
	if err != nil {
		return err
	}

	err = app.LoadData()
	if err != nil {
		return err
	}
	pb.RegisterCalculatorServiceServer(app.grpcServer, app.calcServer)

	app.env.InitEnv(dir.EnvFile()) // Иницилизация переменных среды

	addr := fmt.Sprintf("%s:%d", app.env.HOST, app.env.PORT)

	mux := http.NewServeMux()
	// Создаём новый mux.Router
	/* Инициализация обработчиков роутера */
	mux.HandleFunc("/api/v1/register", app.RegisterHandler)
	mux.HandleFunc("/api/v1/login", app.LoginHandler)
	mux.HandleFunc("/api/v1/calculate", app.AddExpressionHandler)
	mux.HandleFunc("/api/v1/expressions/{id}", app.GetExpressionHandler)
	mux.HandleFunc("/api/v1/expressions", app.GetExpressionsHandler)
	app.server = &http.Server{Addr: addr, Handler: mux}

	addrGRPC := fmt.Sprintf("%s:%d", app.env.HOST, GRPC_PORT)
	app.grpcListener, err = net.Listen("tcp", addrGRPC) // будем ждать запросы по этому адресу
	if err != nil {
		return err
	}
	app.agentStop = make(chan struct{})
	return nil
}

// Запуск системы
func (app *Application) Start() {
	go func() {
		app.logger.Println("Agent runned")
		if err := app.runAgent(); err != nil {
			app.logger.Fatal("falied to run agent: ", err)
		}
	}()

	go func() {
		app.logger.Println("GRPC runned")
		if err := app.grpcServer.Serve(app.grpcListener); err != nil {
			app.logger.Fatal("failed to serving grpc: ", err)
		}
	}()

	app.logger.Println("Main runned")
	app.logger.Fatal("main: failed to serving: ", app.server.ListenAndServe())
}

// Запуск системы
func (app *Application) Run() {
	err := app.Init()
	if err != nil {
		app.logger.Fatal("failed to init application: ", err)
	}
	app.Start()
}

func (app *Application) Stop() {
	close(app.agentStop)
	app.grpcServer.GracefulStop()
	app.logger.Println("Gracefully stopped")
	defer app.storage.Close()
	err := app.SaveData()
	if err != nil {
		app.logger.Fatal("failed to save data: ", err)
	}
}
