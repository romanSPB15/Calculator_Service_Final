package application

import (
	"fmt"
	"log"
	"net"

	"github.com/romanSPB15/Calculator_Service_Final/pckg/dir"
	"github.com/romanSPB15/Calculator_Service_Final/pckg/rpn"
	pb "github.com/romanSPB15/Calculator_Service_Final/proto"
	"google.golang.org/grpc"
)

var current *Application

func Current() *Application {
	return current
}

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

// Приложение
type Application struct {
	// Агент
	Config       *config
	NumGoroutine int
	workerId     int
	Expressions  map[IDExpression]*Expression
	Tasks        *rpn.ConcurrentTaskMap
}

func New() *Application {
	app := &Application{
		Config:       newConfig(),
		NumGoroutine: 0,
		workerId:     0,
		Expressions:  make(map[IDExpression]*Expression),
		Tasks:        rpn.NewConcurrentTaskMap(),
	}
	current = app
	return app
}

const IP = "localhost:8080"

// Запуск всей системы
func (app *Application) Run() {
	host := "localhost"
	port := "8080"

	addr := fmt.Sprintf("%s:%s", host, port)
	lis, err := net.Listen("tcp", addr) // будем ждать запросы по этому адресу
	if err != nil {
		panic(err)
	}
	rpn.InitEnv(dir.EnvFile()) // Иницилизация переменных из среды
	grpcServer := grpc.NewServer()
	calcServer := app.NewServer()

	pb.RegisterCalculatorServiceServer(grpcServer, calcServer)

	go func() {
		log.Fatal("falied to run agent ", app.runAgent())
	}()
	if app.Config.Debug {
		log.Println("main server runned")
	}
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("failed to serving grpc: ", err)
	}
}
