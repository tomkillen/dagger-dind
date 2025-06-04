package main

import (
	"context"
	"log"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

const (
	DAGGER_ENGINE_IMAGE = "registry.dagger.io/engine"
)

func main() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Panicf("failed to create docker client with error: %x", err)
	}

	// Pull dagger engine image
	_, err = cli.ImagePull(ctx, DAGGER_ENGINE_IMAGE, image.PullOptions{})
	if err != nil {
		log.Panicf("failed to pull dagger engine image with error: %x", err)
	}

	// Create dagger engine container
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: DAGGER_ENGINE_IMAGE,
	}, nil, nil, nil, "dagger-engine")
	if err != nil {
		log.Panicf("failed to create engine container with error: %x", err)
	}

	// Start dagger engine container
	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		log.Panicf("failed to start engine container with error: %x", err)
	}

	// Here we would do some work
	time.Sleep(10 * time.Second)

	// Stop dagger engine container
	if err := cli.ContainerStop(ctx, resp.ID, container.StopOptions{}); err != nil {
		log.Panicf("failed to stop engine container with error: %x", err)
	}

	// Remove the engine container
	if err := cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{}); err != nil {
		log.Panicf("failed to remove engine container with error: %x", err)
	}
}
