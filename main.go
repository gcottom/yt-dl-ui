package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"

	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	tag "github.com/gcottom/mp3-mp4-tag"
)

var (
	a                    fyne.App
	w                    fyne.Window
	prefs                fyne.Preferences
	msb                  *widget.Entry
	searchMetaWithArtist func(s string)
	searchMeta           func()
	settings             func()
	saveSettings         func()
	trackSavedNotif      func()
	title                string
	OutFile              string
	DLDir                string
	pendingPrefs         *PendingPreferences
)

const (
	lineLimit = 50
)

type PendingPreferences struct {
	dldir string
}

func formatSearchResults(artist string, song string, album string) string {
	if len(artist)/lineLimit > 0 {
		split := strings.Split(artist, " ")
		temp := ""
		artist = ""
		for i, word := range split {
			if (len(temp) + len(word)) < lineLimit {
				temp = temp + " " + word
				if i+1 == len(split) {
					artist = artist + "\n" + temp
				}
			} else {
				if artist == "" {
					artist = temp
					temp = word
				} else {
					artist = artist + "\n" + temp
					temp = word
					if i+1 == len(split) {
						artist = artist + "\n" + temp
					}
				}
			}
		}
	}
	if len(song)/lineLimit > 0 {
		split := strings.Split(song, " ")
		temp := ""
		song = ""
		for i, word := range split {
			if (len(temp) + len(word)) < lineLimit {
				temp = temp + " " + word
				if i+1 == len(split) {
					song = song + "\n" + temp
				}
			} else {
				if song == "" {
					song = temp
					temp = word
				} else {
					song = song + "\n" + temp
					temp = word
					if i+1 == len(split) {
						song = song + "\n" + temp
					}
				}
			}
		}
	}
	if len(album)/lineLimit > 0 {
		split := strings.Split(artist, " ")
		temp := ""
		album = ""
		for i, word := range split {
			if (len(temp) + len(word)) < lineLimit {
				temp = temp + " " + word
				if i+1 == len(split) {
					album = album + "\n" + temp
				}
			} else {
				if album == "" {
					album = temp
					temp = word
				} else {
					album = album + "\n" + temp
					temp = word
					if i+1 == len(split) {
						album = album + "\n" + temp
					}
				}
			}
		}
	}
	return "Artist: " + artist + "\n\nTrack: " + song + "\n\nAlbum: " + album
}

func init() {
	searchMetaWithArtist = func(s string) {
		err := getMetaFromSongAndArtist(title, s)
		if err != nil {
			fmt.Println(err)
		}
		filteredElements := []*fyne.Container{}
		for _, result := range resultMeta {
			if strings.Contains(strings.ToLower(result.artist), strings.ToLower(s)) {
				meta := result
				img, err := fyne.LoadResourceFromURLString(result.albumImage)
				if err != nil {
					fmt.Println("error getting image")
				}
				button := NewCustomButton(formatSearchResults(meta.artist, meta.song, meta.album), img, func() {
					saveMeta(meta, OutFile)
					showMainScreen()
				})
				hbox := container.New(layout.NewHBoxLayout(), button)
				filteredElements = append(filteredElements, hbox)
			}
		}
		if len(filteredElements) == 0 {
			resultMeta = []Meta{}
			err := getMetaFromSongAndArtistLastFm(title, s)
			if err != nil {
				fmt.Println(err)
			}
			for _, result := range resultMeta {
				if strings.Contains(strings.ToLower(result.artist), strings.ToLower(s)) {
					meta := result
					img, err := fyne.LoadResourceFromURLString(result.albumImage)
					if err != nil {
						fmt.Println("error getting image")
					}
					button := NewCustomButton(formatSearchResults(meta.artist, meta.song, meta.album), img, func() {
						saveMeta(meta, OutFile)
						showMainScreen()
					})
					hbox := container.New(layout.NewHBoxLayout(), button)
					filteredElements = append(filteredElements, hbox)
				}
			}
		}
		if len(filteredElements) == 0 {
			resultMeta = []Meta{}
			err := getMetaFromSongAndArtistMusicBrainz(title, s)
			if err != nil {
				fmt.Println(err)
			}
			for _, result := range resultMeta {
				if strings.Contains(strings.ToLower(result.artist), strings.ToLower(s)) {
					meta := result
					button := NewCustomButton(formatSearchResults(meta.artist, meta.song, meta.album), theme.ErrorIcon(), func() {
						saveMeta(meta, OutFile)
						showMainScreen()
					})
					hbox := container.New(layout.NewHBoxLayout(), button)
					filteredElements = append(filteredElements, hbox)
				}
			}
		}
		addArtistTitle := widget.NewLabel("Search Artist:")
		selectTagsLabel := widget.NewLabel("Select Tags For The Downloaded Track")
		searchBar := msb
		done := widget.NewButton("Done", showMainScreen)
		cbox := container.New(layout.NewVBoxLayout(), selectTagsLabel, container.NewBorder(nil, nil, addArtistTitle, done, searchBar))
		var vbox *fyne.Container
		if len(filteredElements) == 0 {
			noResults := widget.NewLabel("No Matching Results Found!")
			vbox = container.New(layout.NewVBoxLayout(), cbox, noResults)
		} else {
			vbox = container.New(layout.NewVBoxLayout(), cbox)

			for _, element := range filteredElements {
				vbox.Add(element)
			}
		}

		w.SetContent(container.NewVScroll(vbox))
	}
	searchMeta = func() {
		getMetaFromSong(title)
		var elements []*fyne.Container
		for _, result := range resultMeta {
			meta := result
			img, err := fyne.LoadResourceFromURLString(result.albumImage)
			if err != nil {
				fmt.Println("Error getting image")
			}
			button := NewCustomButton(formatSearchResults(meta.artist, meta.song, meta.album), img, func() {
				saveMeta(meta, OutFile)
				showMainScreen()
			})
			hbox := container.New(layout.NewHBoxLayout(), button)
			elements = append(elements, hbox)
		}
		addArtistTitle := widget.NewLabel("Search Artist:")
		selectTagsLabel := widget.NewLabel("Select Tags For The Downloaded Track")
		searchBar := msb
		done := widget.NewButton("Done", showMainScreen)
		vbox := container.New(layout.NewVBoxLayout(), selectTagsLabel, container.NewBorder(nil, nil, addArtistTitle, done, searchBar))
		for _, element := range elements {
			vbox.Add(element)
		}
		w.SetContent(container.NewVScroll(vbox))
	}
	settings = func() {
		pendingPrefs = &PendingPreferences{}
		pendingPrefs.dldir = DLDir
		outDirLabel := widget.NewLabel("Output Directory:")
		outDirText := widget.NewEntry()
		outDirText.Text = prefs.String("dldir")
		outDirText.MinSize()
		outDirFolderButton := widget.NewButtonWithIcon("", theme.FolderIcon(), func() {
			onChosen := func(f fyne.ListableURI, err error) {
				if err != nil {
					fmt.Println(err)
					return
				}
				if f == nil {
					return
				}
				uri := f.Path()
				if runtime.GOOS == "windows" {
					uri = strings.ReplaceAll(uri, "/", "\\")
					if !strings.HasSuffix(uri, "\\") {
						uri += "\\"
					}
				} else {
					if !strings.HasSuffix(uri, "/") {
						uri += "/"
					}
				}
				outDirText.Text = uri
				pendingPrefs.dldir = uri
				outDirText.Refresh()
			}
			dialog.ShowFolderOpen(onChosen, w)
		})
		deleteTempLabel := widget.NewLabel("Delete Temp Files:")
		deleteTempButton := widget.NewButton("Delete", func() {
			dialog.ShowConfirm("Delete Temp Files", "Are you sure you want to delete temporary files?", func(confirm bool) {
				if confirm {
					err := os.RemoveAll(os.TempDir() + "/yt-dl-ui/")
					os.Mkdir(os.TempDir()+"/yt-dl-ui/", 0666)
					if runtime.GOOS == "windows" {
						loadFfmpegWin()
					} else if runtime.GOOS == "darwin" {
						loadFfmpegMac()
					}
					if err != nil {
						fmt.Println(err)
						return
					} else {
						dialog.ShowInformation("Temp Files Deleted", "Temporary files have been successfully deleted!", w)
					}
				}
			}, w)
		})
		w.SetContent(container.NewBorder(container.NewBorder(nil, nil, widget.NewLabel("Settings"), widget.NewButtonWithIcon("", theme.CancelIcon(), func() { showMainScreen() }), nil), nil, container.NewVBox(outDirLabel, deleteTempLabel), container.NewVBox(outDirFolderButton, widget.NewLabel(" "), widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() { saveSettings() })), container.NewVBox(outDirText, deleteTempButton)))
	}
	saveSettings = func() {
		prefs.SetString("dldir", pendingPrefs.dldir)
		DLDir = pendingPrefs.dldir
		dialog.ShowInformation("Settings Saved", "Changes Have Been Saved Successfully!", w)
		showMainScreen()
	}
	trackSavedNotif = func() {
		dialog.ShowInformation("Complete", "Track Downloaded and MetaData Saved Successfully!", w)
	}
}
func main() {
	appIcon, err := fyne.LoadResourceFromPath("appIcon.png")
	if err != nil {
		fmt.Print(err)
	}
	a = app.NewWithID("github.com.gcottom.yt-dl-ui")
	a.SetIcon(appIcon)
	prefs = a.Preferences()
	dldir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	if strings.Contains(dldir, "/") {
		dldir += "/Downloads/YTDownloads/"
	} else {
		dldir += "\\Downloads\\YTDownloads\\"
	}
	DLDir = prefs.StringWithFallback("dldir", dldir)
	if _, err = os.Stat(DLDir); os.IsNotExist(err) {
		os.Mkdir(DLDir, 0666)
	}
	prefs.SetString("dldir", DLDir)
	if runtime.GOOS == "windows" {
		loadFfmpegWin()
	} else if runtime.GOOS == "darwin" {
		loadFfmpegMac()
	}
	os.Mkdir(os.TempDir()+"/yt-dl-ui/", 0666)
	w = a.NewWindow("yt-dl-ui - Youtube Downloader")
	w.SetIcon(appIcon)
	w.Resize(fyne.NewSize(600, 600))
	w.SetFixedSize(true)
	showMainScreen()
	w.ShowAndRun()

}

func showMainScreen() {
	msb = widget.NewEntry()
	msb.OnChanged = searchMetaWithArtist
	titleLabel := widget.NewLabel("Youtube URL:")
	urlBox := widget.NewEntry()
	downloadButton := widget.NewButton("DOWNLOAD", func() {
		if strings.TrimSpace(urlBox.Text) != "" {
			pb := widget.NewProgressBarInfinite()
			tbSpacer := layout.NewSpacer()
			tbSpacer.Resize(fyne.NewSize(0, 200))
			w.SetContent(container.NewCenter(pb))
			var err error
			var tempFile string
			tempFile, title, err = download(urlBox.Text)
			if err != nil {
				fmt.Print("Download Error")
				handleError(err)
				showMainScreen()
			} else {
				OutFile, err = convertToMp3(tempFile, title)
				os.Chmod(tempFile, 0666)
				os.Remove(tempFile)
				if err != nil {
					fmt.Print("Conversion Error")
					handleError(err)
					showMainScreen()
				} else {
					searchMeta()

				}
			}

		}
	})
	topbox := container.New(layout.NewHBoxLayout(), widget.NewLabel("Download A Track"), layout.NewSpacer(), widget.NewButtonWithIcon("Settings", theme.SettingsIcon(), func() {
		settings()
	}))
	hContent := container.New(layout.NewVBoxLayout(), container.New(layout.NewFormLayout(), titleLabel, urlBox), downloadButton)
	vBox := container.New(layout.NewVBoxLayout(), topbox, hContent)
	w.SetContent(vBox)
}
func saveMeta(meta Meta, filepath string) {
	outDir := DLDir
	coverFileName := os.TempDir() + "/yt-dl-ui/cover.jpg"
	url := meta.albumImage
	// don't worry about errors
	if url != "" {
		response, err := http.Get(url)
		if err != nil {
			response.Body.Close()
			fmt.Println(err)
			handleError(err)
		} else {
			file, err := os.Create(coverFileName)
			if err != nil {
				file.Close()
				fmt.Println(err)
				handleError(err)
			} else {
				defer response.Body.Close()
				defer file.Close()
				_, err = io.Copy(file, response.Body)
				if err != nil {
					fmt.Println(err)
					handleError(err)
				} else {
					meta.getAlbumMeta()
					//open a file for writing

					idTag, err := tag.OpenTag(filepath)
					if err != nil {
						fmt.Println(err)
						handleError(err)
					} else {
						idTag.SetTitle(meta.song)
						if url != "" {
							idTag.SetAlbumArtFromFilePath(coverFileName)
						}
						idTag.SetAlbum(meta.album)
						idTag.SetArtist(meta.artist)
						idTag.SetGenre(meta.genre)
						idTag.SetYear(meta.year)
						idTag.SetBPM(meta.bpm)
						idTag.Save()
						os.Remove(coverFileName)

						os.Rename(OutFile, outDir+SanitizeFilename(singleArtist)+" - "+SanitizeFilename(meta.song)+".mp3")
						trackSavedNotif()
					}

				}
			}

		}

	}

}
func handleError(err error) {
	s := fmt.Sprintf("FUCK\n%v", err)
	if len(s)/lineLimit > 0 {
		split := strings.Split(s, " ")
		temp := ""
		s = ""
		for i, word := range split {
			if (len(temp) + len(word)) < lineLimit {
				temp = temp + " " + word
				if i+1 == len(split) {
					s = s + "\n" + temp
				}
			} else {
				if s == "" {
					s = temp
					temp = word
				} else {
					s = s + "\n" + temp
					temp = word
					if i+1 == len(split) {
						s = s + "\n" + temp
					}
				}
			}
		}
	}
	dialog.ShowError(errors.New(s), w)
}
