package application

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/romanSPB15/Calculator_Service/pckg/dir"
)

type config struct {
	Debug bool `json:"debug"`
	Web   bool `json:"web"`
}

func newConfig() *config {
	res := new(config)
	cf, err := os.Open(dir.JsonFile())
	if err != nil {
		panic("cannot open config file")
	}
	decoder := json.NewDecoder(cf)
	err = decoder.Decode(res)
	if err != nil {
		panic(fmt.Errorf("cannot decode config file: %v", err))
	}
	return res
}
