package docker

import "time"

// ContainerItem represents a sanitized, simplified container for TUI & Web
type ContainerItem struct {
	ID        string    `json:"id"`
	FullID    string    `json:"full_id"`
	Name      string    `json:"name"`
	Image     string    `json:"image"`
	State     string    `json:"state"`   // running, exited, paused, restarting, etc.
	Status    string    `json:"status"`  // "Up 4 hours", "Exited (0) 2 hours ago"
	Created   time.Time `json:"created"`
	Ports     string    `json:"ports"`
	IPAddress string    `json:"ip_address,omitempty"`
}

// VolumeItem represents a Docker volume
type VolumeItem struct {
	Name       string `json:"name"`
	Driver     string `json:"driver"`
	Scope      string `json:"scope"`
	Mountpoint string `json:"mountpoint"`
	CreatedAt  string `json:"created_at"`
}

// NetworkItem represents a Docker network
type NetworkItem struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Driver     string `json:"driver"`
	Scope      string `json:"scope"`
	Containers int    `json:"containers_count"`
	Internal   bool   `json:"internal"`
}

// SystemSummary represents the high-level dashboard metrics
type SystemSummary struct {
	TotalContainers   int    `json:"total_containers"`
	RunningContainers int    `json:"running_containers"`
	PausedContainers  int    `json:"paused_containers"`
	StoppedContainers int    `json:"stopped_containers"`
	TotalImages       int    `json:"total_images"`
	TotalVolumes      int    `json:"total_volumes"`
	TotalNetworks     int    `json:"total_networks"`
	DockerVersion     string `json:"docker_version"`
	ServerTime        string `json:"server_time"`
}
