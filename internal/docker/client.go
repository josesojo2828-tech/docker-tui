package docker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
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

// FormatBytes formats byte sizes into human readable units
func FormatBytes(b int64) string {
	if b < 0 {
		return "N/A"
	}
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// ListImages returns images with tags, dates and formatted sizes
func (s *Service) ListImages(ctx context.Context) ([]ImageItem, error) {
	rawImages, err := s.cli.ImageList(ctx, image.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	items := make([]ImageItem, 0, len(rawImages))
	for _, img := range rawImages {
		shortID := img.ID
		if strings.HasPrefix(shortID, "sha256:") {
			shortID = strings.TrimPrefix(shortID, "sha256:")
		}
		if len(shortID) > 12 {
			shortID = shortID[:12]
		}

		repo := "<none>"
		tag := "<none>"
		if len(img.RepoTags) > 0 && img.RepoTags[0] != "<none>:<none>" {
			parts := strings.Split(img.RepoTags[0], ":")
			if len(parts) == 2 {
				repo = parts[0]
				tag = parts[1]
			} else {
				repo = img.RepoTags[0]
			}
		}

		createdTime := time.Unix(img.Created, 0)
		items = append(items, ImageItem{
			ID:            shortID,
			FullID:        img.ID,
			Repository:    repo,
			Tag:           tag,
			Created:       createdTime,
			CreatedStr:    createdTime.Format("2006-01-02 15:04"),
			SizeBytes:     img.Size,
			SizeFormatted: FormatBytes(img.Size),
			Containers:    img.Containers,
		})
	}
	return items, nil
}

// ListVolumes returns all volumes with usage data (size and ref counts)
func (s *Service) ListVolumes(ctx context.Context) ([]VolumeItem, error) {
	// Try DiskUsage first to fetch actual size and references
	du, err := s.cli.DiskUsage(ctx, types.DiskUsageOptions{})
	if err == nil && len(du.Volumes) > 0 {
		items := make([]VolumeItem, 0, len(du.Volumes))
		for _, v := range du.Volumes {
			sizeBytes := int64(-1)
			refCount := int64(0)
			if v.UsageData != nil {
				sizeBytes = v.UsageData.Size
				refCount = v.UsageData.RefCount
			}

			createdAt := v.CreatedAt
			if createdAt == "" {
				createdAt = "-"
			} else if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
				createdAt = t.Format("2006-01-02 15:04")
			}

			items = append(items, VolumeItem{
				Name:          v.Name,
				Driver:        v.Driver,
				Scope:         v.Scope,
				Mountpoint:    v.Mountpoint,
				CreatedAt:     createdAt,
				SizeBytes:     sizeBytes,
				SizeFormatted: FormatBytes(sizeBytes),
				RefCount:      refCount,
			})
		}
		return items, nil
	}

	// Fallback to standard VolumeList
	resp, err := s.cli.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %w", err)
	}

	items := make([]VolumeItem, 0, len(resp.Volumes))
	for _, v := range resp.Volumes {
		createdAt := v.CreatedAt
		if createdAt == "" {
			createdAt = "-"
		} else if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
			createdAt = t.Format("2006-01-02 15:04")
		}
		items = append(items, VolumeItem{
			Name:          v.Name,
			Driver:        v.Driver,
			Scope:         v.Scope,
			Mountpoint:    v.Mountpoint,
			CreatedAt:     createdAt,
			SizeBytes:     -1,
			SizeFormatted: "N/A",
			RefCount:      0,
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
