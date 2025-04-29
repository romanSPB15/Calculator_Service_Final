package env

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type List struct {
	TIME_ADDITION_MS        int
	TIME_SUBTRACTION_MS     int
	TIME_MULTIPLICATIONS_MS int
	TIME_DIVISIONS_MS       int
	COMPUTING_POWER         int
	PORT                    int
	HOST                    string
}

func NewList() *List {
	return &List{}
}

// Считывание переменной среды в виде числа
func getIntEnv(key string) (int, error) {
	str, has := os.LookupEnv(key)
	if !has {
		return 0, fmt.Errorf("not found")
	}
	res, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}
	return res, nil
}

// Считывание переменной среды в виде числа
func getStringEnv(key string) (string, error) {
	str, has := os.LookupEnv(key)
	if !has {
		return "", fmt.Errorf("not found")
	}
	return str, nil
}

// Иницилизация переменных Go из файла .env
func (e *List) InitEnv(file ...string) error {
	err := godotenv.Load(file...)
	if err != nil {
		panic(err)
	}
	e.TIME_ADDITION_MS, err = getIntEnv("TIME_ADDITION_MS")
	if err != nil {
		return err
	}
	e.TIME_SUBTRACTION_MS, err = getIntEnv("TIME_SUBTRACTION_MS")
	if err != nil {
		return err
	}
	e.TIME_MULTIPLICATIONS_MS, err = getIntEnv("TIME_MULTIPLICATIONS_MS")
	if err != nil {
		return err
	}
	e.TIME_DIVISIONS_MS, err = getIntEnv("TIME_DIVISIONS_MS")
	if err != nil {
		return err
	}
	e.COMPUTING_POWER, err = getIntEnv("COMPUTING_POWER")
	if err != nil {
		return err
	}
	e.PORT, err = getIntEnv("PORT")
	if err != nil {
		return err
	}
	e.HOST, err = getStringEnv("HOST")
	return err
}
