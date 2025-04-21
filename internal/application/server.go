package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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

type Server struct {
	app                        *Application
	pb.CalculatorServiceServer // сервис из сгенерированного пакета
}

func (app *Application) NewServer() *Server {
	return &Server{app: app}
}

func (s *Server) Calculate(ctx context.Context, req *pb.CalculateRequest) (*pb.CalculateResponse, error) {
	id := uuid.New().ID()
	e := Expression{req.Expression, WaitStatus, "", 0}
	s.app.Expressions[id] = &e
	go func() {
		e.Status = CalculationStatus
		res, err := rpn.Calc(req.Expression, s.app.Tasks, s.app.Config.Debug)
		if err != nil {
			e.Error = err.Error()
			e.Status = ErrorStatus
		} else {
			e.Status = "OK"
			e.Error = ""
			e.Result = res
		}
	}()
	return &pb.CalculateResponse{Id: id}, nil
}

func (s *Server) GetExpression(ctx context.Context, req *pb.GetExpressionRequest) (*pb.GetExpressionResponse, error) {
	exp, has := s.app.Expressions[req.Id]
	if !has {
		return &pb.GetExpressionResponse{Expression: &pb.Expression{Id: 1<<32 - 1}}, nil
	}
	return &pb.GetExpressionResponse{
		Expression: &pb.Expression{
			Id:     req.Id,
			Status: exp.Status,
			Result: exp.Result,
		},
	}, nil
}

func (s *Server) GetExpressions(ctx context.Context, req *pb.GetExpressionsRequest) (*pb.GetExpressionsResponse, error) {
	var res []*pb.Expression
	for id, v := range s.app.Expressions {
		res = append(res, &pb.Expression{
			Id:     id,
			Status: v.Status,
			Result: v.Result,
		})
	}
	return &pb.GetExpressionsResponse{
		Expressions: res,
	}, nil
}

func (s *Server) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.GetTaskResponse, error) {
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

func (s *Server) SaveTaskResult(ctx context.Context, req *pb.SaveTaskResultRequest) (*pb.SaveTaskResultResponse, error) {
	t, ok := s.app.Tasks.Map()[req.Id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	t.Result = req.Result
	t.Status = "OK"
	close(t.Done)
	return &pb.SaveTaskResultResponse{}, nil
}
