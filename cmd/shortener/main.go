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

	storageRouter := repo.NewStorageRouter()
	db := repo.NewDB(flags.Cnfg.Handlers.PostgreSQLAdress) // temp
	repo, err := storageRouter.GetStorage(flags.Cnfg.Handlers.FileStorageAdress)

	if err != nil {
		return err
	}

	shorter := serv.NewShorter(repo)

	return handlers.Serve(flags.Cnfg.Handlers, shorter, db)
}
