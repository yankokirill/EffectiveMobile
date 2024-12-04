package main

import (
	"github.com/yankokirill/song-library/config"
	"github.com/yankokirill/song-library/internal/delivery/http"
	"github.com/yankokirill/song-library/internal/repository/migrations"
	"github.com/yankokirill/song-library/internal/repository/postgres"
	"github.com/yankokirill/song-library/internal/rpc"
	"log"
)

// @title Song Library API
// @version 1.0
// @description This is the API documentation for managing songs in a library.
// @host localhost:8080
// @BasePath /library
// @schemes http
func main() {
	config.Load()

	if err := migrations.Up("file://migrations", config.DatabaseURL()); err != nil {
		log.Fatalf("%v", err)
	}
	repo, err := postgres.NewSongRepository(config.DatabaseURL())
	if err != nil {
		log.Fatalf("failed to start postgresql database: %v", err)
	}
	defer repo.Close()

	rpc.SetExternalApiURL(config.ExternalApiURL())
	server := http.NewServer(repo, config.ServerAddress())
	go server.Run()
	server.Shutdown()
}
