package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/romanSPB15/Calculator_Service_Final/pckg/consts"
	"github.com/romanSPB15/Calculator_Service_Final/pckg/types"
)

var (
	pasw, login, token string
	last               time.Time
)

func fail(text string, err error) {
	fmt.Println(text, err)
	os.Exit(1)
}

// file content contants
const (
	lastTimeFormat = time.DateTime // last login time format
	contentFormat  = "%v\n%v\n%v\n%v\n"
)

func loadInfo() {
	f, err := os.OpenFile("conn.txt", os.O_RDWR, 0644)
	if err != nil {
		fail("failed to open conn file:", err)
	}
	defer f.Close()
	strTime := ""
	_, err = fmt.Fscanf(f, contentFormat, &login, &pasw, &strTime, &token)
	last, _ = time.Parse(lastTimeFormat, strTime)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Please enter your login")
		fmt.Scanln(&login)
		fmt.Println("Please enter your password")
		fmt.Scanln(&pasw)
		last = time.Now()
		resp, err := http.Post("http://localhost:8080/api/v1/register", "application/json", strings.NewReader(fmt.Sprintf(`{"login": "%s", "password": "%s"}`, login, pasw)))
		if err != nil {
			fail("failed to regster:", err)
		}
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != 200 {
			fail("failed to register:", fmt.Errorf("%s: %s", resp.Status, string(b)))
		}

		token = string(b)
	}
	if time.Now().After(last.Add(time.Minute * 15)) {
		Login()
	}
	f.Seek(0, io.SeekStart)
	fmt.Fprintf(f, contentFormat, login, pasw, last.Format(lastTimeFormat), token)
	fmt.Println(token)
}

func wait(id uint32) (float64, types.Status) {
	for {
		<-time.After(time.Millisecond * 10)
		url := fmt.Sprintf("http://localhost:8080/api/v1/expressions/%d", id)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			fail("falied to make request:", err)
		}
		req.Header.Set("Authorization-Bearer", token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fail("failed to get expression:", err)
		}
		if resp.StatusCode != 200 {
			fail("failed to get expression: status code: ", fmt.Errorf(resp.Status))
		}
		geResp := new(types.GetExpressionHandlerResult)
		json.NewDecoder(resp.Body).Decode(geResp)
		resp.Body.Close()
		if geResp.Expression.Status != consts.WaitStatus && geResp.Expression.Status != consts.CalculationStatus {
			return geResp.Expression.Result, geResp.Expression.Status
		}
	}
}

func Login() {
	resp, err := http.Post("http://localhost:8080/api/v1/login", "application/json", strings.NewReader(fmt.Sprintf(`{"login": "%s", "password": "%s"}`, login, pasw)))
	if err != nil {
		fail("failed to login:", err)
	}
	if resp.StatusCode != 200 {
		fail("failed to login:", fmt.Errorf(resp.Status))
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	last = time.Now()
	token = string(b)
}

func main() {
	loadInfo()
	stop := make(chan os.Signal, 1)
	ch := make(chan string, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	f := func() {
		var expr string = "2+2"
		fmt.Println("Please enter your expression:")
		fmt.Scan(&expr)
		ch <- expr
	}
	go f()
	for {
		select {
		case <-stop:
			return
		case expr := <-ch:
			if func() bool {
				if expr == "" {
					return true
				}
				url := "http://localhost:8080/api/v1/calculate"
				body := strings.NewReader(fmt.Sprintf(`{"expression": "%s"}`, expr))
				req, err := http.NewRequest(http.MethodPost, url, body)
				if err != nil {
					fail("falied to make request:", err)
				}
				req.Header.Set("Authorization-Bearer", token)

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					fail("falied to calculate:", err)
				}
				defer resp.Body.Close()
				if resp.StatusCode == 201 {
					Login()
					body.Seek(0, io.SeekStart)

					req, _ := http.NewRequest(http.MethodPost, url, body)
					resp, err = http.DefaultClient.Do(req)
					if err != nil {
						fail("falied to calculate:", err)
					}
				}
				if resp.StatusCode != 200 {
					b, _ := io.ReadAll(resp.Body)
					fail("falied to calculate:", fmt.Errorf("%s: %s", resp.Status, string(b)))
				}
				cResp := new(types.CalculateHandlerResponse)
				json.NewDecoder(resp.Body).Decode(cResp)
				resp.Body.Close()
				res, s := wait(cResp.ID)
				if s == consts.ErrorStatus {
					fmt.Println("Invalid expression")
					go f()
					return false
				}
				fmt.Println(res)
				go f()
				return false
			}() {
				return
			}
		}
	}
}
