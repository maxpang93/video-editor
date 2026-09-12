package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	ffmpegworker "video-editor-backend/ffmpeg_worker"

	"github.com/moby/moby/client"
)

var MediaFolder string = os.Getenv("MEDIA_FOLDER")

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/files", GetFiles)
	mux.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir(MediaFolder))))
	mux.HandleFunc("POST /api/process", ProcessVideoFile)
	http.ListenAndServe(":8090", mux)
}

type FileEntry struct {
	Name     string `json:"name"`
	FilePath string `json:"filepath"`
	Size     int64  `json:"size"`
	IsFolder bool   `json:"isFolder"`
}

func GetFiles(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	filePath := queryParams.Get("path")

	fullPath := filepath.Join(MediaFolder, filePath)
	files, err := os.ReadDir(fullPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to read directory: %q", filePath), http.StatusInternalServerError)
		return
	}

	var fileList []FileEntry
	for _, file := range files {
		fileInfo, err := file.Info()
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to read file: %q", file.Name()), http.StatusInternalServerError)
			return
		}
		fileList = append(fileList, FileEntry{
			Name:     fileInfo.Name(),
			Size:     fileInfo.Size(),
			FilePath: filepath.Join(filePath, fileInfo.Name()),
			IsFolder: fileInfo.IsDir(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(fileList)
}

type ProcessVideoFilePayload struct {
	VideoPath string `json:"video_path"`
	Segments  []struct {
		Start int `json:"start"`
		End   int `json:"end"`
	}
}

func ProcessVideoFile(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	payload := ProcessVideoFilePayload{}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	if _, err := os.Stat(filepath.Join(MediaFolder, payload.VideoPath)); err != nil {
		http.Error(w, "File not found!", http.StatusNotFound)
		return
	}

	ctx := context.Background()
	if err := ffmpegworker.BuildWorkerImage(ctx, "Dockerfile.worker"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = ffmpegworker.WithWorkerContainer(
		ctx,
		func(cli *client.Client, containerID string) error {
			for i, segment := range payload.Segments {
				cmd := ffmpegworker.GetTrimVideoCmd(payload.VideoPath, segment.Start, segment.End, i+1)
				if err := ffmpegworker.ExecContainerCmd(ctx, cli, containerID, cmd); err != nil {
					return err
				}
			}
			return nil
		},
	)
	if err != nil {
		log.Printf("video processing failed: %v", err)
		http.Error(w, "Video processing failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
