package main

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx := context.Background()
	client, _ := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	defer client.Close()

	certCache := client.CacheVolume("node")
	dockerState := client.CacheVolume("docker-state")

	docker, _ := client.Container().
		From("docker:dind").
		WithExposedPort(2376).
		WithMountedCache("/var/lib/docker", dockerState, dagger.ContainerWithMountedCacheOpts{
			Sharing: dagger.CacheSharingModePrivate,
		}).
		WithMountedCache("/certs", certCache).
		WithExec(nil, dagger.ContainerWithExecOpts{
			InsecureRootCapabilities: true,
		}).
		AsService().
		Start(ctx)

	runner := client.Container().
		From("docker:latest").
		WithServiceBinding("docker", docker).
		WithMountedCache("/certs", certCache).
		WithEnvVariable("DOCKER_HOST", "tcp://docker:2376").
		WithEnvVariable("DOCKER_TLS_CERTDIR", "/certs").
		WithEnvVariable("DOCKER_CERT_PATH", "/certs/client").
		WithEnvVariable("DOCKER_TLS_VERIFY", "1")

	group := errgroup.Group{}

	// Execute two tasks to pull busybox and alpine images
	group.Go(func() error {
		_, _ = runner.
			WithExec([]string{"docker", "pull", "busybox"}).
			Sync(ctx)
		return nil
	})
	group.Go(func() error {
		_, _ = runner.
			WithExec([]string{"docker", "pull", "alpine"}).
			Sync(ctx)
		return nil
	})

	// For for first task group to complete
	if err := group.Wait(); err != nil {
		fmt.Printf("errgroup tasks ended up with an error: %v\n", err)
	} else {
		fmt.Println("all works done successfully")
	}

	// List available images to verify the images were pulled and are available
	_, _ = runner.
		WithExec([]string{"docker", "images"}).
		Sync(ctx)
}
