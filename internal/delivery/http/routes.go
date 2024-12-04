package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/yankokirill/song-library/docs"
	"net/http"
)

func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Route("/library", func(r chi.Router) {
		r.Get("/songs", s.getSongsInfoHandler)
		r.Get("/songs/{group}", s.getGroupSongsInfoHandler)
		r.Get("/song/{id}", s.getSongLyricsHandler)

		r.Post("/song", s.addSongHandler)
		r.Put("/song/{id}", s.updateSongHandler)

		r.Delete("/song/{id}", s.deleteSongHandler)
	})

	r.Get("/swagger/*", httpSwagger.WrapHandler)
	return r
}
