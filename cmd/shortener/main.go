package main

import (
	flags "URLShorter/internal/app/config/flags"
	handlers "URLShorter/internal/app/handlers"
	log "URLShorter/internal/app/logger"
	repo "URLShorter/internal/app/repository"
	serv "URLShorter/internal/app/service"
)

func main() {
	flags.ParsFlags()

	if err := run(); err != nil {
		// log fatal
	}
}

func run() error {
	if err := log.Initialize("info"); err != nil {
		return err
	}

	repo := repo.NewStore()
	shorter := serv.NewShorter(repo)

	return handlers.Serve(flags.Cnfg.Handlers, shorter)
}
