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
	if s.webFS != nil {
		s.mux.Handle("/", http.FileServer(http.FS(s.webFS)))
	} else {
		log.Println("[DEBUG] webFS is nil; frontend serving disabled")
	}
	s.mux.HandleFunc("/api/libraries", s.handleListLibraries)
	s.mux.HandleFunc("/api/browse", s.handleBrowse)
	s.mux.HandleFunc("/media/", s.handleStreamMedia)
}

func (s *Server) Handler() http.Handler {
	var handler http.Handler = s.mux
	if s.webFS == nil {
		handler = corsMiddleware(handler)
	}
	if s.cfg.Username != "" || s.cfg.Password != "" {
		handler = basicAuthMiddleware(s.cfg.Username, s.cfg.Password)(s.mux)
		log.Printf("basic auth enabled (username: %q)", s.cfg.Username)
	} else {
		log.Printf("warning: no username/password set in config — running without authentication")
	}
	return handler
}

// corsMiddleware allows cross-origin requests from the Vite dev server
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle browser pre-flight OPTIONS request directly
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
