package main

import (
	"flag"
	"io/fs"
	"log"
	mediaserver "mediaserver"
	"mediaserver/internal/config"
	"mediaserver/internal/media"
	"mediaserver/internal/server"
	"net/http"
	"os"
)

func main() {
	debugMode := flag.Bool("debug", false, "Run in backend-only debug mode (omits frontend assets)")
	flag.Parse()

	media.RegisterMIMETypes()

	configPath := os.Getenv("CONFIG_FILE")
	if configPath == "" {
		configPath = "config.json"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("failed to read %s: %v", configPath, err)
	}

	libs := media.BuildLibraries(cfg.Libraries)
	if len(libs) == 0 {
		log.Printf("warning: no valid libraries configured — edit %s and restart", configPath)
	}

	var webRoot fs.FS
	if *debugMode {
		log.Println("[DEBUG] Running in debug mode: web frontend routes disabled")
	} else {
		var err error
		webRoot, err = fs.Sub(mediaserver.WebFS, "web-static")
		if err != nil {
			log.Fatal(err)
		}
	}

	srv := server.New(cfg, libs, webRoot)
	addr := cfg.ResolveAddr()

	log.Printf("listening on: http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, srv.Handler()))
}
