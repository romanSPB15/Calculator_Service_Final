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

	"github.com/joho/godotenv"
	pb "github.com/romanSPB15/Calculator_Service_Final/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func WaitResult(c pb.CalculatorServiceClient, resp *pb.CalculateResponse, token string) (float64, string) {
	for {
		<-time.After(time.Millisecond * 10)
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
	err := godotenv.Load(`./test_client/.env`, `./config/.env`)
	if err != nil {
		logger.Fatal("falied to load env: ", err)
	}
	host, has := os.LookupEnv("HOST")
	if !has {
		logger.Fatal("falied to open USERNAME")
	}
	port, has := os.LookupEnv("PORT")
	if !has {
		logger.Fatal("falied to open PORT")
	}

	addr := fmt.Sprintf("%s:%s", host, port) // используем адрес сервера
	// установим соединение
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		logger.Fatal("could not connect to grpc server: ", err)
	}
	// закроем соединение, когда выйдем из функции
	defer conn.Close()
	c := pb.NewCalculatorServiceClient(conn)

	password, has := os.LookupEnv("CLIENT_PASSWORD")
	if !has {
		logger.Fatal("falied to open CLIENT_PASSWORD")
	}
	login, has := os.LookupEnv("CLIENT_LOGIN")
	if !has {
		logger.Fatal("falied to open CLIENT_PASSWORD")
	}
	fmt.Printf("Registering with username %s and password %s...\r\n", login, password)
	body := strings.NewReader(fmt.Sprintf(`{"login": "%s", "password": "%s"}`, login, password))
	resp, err := http.DefaultClient.Post("http://localhost:8181/api/v1/register", "application/json", body)
	if err != nil {
		logger.Fatal("falied to register: ", err)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Fatal("falied to read body: ", err)
	}
	if resp.StatusCode != 200 && string(b) != "user already exists\n" {
		logger.Fatalln("falied to register: status code", resp.StatusCode, string(b))
	}
	body.Seek(0, 0)
	resp, err = http.DefaultClient.Post("http://localhost:8181/api/v1/login", "application/json", body)
	if err != nil {
		logger.Fatal("falied to login: ", err)
	}
	b, err = io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		logger.Fatalln("falied to login: status code", resp.StatusCode, string(b))
	}
	if err != nil {
		logger.Fatal("falied to read body: ", err)
	}
	token := string(b)

	expr := "2+2"

	fmt.Println("--------------Calculator----------------")

	for {
		fmt.Scan(&expr)
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
