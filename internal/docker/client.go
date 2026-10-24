package docker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

type Service struct {
	cli *client.Client
}

// NewService creates a new Docker client connected to the local Docker daemon
func NewService() (*Service, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	// Verify connection by pinging the daemon
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err = cli.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("docker daemon is unreachable: %w", err)
	}

	return &Service{cli: cli}, nil
}

// Close closes the Docker client connection
func (s *Service) Close() error {
	if s.cli != nil {
		return s.cli.Close()
	}
	return nil
}

// GetSummary returns high-level metrics for dashboard
func (s *Service) GetSummary(ctx context.Context) (*SystemSummary, error) {
	containers, err := s.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("error listing containers: %w", err)
	}

	running := 0
	paused := 0
	stopped := 0
	for _, c := range containers {
		switch strings.ToLower(c.State) {
		case "running":
			running++
		case "paused":
			paused++
		default:
			stopped++
		}
	}

	vols, err := s.cli.VolumeList(ctx, volume.ListOptions{})
	volCount := 0
	if err == nil {
		volCount = len(vols.Volumes)
	}

	nets, err := s.cli.NetworkList(ctx, network.ListOptions{})
	netCount := 0
	if err == nil {
		netCount = len(nets)
	}

	images, err := s.cli.ImageList(ctx, image.ListOptions{})
	imgCount := 0
	if err == nil {
		imgCount = len(images)
	}

	versionInfo, err := s.cli.ServerVersion(ctx)
	dVer := "unknown"
	if err == nil {
		dVer = versionInfo.Version
	}

	return &SystemSummary{
		TotalContainers:   len(containers),
		RunningContainers: running,
		PausedContainers:  paused,
		StoppedContainers: stopped,
		TotalImages:       imgCount,
		TotalVolumes:      volCount,
		TotalNetworks:     netCount,
		DockerVersion:     dVer,
		ServerTime:        time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// ListContainers returns simplified container items
func (s *Service) ListContainers(ctx context.Context, all bool) ([]ContainerItem, error) {
	rawContainers, err := s.cli.ContainerList(ctx, container.ListOptions{All: all})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	items := make([]ContainerItem, 0, len(rawContainers))
	for _, c := range rawContainers {
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		shortID := c.ID
		if len(shortID) > 12 {
			shortID = shortID[:12]
		}

		var portStrings []string
		for _, p := range c.Ports {
			if p.PublicPort > 0 {
				portStrings = append(portStrings, fmt.Sprintf("%d->%d/%s", p.PublicPort, p.PrivatePort, p.Type))
			} else if p.PrivatePort > 0 {
				portStrings = append(portStrings, fmt.Sprintf("%d/%s", p.PrivatePort, p.Type))
			}
		}

		ports := strings.Join(portStrings, ", ")
		if ports == "" {
			ports = "-"
		}

		items = append(items, ContainerItem{
			ID:      shortID,
			FullID:  c.ID,
			Name:    name,
			Image:   c.Image,
			State:   c.State,
			Status:  c.Status,
			Created: time.Unix(c.Created, 0),
			Ports:   ports,
		})
	}

	return items, nil
}

// ListVolumes returns all volumes
func (s *Service) ListVolumes(ctx context.Context) ([]VolumeItem, error) {
	resp, err := s.cli.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %w", err)
	}

	items := make([]VolumeItem, 0, len(resp.Volumes))
	for _, v := range resp.Volumes {
		createdAt := v.CreatedAt
		if createdAt == "" {
			createdAt = "-"
		}
		items = append(items, VolumeItem{
			Name:       v.Name,
			Driver:     v.Driver,
			Scope:      v.Scope,
			Mountpoint: v.Mountpoint,
			CreatedAt:  createdAt,
		})
	}
	return items, nil
}

// ListNetworks returns all networks
func (s *Service) ListNetworks(ctx context.Context) ([]NetworkItem, error) {
	nets, err := s.cli.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list networks: %w", err)
	}

	items := make([]NetworkItem, 0, len(nets))
	for _, n := range nets {
		shortID := n.ID
		if len(shortID) > 12 {
			shortID = shortID[:12]
		}
		items = append(items, NetworkItem{
			ID:         shortID,
			Name:       n.Name,
			Driver:     n.Driver,
			Scope:      n.Scope,
			Containers: len(n.Containers),
			Internal:   n.Internal,
		})
	}
	return items, nil
}

// StartContainer starts a container by ID
func (s *Service) StartContainer(ctx context.Context, id string) error {
	return s.cli.ContainerStart(ctx, id, container.StartOptions{})
}

// StopContainer stops a container by ID
func (s *Service) StopContainer(ctx context.Context, id string) error {
	return s.cli.ContainerStop(ctx, id, container.StopOptions{})
}

// RestartContainer restarts a container by ID
func (s *Service) RestartContainer(ctx context.Context, id string) error {
	return s.cli.ContainerRestart(ctx, id, container.StopOptions{})
}
