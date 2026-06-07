package runner

import (
	"context"
	"fmt"
	"net"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type Spawner struct {
	BenchNetName string
	docker       *client.Client
}

func NewSpawner(benchNetName string) *Spawner {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(fmt.Sprintf("spawner: docker client init failed: %v", err))
	}
	return &Spawner{BenchNetName: benchNetName, docker: cli}
}

func (s *Spawner) Spawn(imageTag, submissionID string) (string, int, error) {
	if imageTag == "" || submissionID == "" {
		return "", 0, fmt.Errorf("spawner: image tag and submission id are required")
	}

	port, err := freePort()
	if err != nil {
		return "", 0, fmt.Errorf("spawner: find free port: %w", err)
	}
	hostPort := fmt.Sprintf("%d", port)

	ctx := context.Background()
	resp, err := s.docker.ContainerCreate(ctx,
		&container.Config{
			Image: imageTag,
			ExposedPorts: nat.PortSet{
				"8080/tcp": struct{}{},
			},
		},
		&container.HostConfig{
			PortBindings: nat.PortMap{
				"8080/tcp": []nat.PortBinding{{HostIP: "127.0.0.1", HostPort: hostPort}},
			},
			NetworkMode:    container.NetworkMode(s.BenchNetName),
			ReadonlyRootfs: true,
			SecurityOpt:    []string{"no-new-privileges"},
			Resources: container.Resources{
				Memory:   512 * 1024 * 1024, // 512 MB
				CPUQuota: 100000,             // 1 CPU
			},
			CapDrop: []string{"ALL"},
		},
		&network.NetworkingConfig{},
		nil,
		"bench-"+submissionID,
	)
	if err != nil {
		return "", 0, fmt.Errorf("spawner: container create: %w", err)
	}

	if err := s.docker.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", 0, fmt.Errorf("spawner: container start: %w", err)
	}

	return resp.ID, port, nil
}

func freePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
