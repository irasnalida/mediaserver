package media

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"mediaserver/internal/config"
)

var AllowedExt = map[string]string{
	".mp4":  "video",
	".webm": "video",
	".mkv":  "video",
	".mov":  "video",
	".avi":  "video",
	".mp3":  "audio",
	".wav":  "audio",
	".flac": "audio",
	".ogg":  "audio",
	".m4a":  "audio",
	".jpg":  "image",
	".jpeg": "image",
	".png":  "image",
	".gif":  "image",
	".webp": "image",
	".bmp":  "image",
}

type Entry struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"`
	Size int64  `json:"size,omitempty"`
	URL  string `json:"url,omitempty"`
}

type Library struct {
	Name string
	Root string
}

func BuildLibraries(entries []config.LibraryConfig) map[string]Library {
	result := make(map[string]Library)
	seen := make(map[string]bool)

	for _, e := range entries {
		p := expandHome(strings.TrimSpace(e.Path))
		if p == "" {
			continue
		}
		abs, err := filepath.Abs(p)
		if err != nil {
			log.Printf("skipping library %q: %v", e.Name, err)
			continue
		}
		info, err := os.Stat(abs)
		if err != nil || !info.IsDir() {
			log.Printf("skipping library %q: %s is not a valid, reachable folder", e.Name, abs)
			continue
		}

		name := strings.TrimSpace(e.Name)
		if name == "" {
			name = filepath.Base(abs)
		}
		if strings.Contains(name, "/") {
			log.Printf("skipping library %q: names can't contain \"/\"", name)
			continue
		}

		key := name
		i := 2
		for seen[key] {
			key = fmt.Sprintf("%s (%d)", name, i)
			i++
		}
		seen[key] = true

		result[key] = Library{Name: key, Root: abs}
		log.Printf("library %q -> %s", key, abs)
	}

	return result
}

func (lib Library) ResolvePath(rel string) (fullPath string, cleanRel string, err error) {
	rel = filepath.FromSlash(strings.TrimPrefix(rel, "/"))
	cleaned := filepath.Clean(rel)
	if cleaned == "." {
		cleaned = ""
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(os.PathSeparator)) {
		return "", "", fmt.Errorf("invalid path")
	}

	full := filepath.Join(lib.Root, cleaned)
	if full != lib.Root && !strings.HasPrefix(full, lib.Root+string(os.PathSeparator)) {
		return "", "", fmt.Errorf("invalid path")
	}
	return full, filepath.ToSlash(cleaned), nil
}

func MediaURL(libName, relPath string) string {
	segments := strings.Split(relPath, "/")
	for i, s := range segments {
		segments[i] = url.PathEscape(s)
	}
	return "/media/" + url.PathEscape(libName) + "/" + strings.Join(segments, "/")
}

func expandHome(p string) string {
	if p == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
	}
	if strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, p[2:])
		}
	}
	return p
}
