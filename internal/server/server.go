package server

import (
	"io/fs"
	"log"
	"net/http"

	"mediaserver/internal/config"
	"mediaserver/internal/media"
)

type Server struct {
	cfg       *config.Config
	libraries map[string]media.Library
	webFS     fs.FS
	mux       *http.ServeMux
}

func New(cfg *config.Config, libraries map[string]media.Library, webFS fs.FS) *Server {
	s := &Server{
		cfg:       cfg,
		libraries: libraries,
		webFS:     webFS,
		mux:       http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.Handle("/", http.FileServer(http.FS(s.webFS)))
	s.mux.HandleFunc("/api/libraries", s.handleListLibraries)
	s.mux.HandleFunc("/api/browse", s.handleBrowse)
	s.mux.HandleFunc("/media/", s.handleStreamMedia)
}

func (s *Server) Handler() http.Handler {
	var handler http.Handler = s.mux
	if s.cfg.Username != "" || s.cfg.Password != "" {
		handler = basicAuthMiddleware(s.cfg.Username, s.cfg.Password)(s.mux)
		log.Printf("basic auth enabled (username: %q)", s.cfg.Username)
	} else {
		log.Printf("warning: no username/password set in config — running without authentication")
	}
	return handler
}
