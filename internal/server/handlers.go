package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"mediaserver/internal/media"
)

type BrowseResponse struct {
	Library string        `json:"library"`
	Dir     string        `json:"dir"`
	Entries []media.Entry `json:"entries"`
}

func (s *Server) handleListLibraries(w http.ResponseWriter, r *http.Request) {
	names := make([]string, 0, len(s.libraries))
	for name := range s.libraries {
		names = append(names, name)
	}
	sort.Strings(names)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(names)
}

func (s *Server) handleBrowse(w http.ResponseWriter, r *http.Request) {
	libName := r.URL.Query().Get("lib")
	lib, ok := s.libraries[libName]
	if !ok {
		http.Error(w, "unknown library", http.StatusNotFound)
		return
	}

	target, relClean, err := lib.ResolvePath(r.URL.Query().Get("dir"))
	if err != nil {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	info, err := os.Stat(target)
	if err != nil || !info.IsDir() {
		http.Error(w, "not a folder", http.StatusNotFound)
		return
	}

	dirEntries, err := os.ReadDir(target)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var entries []media.Entry
	for _, de := range dirEntries {
		name := de.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}

		childRel := name
		if relClean != "" {
			childRel = relClean + "/" + name
		}

		if de.IsDir() {
			entries = append(entries, media.Entry{Name: name, Path: childRel, Type: "folder"})
			continue
		}

		ext := strings.ToLower(filepath.Ext(name))
		kind, ok := media.AllowedExt[ext]
		if !ok {
			continue
		}

		var size int64
		if info, err := de.Info(); err == nil {
			size = info.Size()
		}

		entries = append(entries, media.Entry{
			Name: name,
			Path: childRel,
			Type: kind,
			Size: size,
			URL:  media.MediaURL(libName, childRel),
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		iFolder, jFolder := entries[i].Type == "folder", entries[j].Type == "folder"
		if iFolder != jFolder {
			return iFolder
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(BrowseResponse{
		Library: libName,
		Dir:     relClean,
		Entries: entries,
	})
}

func (s *Server) handleStreamMedia(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/media/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		http.NotFound(w, r)
		return
	}

	lib, ok := s.libraries[parts[0]]
	if !ok {
		http.NotFound(w, r)
		return
	}

	fullPath, _, err := lib.ResolvePath(parts[1])
	if err != nil {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	ext := strings.ToLower(filepath.Ext(fullPath))
	if _, ok := media.AllowedExt[ext]; !ok {
		http.Error(w, "unsupported file type", http.StatusForbidden)
		return
	}

	f, err := os.Open(fullPath)
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil || info.IsDir() {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	http.ServeContent(w, r, fullPath, info.ModTime(), f)
}
