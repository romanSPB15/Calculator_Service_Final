package application_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"net/http/httptest"

	app "github.com/romanSPB15/Calculator_Service_Final/internal/application"
	"github.com/romanSPB15/Calculator_Service_Final/pckg/rpn"
)

var testApp = app.New()

func TestAddHandlerAndGetExpressionsHandler(t *testing.T) {
	rpn.InitEnv("test.env")
	expressions := []string{"1+1", "2+2", "3+3", "4+4", "5+5"}
	for _, exp := range expressions {
		r := strings.NewReader(fmt.Sprintf("{\"expression\": \"%s\"}", exp))
		req := httptest.NewRequest("POST", "http://localhost/api/v1/calculate", r)
		w := httptest.NewRecorder()
		testApp.AddExpressionHandler(w, req)
		if w.Code != 201 {
			t.Fatalf("Status code(%d) != 201", w.Code)
		}
		resp := w.Result()
		resp.Body.Close()
	}
	url := `http://localhost/api/v1/expressions`
	req := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()
	testApp.GetExpressionsHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	if w.Code != 200 {
		t.Fatalf("Status code(%d) != 200", w.Code)
	}
}

func TestAddHandlerAndGetTaskHandler(t *testing.T) {
	rpn.InitEnv("test.env")
	r := strings.NewReader("{\"expression\": \"2+8\"}")
	req := httptest.NewRequest("POST", "http://localhost/api/v1/calculate", r)
	w := httptest.NewRecorder()
	testApp.AddExpressionHandler(w, req)
	url := `http://localhost/api/v1/internal/task`
	time.Sleep(100 * time.Millisecond)
	req = httptest.NewRequest("GET", url, nil)
	w = httptest.NewRecorder()
	testApp.TaskHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	if w.Code != 200 {
		t.Fatalf("Status code(%d) != 200", w.Code)
	}
}
