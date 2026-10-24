package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/josesojo2828-tech/docker-tui/internal/docker"
	"github.com/josesojo2828-tech/docker-tui/internal/tui"
	"github.com/josesojo2828-tech/docker-tui/internal/web"
)

func main() {
	port := flag.Int("port", 8080, "Puerto para el servidor web dashboard")
	noWeb := flag.Bool("no-web", false, "Desactivar servidor web dashboard")
	webOnly := flag.Bool("web-only", false, "Ejecutar únicamente el dashboard web (sin interfaz TUI)")
	flag.Parse()

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
