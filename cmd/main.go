package main

import (
	app "github.com/romanSPB15/Calculator_Service/internal/application"
)

func main() {
	a := app.New()
	a.RunServer() // Запускаем приложение
}
