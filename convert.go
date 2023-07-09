package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var useFFMpegFromLocalDir bool

func convertToMp3(inputFilePath string, title string) (string, error) {

	outFile := DLDir
	outFile += title + ".mp3"
	// Specify the output MP3 file path
	outputFilePath := outFile

	// Create a new FFmpeg instance

	var (
		args = []string{"-i", inputFilePath, "-acodec:a", "libmp3lame", "-b:a", "256k", outFile}
	)
	var ffmpeg string
	if !useFFMpegFromLocalDir {
		ffmpeg = os.TempDir() + "/yt-dl-ui"
	} else {
		var err error
		ffmpeg, err = filepath.Abs("./")
		if err != nil {
			fmt.Println(err)
		}
	}
	ffmpeg += "/ffmpeg"
	var err error
	cmd := exec.Command(ffmpeg, args...)
	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
	err = os.Chmod(outFile, 0666)
	if err != nil {
		fmt.Println(err)
	}
	return outputFilePath, nil
}
