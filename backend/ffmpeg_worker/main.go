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

	"github.com/google/uuid"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

func streamBuildContext(dockerfile string) (io.Reader, error) {
	dockerfileContent, err := os.ReadFile(dockerfile)
	if err != nil {
		return nil, fmt.Errorf("read Dockerfile %q: %w", dockerfile, err)
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
		return fmt.Errorf("initialize Docker client: %w", err)
	}
	defer cli.Close()

	buildContext, err := streamBuildContext(dockerfile)
	if err != nil {
		return fmt.Errorf("generate build context for Dockerfile %q: %w", dockerfile, err)
	}

	buildOpts := client.ImageBuildOptions{
		Tags:       []string{"ffmpeg-worker:latest"},
		Dockerfile: dockerfile,
	}
	buildResult, err := cli.ImageBuild(ctx, buildContext, buildOpts)
	if err != nil {
		return fmt.Errorf("build image from Dockerfile %q: %w", dockerfile, err)
	}
	defer buildResult.Body.Close()

	_, err = io.Copy(os.Stdout, buildResult.Body)
	return err
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

func WithWorkerContainer(
	ctx context.Context,
	work func(cli *client.Client, containerID string) error,
) error {
	cli, err := client.New(client.FromEnv, client.WithAPIVersionFromEnv())
	if err != nil {
		return fmt.Errorf("initialize Docker client: %w", err)
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
		return fmt.Errorf("create worker container: %w", err)
	}
	containerID := createResult.ID

	log.Printf("Starting container %s...\n", containerID)
	if _, err := cli.ContainerStart(ctx, containerID, client.ContainerStartOptions{}); err != nil {
		_, _ = cli.ContainerRemove(ctx, containerID, client.ContainerRemoveOptions{
			Force: true,
		})
		return fmt.Errorf("start worker container: %w", err)
	}
	defer func() {
		log.Printf("Cleaning up container %s...\n", containerID)
		if _, err := cli.ContainerRemove(ctx, containerID, client.ContainerRemoveOptions{
			Force: true,
		}); err != nil {
			log.Printf("Failed to remove container: %v\n", err)
		}
	}()

	return work(cli, containerID)
}

func GetOutputVideoPath(videoPath string, part int) string {
	filename, ext := splitFileExt(videoPath)
	return fmt.Sprintf("%s-%d%s", filename, part, ext)
}

func GetTrimVideoCmd(videoPath string, start int, end int, part int) []string {
	log.Printf(
		"Trim video cmd for part %d of video %s. Start %s. End %s",
		part, videoPath, secondsToTimestamp(start), secondsToTimestamp(end),
	)
	return []string{
		"ffmpeg", "-y",
		"-ss", secondsToTimestamp(start),
		"-t", secondsToTimestamp(end - start),
		"-i", filepath.Join("/video", videoPath),
		"-c", "copy",
		filepath.Join("/video", GetOutputVideoPath(videoPath, part)),
	}
}

func GetMergeVideoCmd(videoPath string, mergeMetadataPath string) []string {
	return []string{
		"ffmpeg", "-y",
		"-f", "concat",
		"-safe", "0",
		"-i", filepath.Join("/video", mergeMetadataPath),
		"-c", "copy",
		filepath.Join("/video", GetOutputVideoPath(videoPath, 0)),
	}
}

func MergeVideos(videoPath string, segments int) error {
	// Create unique merge metadata file on host system
	mergeMetadata := fmt.Sprintf("merge-%s.txt", uuid.New().String())
	mergeMetadataPath := filepath.Join(filepath.Dir(videoPath), mergeMetadata)
	mergeMetadataHostPath := filepath.Join(os.Getenv("MEDIA_FOLDER"), mergeMetadataPath)

	file, err := os.Create(mergeMetadataHostPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := os.Remove(mergeMetadataHostPath); err != nil {
			log.Printf("Error removing file: %v", err)
		}
	}()

	var parts []string
	for i := range segments {
		partPath := GetOutputVideoPath(videoPath, i+1)
		parts = append(parts, partPath)
		if _, err := fmt.Fprintf(file, "file '%s'\n", filepath.Base(partPath)); err != nil {
			return fmt.Errorf("write merge metadata: %w", err)
		}
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("error closing file: %w", err)
	}

	ctx := context.Background()
	if err = WithWorkerContainer(
		ctx,
		func(cli *client.Client, containerID string) error {
			cmd := GetMergeVideoCmd(videoPath, mergeMetadataPath)
			return ExecContainerCmd(ctx, cli, containerID, cmd)
		},
	); err != nil {
		return err
	}

	for _, partPath := range parts {
		err := os.Remove(filepath.Join(os.Getenv("MEDIA_FOLDER"), partPath))
		if err != nil {
			return fmt.Errorf("error removing file %s: %w", partPath, err)
		}
	}
	return nil
}

func ExecContainerCmd(ctx context.Context, cli *client.Client, containerID string, cmd []string) error {
	execCreateResult, err := cli.ExecCreate(ctx, containerID, client.ExecCreateOptions{
		AttachStdout: true,
		AttachStderr: true,
		TTY:          true,
		Cmd:          cmd,
	})
	if err != nil {
		return fmt.Errorf("create exec in container %q: %w", containerID, err)
	}

	execID := execCreateResult.ID
	execAttachResult, err := cli.ExecAttach(ctx, execID, client.ExecAttachOptions{})
	if err != nil {
		return fmt.Errorf("attach to exec %q: %w", execID, err)
	}
	defer execAttachResult.Close()

	_, err = cli.ExecStart(ctx, execID, client.ExecStartOptions{})
	if err != nil {
		return fmt.Errorf("start exec %q: %w", execID, err)
	}

	_, err = io.Copy(os.Stdout, execAttachResult.Reader)
	return err
}
