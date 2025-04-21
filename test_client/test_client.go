package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/romanSPB15/Calculator_Service_Final/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func WaitResult(c pb.CalculatorServiceClient, resp *pb.CalculateResponse) (float64, string) {
	for {
		<-time.After(time.Millisecond * 100)
		res, err := c.GetExpression(context.TODO(), &pb.GetExpressionRequest{Id: resp.Id})
		if err != nil {
			log.Fatal(err)
		}
		if res.Expression.Status == "OK" || res.Expression.Status == "Error" {
			return res.Expression.Result, res.Expression.Status
		}
	}
}

func main() {
	host := "localhost"
	port := "8080"

	addr := fmt.Sprintf("%s:%s", host, port) // используем адрес сервера
	// установим соединение
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatal("could not connect to grpc server: ", err)
	}
	// закроем соединение, когда выйдем из функции
	defer conn.Close()

	c := pb.NewCalculatorServiceClient(conn)

	for {
		expr := ""
		fmt.Scan(&expr)
		resp, err := c.Calculate(context.TODO(), &pb.CalculateRequest{
			Expression: expr,
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(WaitResult(c, resp))
	}
}
