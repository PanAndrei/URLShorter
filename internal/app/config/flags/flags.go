package flags

import (
	hadlCnfg "URLShorter/internal/app/handlers/config"
	"flag"
)

type mainConfig struct {
	Handlers hadlCnfg.Config
}

var Cnfg = mainConfig{
	Handlers: hadlCnfg.Config{},
}

func ParsFlags() {
	flag.StringVar(&Cnfg.Handlers.ServerAdress, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&Cnfg.Handlers.ReturnAdress, "b", "http://localhost:8080", "redirect adress")
	flag.Parse()
}
