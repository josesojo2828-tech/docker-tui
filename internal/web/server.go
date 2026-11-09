package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/josesojo2828-tech/docker-tui/internal/docker"
)

//go:embed static/*
var staticFS embed.FS

// Server represents the web dashboard server
type Server struct {
	dockerSvc  *docker.Service
	httpServer *http.Server
	addr       string
}

// NewServer creates a new web dashboard instance
func NewServer(port int, dockerSvc *docker.Service) *Server {
	addr := fmt.Sprintf(":%d", port)
	s := &Server{
		dockerSvc: dockerSvc,
		addr:      addr,
	}

	mux := http.NewServeMux()

	// Static assets and pages
	sub, err := fs.Sub(staticFS, "static")
	if err == nil {
		fileServer := http.FileServer(http.FS(sub))
		mux.Handle("/static/", http.StripPrefix("/static/", fileServer))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			switch path {
			case "/", "/index.html", "/containers":
				serveStaticFile(w, sub, "stitch_containers.html")
			case "/images":
				serveStaticFile(w, sub, "stitch_images.html")
			case "/volumes":
				serveStaticFile(w, sub, "stitch_volumes.html")
			default:
				fileServer.ServeHTTP(w, r)
			}
		})
	}

	// API Routes
	mux.HandleFunc("/api/summary", s.handleSummary)
	mux.HandleFunc("/api/containers", s.handleContainers)
	mux.HandleFunc("/api/containers/", s.handleContainerAction)
	mux.HandleFunc("/api/images", s.handleImages)
	mux.HandleFunc("/api/volumes", s.handleVolumes)
	mux.HandleFunc("/api/networks", s.handleNetworks)

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      corsMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return s
}

// Start begins listening in background
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the web server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// Addr returns the listening address
func (s *Server) Addr() string {
	return s.addr
}

func serveStaticFile(w http.ResponseWriter, fs fs.FS, filename string) {
	data, err := fs.Open(filename)
	if err != nil {
		http.NotFound(w, nil)
		return
	}
	defer data.Close()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	stat, _ := data.Stat()
	if stat != nil {
		http.ServeContent(w, &http.Request{}, filename, stat.ModTime(), data.(io.ReadSeeker))
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (s *Server) handleSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	summary, err := s.dockerSvc.GetSummary(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (s *Server) handleContainers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	all := r.URL.Query().Get("all") != "false"
	containers, err := s.dockerSvc.ListContainers(r.Context(), all)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, containers)
}

func (s *Server) handleContainerAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// path format: /api/containers/{id}/{action}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid action path"})
		return
	}
	containerID := parts[2]
	action := parts[3]

	var err error
	switch action {
	case "start":
		err = s.dockerSvc.StartContainer(r.Context(), containerID)
	case "stop":
		err = s.dockerSvc.StopContainer(r.Context(), containerID)
	case "restart":
		err = s.dockerSvc.RestartContainer(r.Context(), containerID)
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported action: " + action})
		return
	}

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "action": action, "id": containerID})
}

func (s *Server) handleImages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	images, err := s.dockerSvc.ListImages(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, images)
}

func (s *Server) handleVolumes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	volumes, err := s.dockerSvc.ListVolumes(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, volumes)
}

func (s *Server) handleNetworks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	networks, err := s.dockerSvc.ListNetworks(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, networks)
}
