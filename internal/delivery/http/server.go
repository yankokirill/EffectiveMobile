package http

import (
	"context"
	"github.com/yankokirill/song-library/internal/repository/postgres"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	db      postgres.SongRepository
	address string
}

func NewServer(db postgres.SongRepository, address string) *Server {
	return &Server{db, address}
}

func (s *Server) Run() {
	log.Printf("Starting server at %s\n", s.address)
	if err := http.ListenAndServe(s.address, s.Routes()); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}
}

func (s *Server) Shutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	server := http.Server{Addr: s.address}
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP shutdown error: %v", err)
	}
	log.Println("Graceful shutdown complete.")
}
