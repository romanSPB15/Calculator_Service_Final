package rpn

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/romanSPB15/Calculator_Service_Final/internal/env"
	"github.com/romanSPB15/Calculator_Service_Final/pckg/consts"
	"github.com/romanSPB15/Calculator_Service_Final/pckg/types"
)

// Конвертирует строку в float64. Если произойдёт ошибка, то будет panic
func convertString(str string) float64 {
	str = strings.ReplaceAll(str, ",", ".") // запятые strconv.ParseFloat() не поддерживает, меняем на точки
	res, err := strconv.ParseFloat(str, 64)
	if err != nil {
		panic(err)
	}
	return res
}

// Проверяет, является ли руна знаком действия
func isSign(value rune) bool {
	return value == '+' || value == '-' || value == '*' || value == '/'
}

var ErrorInvalidExpr = errors.New("expression is invalid")
var ErrorDivByZero = errors.New("division by zero")

func Calc(expression string, tasks *types.ConcurrentTasksMap, env *env.List) (res float64, err0 error) {
	if len(expression) < 3 {
		return 0, ErrorInvalidExpr
	}
	//////////////////////////////////////////////////////////////////////////////////////////////////////
	b := ""
	c := rune(0)
	resflag := false
	isc := -1
	scc := 0
	//////////////////////////////////////////////////////////////////////////////////////////////////////
	if isSign(rune(expression[0])) || isSign(rune(expression[len(expression)-1])) { // Если в начале или в конце знак
		return 0, ErrorInvalidExpr
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
					calc, err := Calc(exp, tasks, env)
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
	if isc != -1 { // Нет закрывающейся скобки - как миниум 1
		return 0, ErrorInvalidExpr
	}
	priority := strings.ContainsRune(expression, '*') || strings.ContainsRune(expression, '/')    // Приоритетные знаки - * и /
	notPriority := strings.ContainsRune(expression, '+') || strings.ContainsRune(expression, '-') // Неприоритетные знаки - + и -
	if !priority && !notPriority {                                                                // Знаков нет
		return 0, ErrorInvalidExpr
	}
	if priority && notPriority { // Если и то, и другое
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
				calc, err := Calc(exp, tasks, env)
				if err != nil {
					return 0, err
				}
				calcstr := fmt.Sprint(calc)
				expression = strings.Replace(expression, expression[imin:imax], calcstr, 1) // Меняем  на результат
				i = imin
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
		case value > 47 && value < 58 || value == '.' || value == ',': // Если это цифра
			b += string(value)
		case isSign(value) || value == 's': // Если это знак
			if resflag {
				switch c {
				case '+':
					if b == "" {
						return 0, ErrorInvalidExpr
					}
					uuid := uuid.New()
					id := uuid.ID()
					t := types.Task{
						Arg1:          res,
						Arg2:          convertString(b),
						Operation:     "+",
						Status:        consts.WaitStatus,
						OperationTime: env.TIME_ADDITION_MS,
						Done:          make(chan struct{}),
					}

					tasks.Add(id, &t) // Записываем задачу
					<-t.Done
					res = t.Result
				case '-':
					if b == "" {
						return 0, ErrorInvalidExpr
					}
					uuid := uuid.New()
					id := uuid.ID()
					t := types.Task{
						Arg1:          res,
						Arg2:          convertString(b),
						Operation:     "-",
						Status:        consts.WaitStatus,
						OperationTime: env.TIME_SUBTRACTION_MS,
						Done:          make(chan struct{}),
					}

					tasks.Add(id, &t) // Записываем задачу
					<-t.Done
					res = t.Result
				case '*':
					if b == "" {
						return 0, ErrorInvalidExpr
					}
					uuid := uuid.New()
					id := uuid.ID()
					t := types.Task{
						Arg1:          res,
						Arg2:          convertString(b),
						Operation:     "*",
						Status:        consts.WaitStatus,
						OperationTime: env.TIME_MULTIPLICATIONS_MS,
						Done:          make(chan struct{}),
					}

					tasks.Add(id, &t) // Записываем задачу
					<-t.Done
					res = t.Result
				case '/':
					if b == "" {
						return 0, ErrorInvalidExpr
					}
					uuid := uuid.New()
					id := uuid.ID()
					arg2 := convertString(b)
					if arg2 == 0 {
						return 0, ErrorDivByZero
					}
					t := types.Task{
						Arg1:          res,
						Arg2:          arg2,
						Operation:     "/",
						Status:        consts.WaitStatus,
						OperationTime: env.TIME_DIVISIONS_MS,
						Done:          make(chan struct{}),
					}

					tasks.Add(id, &t) // Записываем задачу
					<-t.Done
					res = t.Result
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
			return 0, ErrorInvalidExpr // Неизвестный символ
		}
	}
	return res, nil
}
