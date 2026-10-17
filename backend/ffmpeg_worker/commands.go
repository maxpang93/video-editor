package ffmpegworker

import (
	"fmt"
	"log"
	"path/filepath"
)

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
