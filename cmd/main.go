package main

import (
	"os"
	"os/signal"
	"syscall"

	app "github.com/romanSPB15/Calculator_Service_Final/internal/application"
)

func main() {
	a := app.New()
	go a.Run() // Запускаем приложение

	// Создаём канал для передачи информации о сигналах
	stop := make(chan os.Signal, 1)
	// "Слушаем" перечисленные сигналы
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	// Ждём данных из канала
	<-stop
	a.Stop()
}
