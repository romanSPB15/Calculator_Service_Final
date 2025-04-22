package rpn

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type (
	TaskArg1Type   = float64
	TaskArg2Type   = float64
	TaskResultType = float64
)

type ExpressionResultType = float64

// Задача
type Task struct {
	Arg1          TaskArg1Type   `json:"arg1"`
	Arg2          TaskArg2Type   `json:"arg2"`
	Operation     string         `json:"operation"`
	OperationTime int            `json:"operation_time"`
	Status        string         `json:"-"`
	Result        TaskResultType `json:"-"`
	Done          chan struct{}  `json:"-"`
}

// Мап задач
type TaskMap = map[IDTask]*Task

// Конкурентный мап задач
type ConcurrentTaskMap struct {
	m  TaskMap
	mx sync.Mutex
}

// Новый конкурентный мап задач
func NewConcurrentTaskMap() *ConcurrentTaskMap {
	return &ConcurrentTaskMap{make(map[IDTask]*Task), sync.Mutex{}}
}

// Получение ссылки на задачу.
func (cm *ConcurrentTaskMap) Get(id IDTask) *Task {
	cm.mx.Lock()
	res, ok := cm.m[id]
	if !ok {
		t := &Task{}
		cm.m[id] = t
		cm.mx.Unlock()
		return t
	}
	cm.mx.Unlock()
	return res
}

// Добавление ссылки на задачу.
func (cm *ConcurrentTaskMap) Add(id IDTask, t *Task) {
	cm.mx.Lock()
	cm.m[id] = t
	cm.mx.Unlock()
}

// Получение простого мапа
func (cm *ConcurrentTaskMap) Map() map[IDTask]*Task {
	return cm.m
}

// Задача с ID
type TaskID struct {
	ID IDTask `json:"id"`
	Task
}

func (t *TaskID) Run(debug bool, logger *log.Logger) (res float64) {
	if debug {
		logger.Printf("Task %d Runned\r\n", t.ID)
	}
	s := time.Now()
	switch t.Operation {
	case "+":
		res = t.Arg1 + t.Arg2
	case "-":
		res = t.Arg1 - t.Arg2
	case "*":
		res = t.Arg1 * t.Arg2
	case "/":
		res = t.Arg1 / t.Arg2
	}
	d := time.Since(s)
	d = (time.Millisecond * time.Duration(t.OperationTime)) - d
	time.Sleep(d)
	if debug {
		logger.Printf("Task %d Completed With Result %.2F\r\n", t.ID, res)
	}
	return
}

func convertString(str string) float64 {
	res, err := strconv.ParseFloat(str, 64)
	if err != nil {
		panic(err)
	}
	return res
}

func isSign(value rune) bool {
	return value == '+' || value == '-' || value == '*' || value == '/'
}

type IDTask = uint32

var Errorexp = errors.New("expression is not valid")
var Errordel = errors.New("division by zero")

func Calc(expression string, tasks *ConcurrentTaskMap, debug bool, logger *log.Logger) (res ExpressionResultType, err0 error) {
	if len(expression) < 3 {
		return 0, Errorexp
	}
	//////////////////////////////////////////////////////////////////////////////////////////////////////
	b := ""
	c := rune(0)
	resflag := false
	isc := -1
	scc := 0
	//////////////////////////////////////////////////////////////////////////////////////////////////////
	if isSign(rune(expression[0])) || isSign(rune(expression[len(expression)-1])) {
		return 0, Errorexp
	}
	if strings.Contains(expression, "(") || strings.Contains(expression, ")") {
		for i := 0; i < len(expression); i++ {
			value := expression[i]
			if value == '(' {
				if scc == 0 {
					isc = i
				}
				scc++
			}
			if value == ')' {
				scc--
				if scc == 0 {
					exp := expression[isc+1 : i]
					calc, err := Calc(exp, tasks, debug, logger)
					if err != nil {
						return 0, err
					}
					calcstr := strconv.FormatFloat(calc, 'f', 0, 64)
					expression = strings.Replace(expression, expression[isc:i+1], calcstr, 1) // Меняем скобки на результат выражения в них

					i -= len(exp)
					isc = -1
				}
			}
		}
	}
	if isc != -1 {
		return 0, Errorexp
	}
	priority := strings.ContainsRune(expression, '*') || strings.ContainsRune(expression, '/')
	notpriority := strings.ContainsRune(expression, '+') || strings.ContainsRune(expression, '-')
	if priority && notpriority {
		for i := 1; i < len(expression); i++ {
			value := rune(expression[i])
			///////////////////////////////////////////////////////////////////////////////////////////////////////////////
			//Умножение и деление
			if value == '*' || value == '/' {
				var imin int = i - 1
				if imin != 0 {
					for imin >= 0 {
						if imin >= 0 {
							if isSign(rune(expression[imin])) {
								break
							}
						}
						imin--
					}
					imin++
				}
				imax := i + 1
				if imax == len(expression) {
					imax--
				} else {
					for !isSign(rune(expression[imax])) && imax < len(expression)-1 {
						imax++
					}
				}
				if imax == len(expression)-1 {
					imax++
				}
				exp := expression[imin:imax]
				calc, err := Calc(exp, tasks, debug, logger)
				if err != nil {
					return 0, err
				}
				calcstr := fmt.Sprint(calc)
				expression = strings.Replace(expression, expression[imin:imax], calcstr, 1) // Меняем скобки на результат выражения в них
				i -= len(exp) - 1
			}
			if value == '+' || value == '-' || value == '*' || value == '/' {
				c = value
			}
		}
	}
	//////////////////////////////////////////////////////////////////////////////////////////////////////
	for _, value := range expression + "s" {
		switch {
		case value == ' ':
			continue
		case value > 47 && value < 58 || value == '.': // Если это цифра
			b += string(value)
		case isSign(value) || value == 's': // Если это знак
			if resflag {
				switch c {
				case '+':
					uuid := uuid.New()
					id := uuid.ID()
					t := Task{
						Arg1:          res,
						Arg2:          convertString(b),
						Operation:     "+",
						Status:        "Wait",
						OperationTime: TIME_ADDITION_MS,
						Done:          make(chan struct{}),
					}
					if debug {
						logger.Println("rpn.Calc: Create New Task With ID", id)
					}

					tasks.Add(id, &t) // Записываем задачу
					<-t.Done
					res = t.Result
					if debug {
						logger.Printf("Result Task %d(%.2F) is handle in Calc", id, t.Result)
					}
				case '-':
					uuid := uuid.New()
					id := uuid.ID()
					t := Task{
						Arg1:          res,
						Arg2:          convertString(b),
						Operation:     "-",
						Status:        "Wait",
						OperationTime: TIME_SUBTRACTION_MS,
						Done:          make(chan struct{}),
					}
					if debug {
						logger.Println("rpn.Calc: Create New Task With ID", id)
					}

					tasks.Add(id, &t) // Записываем задачу
					<-t.Done
					res = t.Result
					if debug {
						logger.Printf("Result Task %d(%.2F) is handle in Calc", id, t.Result)
					}
				case '*':
					uuid := uuid.New()
					id := uuid.ID()
					t := Task{
						Arg1:          res,
						Arg2:          convertString(b),
						Operation:     "*",
						Status:        "Wait",
						OperationTime: TIME_MULTIPLICATIONS_MS,
						Done:          make(chan struct{}),
					}
					if debug {
						logger.Println("rpn.Calc: Create New Task With ID", id)
					}

					tasks.Add(id, &t) // Записываем задачу
					<-t.Done
					res = t.Result
					if debug {
						logger.Printf("Result Task %d(%.2F) is handle in Calc", id, t.Result)
					}
				case '/':
					uuid := uuid.New()
					id := uuid.ID()
					arg2 := convertString(b)
					if arg2 == 0 {
						return 0, Errordel
					}
					t := Task{
						Arg1:          res,
						Arg2:          arg2,
						Operation:     "/",
						Status:        "Wait",
						OperationTime: TIME_DIVISIONS_MS,
						Done:          make(chan struct{}),
					}
					if debug {
						logger.Println("rpn.Calc: Create New Task With ID", id)
					}

					tasks.Add(id, &t) // Записываем задачу
					<-t.Done
					res = t.Result
					if debug {
						logger.Printf("Result Task %d(%.2F) is handle in Calc", id, t.Result)
					}
				}
			} else {
				resflag = true
				res = convertString(b)
			}
			b = ""
			c = value

			/////////////////////////////////////////////////////////////////////////////////////////////
		case value == 's':
		default:
			return 0, Errorexp
		}
	}
	return res, nil
}
