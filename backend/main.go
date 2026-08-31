package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

var MediaFolder string = os.Getenv("MEDIA_FOLDER")

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/files", GetFiles)
	mux.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir(MediaFolder))))
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
