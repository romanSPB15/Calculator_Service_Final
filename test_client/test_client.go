package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	pb "github.com/romanSPB15/Calculator_Service_Final/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func WaitResult(c pb.CalculatorServiceClient, resp *pb.CalculateResponse, token string) (float64, string) {
	for {
		<-time.After(time.Millisecond * 100)
		res, err := c.GetExpression(context.TODO(), &pb.GetExpressionRequest{Id: resp.Id, Token: token})
		if err != nil {
			logger.Fatal(err)
		}
		if res.Expression.Status == "OK" || res.Expression.Status == "Error" {
			return res.Expression.Result, res.Expression.Status
		}
	}
}

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "calculator: ", log.Llongfile)
}

func main() {
	host := "localhost"
	port := "8080"

	addr := fmt.Sprintf("%s:%s", host, port) // используем адрес сервера
	// установим соединение
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		logger.Fatal("could not connect to grpc server: ", err)
	}
	// закроем соединение, когда выйдем из функции
	defer conn.Close()

	c := pb.NewCalculatorServiceClient(conn)
	_, err = http.DefaultClient.Post("http://localhost:8181/api/v1/register", "application/json", strings.NewReader("{\"login\": \"roman\", \"password\": \"asdfgh\"}"))
	if err != nil {
		logger.Println("falied to register: ", err)
	}
	resp, err := http.DefaultClient.Post("http://localhost:8181/api/v1/login", "application/json", strings.NewReader("{\"login\": \"roman\", \"password\": \"asdfgh\"}"))
	if err != nil {
		logger.Fatal("falied to login: ", err)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Fatal("falied to read token: ", err)
	}
	token := string(b)
	expr := ""
	for {
		fmt.Scan(&expr)
		time.Sleep(time.Second)
		resp, err := c.Calculate(context.TODO(), &pb.CalculateRequest{
			Expression: expr,
			Token:      token,
		})
		if err != nil {
			logger.Println(err)
		} else {
			fmt.Println(WaitResult(c, resp, token))
		}
	}
}
