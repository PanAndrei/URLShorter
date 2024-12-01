package main

import (
	flags "URLShorter/internal/app/config/flags"
	handlers "URLShorter/internal/app/handlers"
	repo "URLShorter/internal/app/repository"
	serv "URLShorter/internal/app/service"
	"log"
)

func main() {
	flags.ParsFlags()

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	repo := repo.NewStore()
	shorter := serv.NewShorter(repo)

	return handlers.Serve(flags.Cnfg.Handlers, shorter)
}
