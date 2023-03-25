/*
 * mastercoderk@gmail.com
 */package util

import (
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"chloe/def"

	log "github.com/jeanphorn/log4go"
)

func DownloadTempFile(link string) (string, def.CleanFunc) {
	ext := filepath.Ext(link)
	f, err := os.CreateTemp("", "*"+ext)
	if err != nil {
		log.Error("creating temp file failed, %v", err)
		return "", nil
	}
	defer f.Close()
	fpath := f.Name()

	resp, err := http.Get(link)
	if err != nil {
		log.Error("download file %s failed, %v", link, err)
		return "", nil
	}
	defer resp.Body.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		log.Error("copy file failed, %v", err)
		return "", nil
	}
	return fpath, func() {
		_ = os.Remove(fpath)
	}
}

func ConvertToMp3(audioFile string) (string, def.CleanFunc) {
	ext := filepath.Ext(audioFile)
	mp3Filepath := audioFile[:len(audioFile)-1-len(ext)] + ".mp3"

	cmd := exec.Command(
		"ffmpeg",
		"-i",
		audioFile,
		"-vn",
		"-ar",
		"44100",
		"-ac",
		"2",
		"-ab",
		"192k",
		"-f",
		"mp3",
		mp3Filepath,
	)
	if err := cmd.Run(); err != nil {
		log.Error("failed to convert voice file %s to mp3, %v", audioFile, err)
		return "", nil
	}

	return mp3Filepath, func() {
		_ = os.Remove(mp3Filepath)
	}
}
