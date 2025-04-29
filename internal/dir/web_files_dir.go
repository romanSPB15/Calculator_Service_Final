package dir

import (
	"os"
	"strings"
)

func templates() string { // Получаем директорию папки templates
	dir, err := os.Getwd() // Рабочая директория(.\cmd)
	if err != nil {
		panic(err)
	}
	dir, _, _ = strings.Cut(dir, "cmd")
	return dir + `templates\` // .\templates\
}

func GetTemplateFile(name string) string { // Получаем полную директорию нужного файла
	return templates() + name // .\templates\file.html
}
