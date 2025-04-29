package dir

import (
	"os"
	"strings"
)

// Папка config
func configFiles() string {
	dir, err := os.Getwd() // Рабочая директория(.\cmd)
	if err != nil {
		panic(err)
	}
	dir, _ = strings.CutSuffix(dir, "cmd")
	dir += `\config\`
	return dir
}

// config/config.json
func JsonFile() string {
	res := configFiles() + `.json`
	return res
}

// config/.env
func EnvFile() string {
	res := configFiles() + `.env`
	return res
}
