package application

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/romanSPB15/Calculator_Service_Final/pckg/rpn"
	pb "github.com/romanSPB15/Calculator_Service_Final/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

// Все запросы с токеном
type RequestWithToken interface {
	GetToken() string
}

// Получить, есть ли такой пользователь
func (s *Server) GetUserByToken(rwt RequestWithToken) (*User, error) {
	token, err := jwt.Parse(rwt.GetToken(), func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid token")
		}

		return []byte(SECRETKEY), nil
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}
	login := token.Claims.(jwt.MapClaims)["login"]
	password := token.Claims.(jwt.MapClaims)["password"]
	u, ok := s.app.GetUser(login.(string), password.(string))
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "user not found")
	}
	return u, nil
}

// Вычисление выражения
func (s *Server) Calculate(ctx context.Context, req *pb.CalculateRequest) (*pb.CalculateResponse, error) {
	u, err := s.GetUserByToken(req)
	if err != nil {
		return nil, err
	}
	id := uuid.New().ID()
	e := Expression{req.Expression, WaitStatus, 0}
	u.Expressions[id] = &e
	go func() {
		e.Status = CalculationStatus
		res, err := rpn.Calc(req.Expression, s.app.Tasks, s.app.Config.Debug, s.app.logger)
		if err != nil {
			e.Status = ErrorStatus
		} else {
			e.Status = "OK"
			e.Result = res
		}
	}()
	return &pb.CalculateResponse{Id: id}, nil
}

// Получение выражения пользователя по его ID
func (s *Server) GetExpression(ctx context.Context, req *pb.GetExpressionRequest) (*pb.GetExpressionResponse, error) {
	u, err := s.GetUserByToken(req)
	if err != nil {
		return nil, err
	}
	exp, has := u.Expressions[req.Id]
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

// Получение всех выражений пользователя
func (s *Server) GetExpressions(ctx context.Context, req *pb.GetExpressionsRequest) (*pb.GetExpressionsResponse, error) {
	u, err := s.GetUserByToken(req)
	if err != nil {
		return nil, err
	}
	var res []*pb.Expression
	for id, v := range u.Expressions {
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

// Получение задачи на выполнение
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
