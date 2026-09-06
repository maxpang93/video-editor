package ffmpegworker

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"log"
	"os"

	"github.com/moby/moby/client"
)

func streamBuildContext(dockerfile string) (io.Reader, error) {
	dockerfileContent, err := os.ReadFile(dockerfile)
	if err != nil {
		log.Printf("Failed to open dockerfile: %s\n", dockerfile)
		return nil, err
	}
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	hdr := &tar.Header{
		Name: dockerfile,
		Size: int64(len(dockerfileContent)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return nil, err
	}
	if _, err := tw.Write([]byte(dockerfileContent)); err != nil {
		return nil, err
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	return bytes.NewReader(buf.Bytes()), nil
}

func BuildWorkerImage(ctx context.Context, dockerfile string) error {
	cli, err := client.New(client.FromEnv, client.WithAPIVersionFromEnv())
	if err != nil {
		log.Println("Failed to initialize docker client")
		return err
	}
	defer cli.Close()

	buildContext, err := streamBuildContext(dockerfile)
	if err != nil {
		log.Printf("Failed generate build context tar stream for dockerfile: %s\n", dockerfile)
		return err
	}

	buildOpts := client.ImageBuildOptions{
		Tags:       []string{"ffmpeg-worker:latest"},
		Dockerfile: dockerfile,
	}
	buildResult, err := cli.ImageBuild(ctx, buildContext, buildOpts)
	if err != nil {
		log.Printf("Failed to build image. dockerfile: %s buildOpts: %v\n", dockerfile, buildOpts)
		return err
	}
	defer buildResult.Body.Close()

	_, err = io.Copy(os.Stdout, buildResult.Body)
	return nil
}
