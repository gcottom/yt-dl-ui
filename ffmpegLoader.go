package main

import (
	"io/ioutil"
	"os"
)

func loadFfmpegWin() {
	// Open the resource for reading
	reader := ffmpegwin.StaticContent

	os.Mkdir(os.TempDir()+"\\yt-dl-ui", 0666)
	// Write the resource content to the file
	err := ioutil.WriteFile(os.TempDir()+"\\yt-dl-ui\\ffmpeg.exe", reader, 0666)
	if err != nil {
		panic(err)
	}
}
func loadFfmpegMac() {
	// Open the resource for reading
	reader := ffmpegwin.StaticContent

	os.Mkdir(os.TempDir()+"/yt-dl-ui", 0666)
	// Write the resource content to the file
	err := ioutil.WriteFile(os.TempDir()+"/yt-dl-ui/ffmpeg", reader, 0666)
	if err != nil {
		panic(err)
	}
}
