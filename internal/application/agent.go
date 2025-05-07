package application

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/romanSPB15/Calculator_Service_Final/pckg/consts"
	"github.com/romanSPB15/Calculator_Service_Final/pckg/types"
	pb "github.com/romanSPB15/Calculator_Service_Final/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (app *Application) worker(resp *pb.GetTaskResponse, client pb.CalculatorServiceClient) {
	app.workerId++

	app.NumGoroutine++

	t := &types.Task{
		Arg1:          resp.Arg1,
		Arg2:          resp.Arg2,
		Operation:     resp.Operation,
		OperationTime: int(resp.OperationTime),
	}
	res := t.Run()

	_, err := client.SaveTaskResult(context.Background(), &pb.SaveTaskResultRequest{
		Id:     resp.Id,
		Result: res,
	})
	if err != nil {
		app.logger.Fatalf("Falied to set result task: %d: %v", resp.Id, err)
	}

	app.NumGoroutine--
}

// Запуск агента
func (app *Application) runAgent(wg *sync.WaitGroup) error {
	time.Sleep(time.Millisecond * 10)
	res := make(chan error)
	addr := fmt.Sprintf("%s:%d", app.env.HOST, GRPC_PORT) // используем адрес сервера
	// установим соединение
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		app.logger.Fatal("Agent: Could not connect to grpc server: ", err)
	}
	// закроем соединение, когда выйдем из функции
	defer conn.Close()

	c := pb.NewCalculatorServiceClient(conn)

	go func() {
		if app.env.DEBUG {
			app.logger.Println("Agent runned")
		}
		wg.Done()
		for {
			select {
			case <-time.After(consts.AgentReqestDelay):
				if app.NumGoroutine < app.env.COMPUTING_POWER {
					resp, err := c.GetTask(context.Background(), &pb.GetTaskRequest{})
					if err != nil {
						if err.Error() == TaskNotFound.Error() {
							continue
						}
						res <- err
						return
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
