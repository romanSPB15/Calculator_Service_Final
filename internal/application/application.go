package application

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/google/uuid"
	"github.com/romanSPB15/Calculator_Service_Final/internal/env"
	"github.com/romanSPB15/Calculator_Service_Final/internal/front"
	"github.com/romanSPB15/Calculator_Service_Final/internal/hash"
	"github.com/romanSPB15/Calculator_Service_Final/internal/storage"
	"github.com/romanSPB15/Calculator_Service_Final/pckg/consts"
	"github.com/romanSPB15/Calculator_Service_Final/pckg/types"
	pb "github.com/romanSPB15/Calculator_Service_Final/proto"
	"google.golang.org/grpc"
)

// Приложение
type Application struct {
	NumGoroutine int
	workerId     int
	grpcServer   *grpc.Server
	calcServer   pb.CalculatorServiceServer
	Users        []*types.User
	Tasks        *types.ConcurrentTasksMap
	logger       *log.Logger
	Storage      *storage.Storage
	server       *http.Server
	grpcListener net.Listener
	env          *env.List     // Переменные среды
	agentStop    chan struct{} // Канал остановки агента
	init         bool
}

func New() *Application {
	app := &Application{
		NumGoroutine: 0,
		workerId:     0,
		Tasks:        types.NewConcurrentTasksMap(),
		grpcServer:   grpc.NewServer(),
		logger:       log.New(os.Stdout, "app: ", log.LstdFlags),
	}
	app.calcServer = app.NewServer()
	app.env = env.NewList()
	return app
}

func (a *Application) LoadData() error {
	usersList, err := a.Storage.SelectAllUsers()
	if err != nil {
		return err
	}
	a.Users = usersList
	for _, u := range usersList {
		expressionsList, err := a.Storage.SelectExpressionsForUser(u)
		if err != nil {
			return err
		}
		u.Expressions = make(types.ExpressionsMap)
		for _, v := range expressionsList {
			u.Expressions[v.ID] = &v.Expression
		}
	}
	return nil
}

func (a *Application) SaveData() error {
	err := a.Storage.Clear()
	if err != nil {
		return err
	}
	for _, u := range a.Users {
		err = a.Storage.InsertUser(u)
		if err != nil {
			return err
		}
		for id, expr := range u.Expressions {
			err = a.Storage.InsertExpression(&types.ExpressionWithID{Expression: *expr, ID: id}, u)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *Application) GetUser(login, password string) (u *types.User, ok bool) {
	for _, v := range a.Users {
		if hash.Compare(v.Password, password) && v.Login == login { // пользователь найден!
			u = v
			ok = true
			return
		}
	}
	return
}

func (a *Application) GetUserByID(id types.UserID) (u *types.User, ok bool) {
	for _, v := range a.Users {
		if v.ID == id { // пользователь найден!
			u = v
			ok = true
			return
		}
	}
	return
}

func (a *Application) AddUser(login, password string) (types.UserID, error) {
	h, err := hash.Generate(password)
	if err != nil {
		return "", err
	}
	u := &types.User{
		Login:       login,
		Password:    h,
		Expressions: make(types.ExpressionsMap),
		ID:          uuid.New().String(),
	}
	a.Users = append(a.Users, u)
	return u.ID, nil
}

const GRPC_PORT = 8081

func envFile(testFlag bool) string {
	if testFlag {
		return "../../config/.env"
	}
	return "config/.env"
}

func (app *Application) Init(test ...bool) error {
	var err error
	path := consts.AppStoragePath
	testFlag := len(test) > 0 && test[0]
	if testFlag {
		path = consts.TestStoragePath
	}
	app.Storage, err = storage.Open(path)
	if err != nil {
		return err
	}

	err = app.LoadData()
	if err != nil {
		return err
	}
	pb.RegisterCalculatorServiceServer(app.grpcServer, app.calcServer)

	app.env.InitEnv(envFile(testFlag)) // Иницилизация переменных среды

	addr := fmt.Sprintf("%s:%d", app.env.HOST, app.env.PORT)

	mux := http.NewServeMux()
	// Создаём новый mux.Router
	/* Инициализация обработчиков роутера */
	mux.HandleFunc("/api/v1/register", app.RegisterHandler)
	mux.HandleFunc("/api/v1/login", app.LoginHandler)
	mux.HandleFunc("/api/v1/calculate", app.AddExpressionHandler)
	mux.HandleFunc("/api/v1/expressions/{id}", app.GetExpressionHandler)
	mux.HandleFunc("/api/v1/expressions", app.GetExpressionsHandler)
	if !testFlag && app.env.WEB {
		front.SetAddr(addr)
		front.Handle(mux)
	}

	app.server = &http.Server{Addr: addr, Handler: mux}

	addrGRPC := fmt.Sprintf("%s:%d", app.env.HOST, GRPC_PORT)
	app.grpcListener, err = net.Listen("tcp", addrGRPC) // будем ждать запросы по этому адресу
	if err != nil {
		return err
	}
	app.agentStop = make(chan struct{})
	app.init = true
	return nil
}

// Старт системы. Запускает 3 горутины, и ждёт когда они запустятся
func (app *Application) Start() {
	if !app.init {
		app.logger.Fatal("Application is not init")
	}
	// Создаём wait-группу
	wg := &sync.WaitGroup{}
	wg.Add(3)

	// Агент
	go func() {
		// wg.Done() здесь нет, он в runAgent
		if err := app.runAgent(wg); err != nil {
			app.logger.Fatal("falied to run agent: ", err)
		}
	}()

	// GRPC сервер
	go func() {
		if app.env.DEBUG {
			app.logger.Println("GRPC runned")
		}
		wg.Done()
		if err := app.grpcServer.Serve(app.grpcListener); err != nil {
			app.logger.Fatal("GRPC: failed to serve: ", err)
		}
	}()
	go func() { // Основной сервер
		if app.env.DEBUG {
			app.logger.Println("Main runned")
		}
		wg.Done()
		if err := app.server.ListenAndServe(); err != http.ErrServerClosed {
			app.logger.Fatal("Main: failed to serve: ", err) // HTTP
		}
	}()
	wg.Wait() // Ждём, когда все горутины запустятся
}

// Запуск системы. Для остановки ждёт сигналы SIGTERM и SIGINT
func (app *Application) Run(test ...bool) {
	err := app.Init(test...)
	if err != nil {
		app.logger.Fatal("failed to init application: ", err)
	}
	app.Start()

	// Создаём канал для передачи информации о сигналах
	stop := make(chan os.Signal, 1)
	// "Слушаем" перечисленные сигналы
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	// Ждём данных из канала
	<-stop
	app.Stop()
}

func (app *Application) Stop() {
	close(app.agentStop)
	app.grpcServer.GracefulStop()
	app.grpcListener.Close()
	app.server.Shutdown(context.Background())
	defer app.server.Close()
	if app.env.DEBUG {
		app.logger.Println("Gracefully stopped")
	}
	defer app.Storage.Close()
	err := app.SaveData()
	if err != nil {
		app.logger.Fatal("failed to save data: ", err)
	}
}
