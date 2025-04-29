package application

import (
	"context"
	"fmt"
	"time"

	"github.com/romanSPB15/Calculator_Service_Final/pckg/rpn"
	pb "github.com/romanSPB15/Calculator_Service_Final/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Время запроса агента
var AgentReqestTime = time.Millisecond * 1

func (app *Application) worker(resp *pb.GetTaskResponse, client pb.CalculatorServiceClient) {
	n := app.workerId
	app.workerId++
	if app.Config.Debug {
		app.logger.Println("worker runned", n)
	}
	app.NumGoroutine++

	t := &rpn.TaskID{
		ID: resp.Id,
		Task: rpn.Task{
			Arg1:          resp.Arg1,
			Arg2:          resp.Arg2,
			Operation:     resp.Operation,
			OperationTime: int(resp.OperationTime),
		},
	}
	res := t.Run(app.Config.Debug, app.logger)
	if app.Config.Debug {
		app.logger.Println("worker", n, "add result reqest")
	}
	_, err := client.SaveTaskResult(context.TODO(), &pb.SaveTaskResultRequest{
		Id:     resp.Id,
		Result: res,
	})
	if err != nil {
		app.logger.Fatalf("falied to set result task: %d: %v", resp.Id, err)
	}
	app.NumGoroutine--
	if app.Config.Debug {
		app.logger.Println("worker completed", n)
	}
}

// Запуск агента
func (app *Application) runAgent() error {
	time.Sleep(time.Millisecond * 100)
	res := make(chan error)
	addr := fmt.Sprintf("%s:%d", app.env.HOST, GRPC_PORT) // используем адрес сервера
	// установим соединение
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		app.logger.Fatal("could not connect to grpc server: ", err)
	}
	// закроем соединение, когда выйдем из функции
	defer conn.Close()

	c := pb.NewCalculatorServiceClient(conn)

	go func() {
		if app.Config.Debug {
			app.logger.Println("agent runned")
		}
		for {
			select {
			case <-time.After(AgentReqestTime):
				if app.NumGoroutine < app.env.COMPUTING_POWER {
					resp, err := c.GetTask(context.Background(), &pb.GetTaskRequest{})
					if err != nil {
						if err.Error() == TaskNotFound.Error() {
							continue
						}
						res <- err
						return
					}
					if app.Config.Debug {
						app.logger.Println("agent received task")
					}
					go app.worker(resp, c)
				}
			case <-app.agentStop:
				res <- nil
				return
			}

		}
	}()
	return <-res
}
