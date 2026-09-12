package ffmpegworker

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/moby/moby/api/types/container"
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

func secondsToTimestamp(t int) string {
	seconds := t % 60
	totalMinute := t / 60
	minutes := totalMinute % 60
	hours := totalMinute / 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func splitFileExt(filename string) (name string, ext string) {
	ext = filepath.Ext(filename)
	name = filename[:len(filename)-len(ext)]
	return
}

func RunWorker(ctx context.Context, videoPath string, start int, end int, part int) error {
	log.Printf(
		"Processing part %d of video %s. Start %s. End %s",
		part, videoPath, secondsToTimestamp(start), secondsToTimestamp(end),
	)
	cli, err := client.New(client.FromEnv, client.WithAPIVersionFromEnv())
	if err != nil {
		log.Println("Failed to initialize docker client")
		return err
	}
	defer cli.Close()

	config := &container.Config{
		Image: "ffmpeg-worker:latest",
		Volumes: map[string]struct{}{
			"/video": {},
		},
	}

	sourcePath := os.Getenv("MEDIA_FOLDER")
	hostConfig := &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/video", sourcePath),
		},
	}

	createResult, err := cli.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config:     config,
		HostConfig: hostConfig,
	})
	if err != nil {
		log.Printf("Failed to create container: %v\n", err)
		return err
	}
	containerID := createResult.ID

	log.Printf("Starting container %s...\n", containerID)
	_, err = cli.ContainerStart(ctx, containerID, client.ContainerStartOptions{})
	if err != nil {
		log.Printf("Failed to start container: %v\n", err)
		return err
	}
	defer func() {
		log.Printf("Cleaning up container %s...\n", containerID)
		_, err := cli.ContainerRemove(ctx, containerID, client.ContainerRemoveOptions{
			Force: true,
		})
		if err != nil {
			log.Printf("Failed to remove container: %v\n", err)
		}
	}()

	execCreateResult, err := cli.ExecCreate(ctx, containerID, client.ExecCreateOptions{
		AttachStdout: true,
		AttachStderr: true,
		TTY:          true,
		Cmd: []string{
			"ffmpeg", "-y",
			"-ss", secondsToTimestamp(start),
			"-t", secondsToTimestamp(end - start),
			"-i", filepath.Join("/video", videoPath),
			"-c", "copy",
			filepath.Join("/video", fmt.Sprintf("output-%d.mp4", part)),
		},
	})
	if err != nil {
		log.Printf("Failed to exec create: %v\n", err)
		return err
	}

	execID := execCreateResult.ID
	execAttachResult, err := cli.ExecAttach(ctx, execID, client.ExecAttachOptions{})
	if err != nil {
		log.Printf("Failed to exec attach: %v\n", err)
		return err
	}
	defer execAttachResult.Close()

	_, err = cli.ExecStart(ctx, execID, client.ExecStartOptions{})
	if err != nil {
		log.Printf("Failed to exec start: %v\n", err)
		return err
	}

	_, err = io.Copy(os.Stdout, execAttachResult.Reader)
	if err != nil {
		return err
	}

	return nil
}
