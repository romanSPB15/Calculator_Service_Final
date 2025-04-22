package application

import (
	"fmt"
	"log"
	"net"
	"os"

	//"github.com/romanSPB15/Calculator_Service_Final/internal/web"
	"github.com/romanSPB15/Calculator_Service_Final/pckg/dir"
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
	authServer   *AuthServer
	logger       *log.Logger
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
	app.authServer = NewAuthServer(app)
	rpn.InitEnv(dir.EnvFile())
	return app
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
	}
	a.Users = append(a.Users, u)
}

// Запуск всей системы
func (app *Application) Run() {
	addr := fmt.Sprintf("%s:%d", rpn.HOST, rpn.PORT)
	lis, err := net.Listen("tcp", addr) // будем ждать запросы по этому адресу
	if err != nil {
		panic(err)
	}
	rpn.InitEnv(dir.EnvFile()) // Иницилизация переменных из среды

	pb.RegisterCalculatorServiceServer(app.grpcServer, app.calcServer)

	go func() {
		app.logger.Fatal("falied to run agent ", app.runAgent())
	}()
	if app.Config.Debug {
		app.logger.Println("main server runned")
	}

	app.logger.Println("App runned")
	/*if app.Config.Web {
		go web.Run()
	}*/
	go func() {
		app.logger.Fatal("auth: failed to serving: ", app.authServer.ListenAndServe())
	}()
	if err := app.grpcServer.Serve(lis); err != nil {
		app.logger.Fatal("failed to serving grpc: ", err)
	}
}

func (app *Application) Stop() {
	app.grpcServer.GracefulStop()
	app.authServer.Close()
	app.logger.Println("Gracefully stopped")
}
