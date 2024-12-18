package main

import (
	flags "URLShorter/internal/app/config/flags"
	handlers "URLShorter/internal/app/handlers"
	log "URLShorter/internal/app/logger"
	repo "URLShorter/internal/app/repository"
	serv "URLShorter/internal/app/service"

	"fmt"
	l "log"
)

func main() {
	flags.ParsFlags()

	if err := run(); err != nil {
		l.Fatal(fmt.Println("Error launching server"))
	}
}

func run() error {
	if err := log.Initialize("info"); err != nil {
		return err
	}

	// repo := repo.NewStore()
	repo, _ := repo.NewFileStore(flags.Cnfg.Handlers.FileStorageAdress)
	shorter := serv.NewShorter(repo)

	return handlers.Serve(flags.Cnfg.Handlers, shorter)
}
