package flags

import (
	hadlCnfg "URLShorter/internal/app/handlers/config"
	"flag"
	"fmt"
	"os"
)

const (
	DBhost     = "localhost"
	DBport     = 5432
	DBuser     = "postgres"
	DBpassword = ""
	DBdbname   = "short"
)

type mainConfig struct {
	Handlers hadlCnfg.Config
}

var Cnfg = mainConfig{
	Handlers: hadlCnfg.Config{},
}

func ParsFlags() {
	ps := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		DBhost, DBport, DBuser, DBpassword, DBdbname)

	flag.StringVar(&Cnfg.Handlers.ServerAdress, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&Cnfg.Handlers.ReturnAdress, "b", "http://localhost:8080", "redirect adress")
	flag.StringVar(&Cnfg.Handlers.FileStorageAdress, "f", "repository.json", "local file url's storage")
	flag.StringVar(&Cnfg.Handlers.PostgreSQLAdress, "d", ps, "SQL base adress")
	flag.Parse()

	if serverAdress := os.Getenv("SERVER_ADDRESS"); serverAdress != "" {
		Cnfg.Handlers.ServerAdress = serverAdress
	}

	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		Cnfg.Handlers.ReturnAdress = baseURL
	}

	if fileAdress := os.Getenv("FILE_STORAGE_PATH"); fileAdress != "" {
		Cnfg.Handlers.FileStorageAdress = fileAdress
	}

	if posgresqlAdress := os.Getenv("DATABASE_DSN"); posgresqlAdress != "" {
		Cnfg.Handlers.PostgreSQLAdress = posgresqlAdress
	}
}
