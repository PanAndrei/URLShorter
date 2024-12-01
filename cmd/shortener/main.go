package main

import (
	flags "URLShorter/internal/app/config/flags"
	handlers "URLShorter/internal/app/handlers"
	repo "URLShorter/internal/app/repository"
	serv "URLShorter/internal/app/service"
	"fmt"
	"log"
)

func main() {
	flags.ParsFlags()
	fmt.Println("Running server on", flags.Cnfg.Handlers.ServerAdress)
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	repo := repo.NewStore()
	shorter := serv.NewShorter(repo)

	return handlers.Serve(flags.Cnfg.Handlers, shorter)
}
