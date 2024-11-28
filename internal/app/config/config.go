package config

import (
	hadlCnfg "URLShorter/internal/app/handlers/config"
)

const (
	LocalHost = "http://localhost:8080/"
	Port      = "localhost:8080"
)

type Config struct {
	Handlers hadlCnfg.Config
}

func GetConfig() Config {
	cfg := Config{
		Handlers: hadlCnfg.Config{
			ServerAdress: Port,
		},
	}

	return cfg
}
