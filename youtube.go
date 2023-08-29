package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/kkdai/youtube/v2"
)

var tempFile string

func download(id string) (string, string, error) {
	id = strings.Replace(id, "&feature=share", "", 1)
	id = strings.Replace(id, "https://music.youtube.com/watch?v=", "", 1)
	id = strings.Replace(id, "https://www.music.youtube.com/watch?v=", "", 1)
	id = strings.Replace(id, "https://www.youtube.com/watch?v", "", 1)
	id = strings.Replace(id, "https://youtube.com/watch?v", "", 1)
	id = strings.Split(id, "&")[0]
	videoID := id // Replace with the YouTube video ID you want to download

	// Create a new YouTube client
	client := youtube.Client{}

	// Get the video info
	videoInfo, err := client.GetVideo(videoID)
	if err != nil {
		fmt.Printf("Failed to get video info: %v\n", err)
		err = errors.New(fmt.Sprintf("Failed to get video info: %v\n", err))
		return "", "", err
	}
	// Find the best audio format available
	bestFormat := getBestAudioFormat(videoInfo.Formats.Type("audio"))
	if bestFormat == nil {
		err = errors.New(fmt.Sprintf("No audio formats found for the video"))
		return "", "", err
	}
	stream, _, err := client.GetStream(videoInfo, bestFormat)
	if err != nil {
		err = errors.New(fmt.Sprintf("No Stream found"))
		return "", "", err
	}
	title := SanitizeFilename(videoInfo.Title)
	tempFile = os.TempDir() + "/yt-dl-ui/" + title + ".tmp"
	os.Remove(tempFile)
	if err != nil {
		fmt.Printf("Unable to remove temp file: %v\n", err)
	}
	// Download the video in the best audio format
	file, err := os.Create(tempFile)
	if err != nil {
		err = errors.New(fmt.Sprintf("Unable to create temp file: %v\n", err))
		return "", "", err
	}

	defer file.Close()
	_, err = io.Copy(file, stream)
	if err != nil {
		err = errors.New(fmt.Sprintf("Unable to copy stream data to file object: %v\n", err))
		return "", "", err
	}
	return tempFile, title, nil
}

// getBestAudioFormat finds the best audio format from a list of formats
func getBestAudioFormat(formats youtube.FormatList) *youtube.Format {
	var bestFormat *youtube.Format
	maxBitrate := 0

	for _, format := range formats {
		if format.Bitrate > maxBitrate {
			bestFormat = &format
			maxBitrate = format.Bitrate
		}
	}
	fmt.Println(bestFormat.QualityLabel)
	return bestFormat
}
func SanitizeFilename(fileName string) string {
	// Characters not allowed on mac
	//	:/
	// Characters not allowed on linux
	//	/
	// Characters not allowed on windows
	//	<>:"/\|?*

	// Ref https://docs.microsoft.com/en-us/windows/win32/fileio/naming-a-file#naming-conventions

	fileName = regexp.MustCompile(`[:/<>\:"\\|?*]`).ReplaceAllString(fileName, "")
	fileName = regexp.MustCompile(`\s+`).ReplaceAllString(fileName, " ")

	return fileName
}
