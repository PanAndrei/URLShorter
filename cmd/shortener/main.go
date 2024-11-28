package main

import (
	"log"

	cnfg "URLShorter/internal/app/config"
	handlers "URLShorter/internal/app/handlers"
	repo "URLShorter/internal/app/repository"
	serv "URLShorter/internal/app/service"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cng := cnfg.GetConfig()
	repo := repo.NewStore()
	shorter := serv.NewShorter(repo)

	return handlers.Serve(cng.Handlers, *shorter)
}
