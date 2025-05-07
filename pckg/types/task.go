package types

import "time"

// Задача - без ID, он в мапе
type Task struct {
	Arg1          TaskArg1Type   `json:"arg1"`
	Arg2          TaskArg2Type   `json:"arg2"`
	Operation     string         `json:"operation"`
	OperationTime int            `json:"operation_time"`
	Status        Status         `json:"-"`
	Result        TaskResultType `json:"-"`
	Done          chan struct{}  `json:"-"`
}

func (t *Task) Run() (res float64) {
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
	return
}

type TaskID = uint32
