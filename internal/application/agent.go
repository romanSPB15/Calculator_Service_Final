package application

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/romanSPB15/Calculator_Service_Final/pckg/rpn"
)

// Время запроса агента
var AgentReqestTime = time.Millisecond * 1

func (a *Application) worker(body io.ReadCloser) {
	n := a.NumGoroutine
	if a.Config.Debug {
		log.Println("worker runned", n)
	}
	a.NumGoroutine++
	defer body.Close()
	b, err := io.ReadAll(body)
	if err != nil {
		panic(err)
	}
	var ResultServer GetTaskHandlerResult
	err = json.Unmarshal(b, &ResultServer)
	if err != nil {
		panic(err)
	}
	t := ResultServer.Task
	res, err := json.Marshal(AgentResult{ResultServer.Task.ID, t.Run(a.Config.Debug)})
	if err != nil {
		panic(err)
	}
	if a.Config.Debug {
		log.Println("worker", n, "add result reqest")
	}
	resp, err := a.Agent.Post("http://localhost:8080/api/v1/internal/task", "application/json", bytes.NewReader(res))
	if err != nil {
		panic(err)
	}
	if a.Config.Debug {
		log.Println(resp.Status)
		log.Println(io.ReadAll(resp.Body))
	}
	a.NumGoroutine--
	if a.Config.Debug {
		log.Println("worker completed", n)
	}
}

// Запуск агента
func (a *Application) runAgent() error {
	var res error
	done := make(chan struct{})
	go func() {
		if a.Config.Debug {
			log.Println("Agent Runned")
		}
		for {
			<-time.After(AgentReqestTime)
			if a.NumGoroutine < rpn.COMPUTING_POWER {
				resp, err := a.Agent.Get("http://localhost:8080/api/v1/internal/task")
				if err != nil {
					res = err
					return
				}
				if resp.StatusCode == http.StatusNotFound {
					continue
				}
				if a.Config.Debug {
					log.Println("agent received task")
				}
				defer resp.Body.Close()
				go a.worker(resp.Body)
			}
		}
	}()
	<-done
	return res
}
