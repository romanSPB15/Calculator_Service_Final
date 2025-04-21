package application

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/romanSPB15/Calculator_Service_Final/pckg/rpn"
	pb "github.com/romanSPB15/Calculator_Service_Final/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Время запроса агента
var AgentReqestTime = time.Millisecond * 1

func (a *Application) worker(resp *pb.GetTaskResponse, client pb.CalculatorServiceClient) {
	n := a.workerId
	a.workerId++
	if a.Config.Debug {
		log.Println("worker runned", n)
	}
	a.NumGoroutine++

	t := &rpn.TaskID{
		ID: resp.Id,
		Task: rpn.Task{
			Arg1:          resp.Arg1,
			Arg2:          resp.Arg2,
			Operation:     resp.Operation,
			OperationTime: int(resp.OperationTime),
		},
	}
	res := t.Run(a.Config.Debug)
	if a.Config.Debug {
		log.Println("worker", n, "add result reqest")
	}
	_, err := client.SaveTaskResult(context.TODO(), &pb.SaveTaskResultRequest{
		Id:     resp.Id,
		Result: res,
	})
	if err != nil {
		log.Fatalf("falied to set result task: %d: %v", resp.Id, err)
	}
	a.NumGoroutine--
	if a.Config.Debug {
		log.Println("worker completed", n)
	}
}

// Запуск агента
func (a *Application) runAgent() error {
	time.Sleep(time.Millisecond * 100)
	res := make(chan error)
	addr := fmt.Sprintf("%s:%d", rpn.HOST, rpn.PORT) // используем адрес сервера
	// установим соединение
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatal("could not connect to grpc server: ", err)
	}
	// закроем соединение, когда выйдем из функции
	defer conn.Close()

	c := pb.NewCalculatorServiceClient(conn)

	go func() {
		if a.Config.Debug {
			log.Println("agent runned")
		}
		for {
			<-time.After(AgentReqestTime)
			if a.NumGoroutine < rpn.COMPUTING_POWER {
				resp, err := c.GetTask(context.TODO(), &pb.GetTaskRequest{})

				if err != nil {
					res <- err
					return
				}
				if resp.Id == 1<<32-1 {
					continue
				}
				if a.Config.Debug {
					log.Println("agent received task")
				}
				go a.worker(resp, c)
			}
		}
	}()
	return <-res
}
