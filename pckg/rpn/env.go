package rpn

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var (
	TIME_ADDITION_MS        int
	TIME_SUBTRACTION_MS     int
	TIME_MULTIPLICATIONS_MS int
	TIME_DIVISIONS_MS       int
	COMPUTING_POWER         int
)

// Считывание переменной среды в виде числа
func getIntEnv(key string) int {
	str, has := os.LookupEnv(key)
	if !has {
		log.Panicf("System has not %s", key)
	}
	res, err := strconv.Atoi(str)
	if err != nil {
		log.Panicf("Env %s is not int", key)
	}
	return res
}

// Иницилизация переменных Go из файла .env
func InitEnv(file ...string) {
	err := godotenv.Load(file...)
	if err != nil {
		panic(err)
	}
	TIME_ADDITION_MS = getIntEnv("TIME_ADDITION_MS")
	TIME_SUBTRACTION_MS = getIntEnv("TIME_SUBTRACTION_MS")
	TIME_MULTIPLICATIONS_MS = getIntEnv("TIME_MULTIPLICATIONS_MS")
	TIME_DIVISIONS_MS = getIntEnv("TIME_DIVISIONS_MS")
	COMPUTING_POWER = getIntEnv("COMPUTING_POWER")
}
