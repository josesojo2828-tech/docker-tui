package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/josesojo2828-tech/docker-tui/internal/docker"
	"github.com/josesojo2828-tech/docker-tui/internal/tui"
	"github.com/josesojo2828-tech/docker-tui/internal/web"
)

func main() {
	mode, rest, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ %v\n\n", err)
		usage()
		os.Exit(2)
	}

	fs := flag.NewFlagSet("docker-tui", flag.ExitOnError)
	port := fs.Int("port", 8080, "Puerto para el servidor web dashboard")
	noWeb := fs.Bool("no-web", false, "Desactivar servidor web dashboard")
	webOnly := fs.Bool("web-only", false, "Ejecutar únicamente el dashboard web (sin interfaz TUI)")
	fs.Parse(rest)

	switch mode {
	case "web":
		*webOnly = true
	case "tui":
		*noWeb = true
	}

	// Initialize Docker Service
	dockerSvc, err := docker.NewService()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error al conectar con Docker Daemon:\n   %v\n\n", err)
		fmt.Fprintf(os.Stderr, "Asegúrate de que Docker esté corriendo:\n")
		fmt.Fprintf(os.Stderr, "  sudo systemctl start docker\n")
		fmt.Fprintf(os.Stderr, "  sudo usermod -aG docker $USER\n\n")
		os.Exit(1)
	}
	defer dockerSvc.Close()

	var webServer *web.Server
	webPortToReport := 0

	// Start Web Server if not disabled
	if !*noWeb {
		webPortToReport = *port
		webServer = web.NewServer(*port, dockerSvc)
		go func() {
			if err := webServer.Start(); err != nil && err != http.ErrServerClosed {
				log.Printf("Web server error: %v", err)
			}
		}()
	}

	// If web-only mode requested
	if *webOnly {
		fmt.Printf("🐳 Docker Web Dashboard corriendo en http://localhost:%d\n", *port)
		fmt.Println("Presiona Ctrl+C para salir.")

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		fmt.Println("\nDeteniendo servidor...")
		if webServer != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			_ = webServer.Shutdown(ctx)
		}
		return
	}

	// Run TUI
	model := tui.NewModel(dockerSvc, webPortToReport)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error ejecutando TUI: %v\n", err)
		os.Exit(1)
	}

	// Graceful shutdown of web server after TUI exits
	if webServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = webServer.Shutdown(ctx)
	}
}

// parseArgs resolves the optional subcommand and returns the remaining
// arguments for the flag parser. Flags may come before the subcommand so the
// previous CLI keeps working.
func parseArgs(args []string) (mode string, rest []string, err error) {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return "all", args, nil
	}

	switch args[0] {
	case "all", "tui", "web":
		return args[0], args[1:], nil
	default:
		return "", nil, fmt.Errorf("comando desconocido: %q", args[0])
	}
}

func usage() {
	fmt.Print(`docker-tui — Docker TUI & Web Dashboard

Uso:
  docker-tui [all|tui|web] [flags]

Comandos:
  all    TUI + dashboard web (por defecto)
  tui    Solo interfaz de terminal
  web    Solo dashboard web

Flags:
  -port int      Puerto del dashboard web (por defecto 8080)
  -no-web        Desactivar el dashboard web
  -web-only      Ejecutar solo el dashboard web
`)
}
