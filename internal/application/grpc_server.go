package application

import (
	"context"

	"github.com/romanSPB15/Calculator_Service_Final/pckg/rpn"
	pb "github.com/romanSPB15/Calculator_Service_Final/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	app                        *Application
	pb.CalculatorServiceServer // сервис из сгенерированного пакета
}

func (app *Application) NewServer() *GRPCServer {
	return &GRPCServer{app: app}
}

var (
	TaskNotFound = status.Errorf(codes.NotFound, "not found task")
)

// Получение задачи на выполнение
func (s *GRPCServer) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.GetTaskResponse, error) {
	var id rpn.IDTask = 1<<32 - 1
	for k, v := range s.app.Tasks.Map() {
		if v.Status == WaitStatus {
			id = k
			break
		}
	}
	if id == 1<<32-1 {
		return nil, TaskNotFound
	}
	t := s.app.Tasks.Get(id)
	return &pb.GetTaskResponse{
		Id:            id,
		Arg1:          t.Arg1,
		Arg2:          t.Arg2,
		Operation:     t.Operation,
		OperationTime: int32(t.OperationTime),
	}, nil
}

func (s *GRPCServer) SaveTaskResult(ctx context.Context, req *pb.SaveTaskResultRequest) (*pb.SaveTaskResultResponse, error) {
	t, ok := s.app.Tasks.Map()[req.Id]
	if !ok {
		return nil, TaskNotFound
	}
	t.Result = req.Result
	t.Status = "OK"
	close(t.Done)
	return &pb.SaveTaskResultResponse{}, nil
}
