# 🐳 Docker TUI & Web Dashboard

Una herramienta en Go moderna, rápida y ligera para monitorizar y gestionar contenedores Docker tanto desde la **terminal (TUI)** como desde una **interfaz web reactiva**.

---

## 🚀 Características

- **Terminal UI (TUI):**
  - Desarrollado con el ecosistema de [Bubble Tea](https://github.com/charmbracelet/bubbletea) y [Lip Gloss](https://github.com/charmbracelet/lipgloss).
  - Pestañas dinámicas: **Contenedores**, **Volúmenes** y **Redes**.
  - Tarjetas de resumen métrico: Contenedores activos, detenidos, total de imágenes, volúmenes y redes.
  - Gestión rápida por teclado: Iniciar (`s`), Detener (`x`), Reiniciar (`r`), Alternar filtro activos/todos (`a`), Refrescar (`R`).
  - Auto-actualización periódica en tiempo real.

- **Dashboard Web Integrado:**
  - Servidor HTTP ligero embebido con `embed.FS` (sin dependencias adicionales).
  - Vista limpia con diseño responsivo oscuro basado en Tailwind CSS.
  - Endpoints REST JSON listos para desacoplar o personalizar tus propias vistas frontend:
    - `GET /api/summary`: Métricas globales del daemon.
    - `GET /api/containers?all=true`: Listado detallado de contenedores.
    - `POST /api/containers/{id}/start`: Iniciar contenedor.
    - `POST /api/containers/{id}/stop`: Detener contenedor.
    - `POST /api/containers/{id}/restart`: Reiniciar contenedor.
    - `GET /api/volumes`: Listado de volúmenes de almacenamiento.
    - `GET /api/networks`: Listado de redes y contenedores enlazados.

---

## 📦 Estructura del Proyecto

```
docker-tui/
├── cmd/
│   └── docker-tui/
│       └── main.go           # Punto de entrada y gestión de flags CLI
├── internal/
│   ├── docker/
│   │   ├── client.go         # SDK de Docker (Listados, Acciones, Ping, Engine)
│   │   └── models.go         # Modelos de datos compartidos (JSON)
│   ├── tui/
│   │   ├── app.go            # Modelo Bubble Tea, tablas y navegación por teclado
│   │   └── styles.go         # Temas, estilos Lip Gloss y paleta de colores
│   └── web/
│       ├── server.go         # Servidor HTTP & API REST
│       └── static/
│           └── index.html    # Dashboard web embebido
├── Makefile                  # Comandos rápidos de compilación y ejecución
├── go.mod
└── go.sum
```

---

## 🛠️ Instalación y Requisitos

- **Go 1.22+**
- **Docker Engine** en ejecución local y permisos de acceso al socket Docker (`/var/run/docker.sock`).

---

## ⚡ Uso y Ejecución

### 1. Compilar y ejecutar todo (TUI + Servidor Web en paralelo)
Por defecto, al iniciar la aplicación en la terminal, se levanta la TUI y automáticamente el servidor web en `http://localhost:8080`:

```bash
make run
```
o directamente con `go run`:
```bash
go run ./cmd/docker-tui
```

### 2. Opciones de línea de comandos (Flags)

```bash
# Cambiar el puerto del dashboard web
./bin/docker-tui -port 9090

# Ejecutar únicamente el Dashboard Web (modo headless / servidor de fondo)
./bin/docker-tui -web-only -port 8080

# Ejecutar únicamente la TUI en terminal (sin servidor web)
./bin/docker-tui -no-web
```

---

## ⌨️ Controles de Teclado (TUI)

| Tecla | Acción |
|---|---|
| `↑` / `k` | Subir selección en la tabla |
| `↓` / `j` | Bajar selección en la tabla |
| `Tab` / `Shift+Tab` | Cambiar entre pestañas (Contenedores, Volúmenes, Redes) |
| `1`, `2`, `3` | Ir directo a la pestaña correspondiente |
| `s` | **Iniciar** el contenedor seleccionado |
| `x` | **Detener** el contenedor seleccionado |
| `r` | **Reiniciar** el contenedor seleccionado |
| `a` | Alternar entre ver **todos** los contenedores o sólo los activos |
| `R` | Forzar recarga inmediata de datos |
| `q` / `Ctrl+C` | Salir de la aplicación |

---

## 🔮 Próximas Mejoras (Roadmap)
- Monitorización de recursos en tiempo real (CPU %, Memoria RAM, I/O de red con Docker Stats stream).
- Visualización de logs del contenedor seleccionado con streaming (`docker logs -f`).
- Soporte para creación y eliminación de volúmenes y redes.
- Inspección detallada JSON de cada recurso.
