package application

import (
	"context"
	"fmt"

	"github.com/romanSPB15/Calculator_Service_Final/pckg/rpn"
	pb "github.com/romanSPB15/Calculator_Service_Final/proto"
)

/*
type CalculatorServiceServer interface {
	// методы, которые можно будет реализовать и использовать
	Calculate(context.Context, *CalculateRequest) (*CalculateResponse, error)
	GetExpression(context.Context, *GetExpressionRequest) (*GetExpressionResponse, error)
	GetExpressions(context.Context, *GetExpressionsRequest) (*GetExpressionsResponse, error)
	GetTask(context.Context, *GetTaskRequest) (*GetTaskResponse, error)
	SetTaskResult(context.Context, *SaveTaskResultResponse) (*SaveTaskResultResponse, error)
}
*/

type GRPCServer struct {
	app                        *Application
	pb.CalculatorServiceServer // сервис из сгенерированного пакета
}

func (app *Application) NewServer() *GRPCServer {
	return &GRPCServer{app: app}
}

// Получение задачи на выполнение
func (s *GRPCServer) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.GetTaskResponse, error) {
	var id rpn.IDTask = 1<<32 - 1
	for k, v := range s.app.Tasks.Map() {
		if v.Status != "OK" {
			id = k
			break
		}
	}
	if id == 1<<32-1 {
		return &pb.GetTaskResponse{Arg1: -1, Arg2: -1, OperationTime: -1, Id: id, Operation: "invalid"}, nil
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
		return nil, fmt.Errorf("not found")
	}
	t.Result = req.Result
	t.Status = "OK"
	close(t.Done)
	return &pb.SaveTaskResultResponse{}, nil
}
