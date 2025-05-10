package main

import (
	app "github.com/romanSPB15/Calculator_Service_Final/internal/application"
)

func main() {
	a := app.New()
	a.Run() // Запускаем приложение
}
