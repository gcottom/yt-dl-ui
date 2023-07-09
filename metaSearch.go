package main

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/gcottom/musicbrainz"
	spotifyauth "github.com/zmb3/spotify/v2/auth"

	"golang.org/x/oauth2/clientcredentials"

	lastfm "github.com/shkh/lastfm-go"

	spotify "github.com/zmb3/spotify/v2"
)

type Meta struct {
	albumImage string
	album      string
	albumID    spotify.ID
	artist     string
	song       string
	trackID    spotify.ID
	genre      string
	year       string
	bpm        string
	infoSource string
}

var genres = [...]string{
	"Blues", "Classic Rock", "Country", "Dance", "Disco", "Funk", "Grunge",
	"Hip-Hop", "Jazz", "Metal", "New Age", "Oldies", "Other", "Pop", "R&B",
	"Rap", "Reggae", "Rock", "Techno", "Industrial", "Alternative", "Ska",
	"Death Metal", "Pranks", "Soundtrack", "Euro-Techno", "Ambient",
	"Trip-Hop", "Vocal", "Jazz+Funk", "Fusion", "Trance", "Classical",
	"Instrumental", "Acid", "House", "Game", "Sound Clip", "Gospel",
	"Noise", "AlternRock", "Bass", "Soul", "Punk", "Space", "Meditative",
	"Instrumental Pop", "Instrumental Rock", "Ethnic", "Gothic",
	"Darkwave", "Techno-Industrial", "Electronic", "Pop-Folk",
	"Eurodance", "Dream", "Southern Rock", "Comedy", "Cult", "Gangsta",
	"Top 40", "Christian Rap", "Pop/Funk", "Jungle", "Native American",
	"Cabaret", "New Wave", "Psychedelic", "Rave", "Showtunes", "Trailer",
	"Lo-Fi", "Tribal", "Acid Punk", "Acid Jazz", "Polka", "Retro",
	"Musical", "Rock & Roll", "Hard Rock", "Folk", "Folk-Rock",
	"National Folk", "Swing", "Fast Fusion", "Bebob", "Latin", "Revival",
	"Celtic", "Bluegrass", "Avantgarde", "Gothic Rock", "Progressive Rock",
	"Psychedelic Rock", "Symphonic Rock", "Slow Rock", "Big Band",
	"Chorus", "Easy Listening", "Acoustic", "Humour", "Speech", "Chanson",
	"Opera", "Chamber Music", "Sonata", "Symphony", "Booty Bass", "Primus",
	"Porn Groove", "Satire", "Slow Jam", "Club", "Tango", "Samba",
	"Folklore", "Ballad", "Power Ballad", "Rhythmic Soul", "Freestyle",
	"Duet", "Punk Rock", "Drum Solo", "A capella", "Euro-House", "Dance Hall",
	"Goa", "Drum & Bass", "Club-House", "Hardcore", "Terror", "Indie",
	"Britpop", "Negerpunk", "Polsk Punk", "Beat", "Christian Gangsta Rap",
	"Heavy Metal", "Black Metal", "Crossover", "Contemporary Christian",
	"Christian Rock ", "Merengue", "Salsa", "Thrash Metal", "Anime", "JPop",
	"Synthpop",
}

var resultMeta []Meta
var SongTitle string
var singleArtist string

func getMetaFromSong(songName string) {
	ctx := context.Background()
	config := &clientcredentials.Config{
		ClientID:     spotifyClientID,
		ClientSecret: spotifySecret,
		TokenURL:     spotifyauth.TokenURL,
	}
	token, err := config.Token(ctx)
	if err != nil {
		log.Fatalf("couldn't get token: %v", err)
	}

	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient)
	// search for playlists and albums containing "holiday"
	results, err := client.Search(ctx, songName, spotify.SearchTypeTrack)
	if err == nil {
		processMeta(results)
	}

}
func getMetaFromSongAndArtist(songName string, artist string) error {
	resultMeta = []Meta{}
	ctx := context.Background()
	config := &clientcredentials.Config{
		ClientID:     spotifyClientID,
		ClientSecret: spotifySecret,
		TokenURL:     spotifyauth.TokenURL,
	}
	token, err := config.Token(ctx)
	if err != nil {
		return err
	}
	searchTerm := songName + "artist:" + artist
	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient)
	results, err := client.Search(ctx, searchTerm, spotify.SearchTypeTrack)
	if err == nil {
		processMeta(results)
	}
	return nil
}
func getMetaFromSongAndArtistLastFm(songName string, artist string) error {
	api := lastfm.New(lastFmApiKey, lastFmSecret) //lastfm creds
	lfmtoken, _ := api.GetToken()                 //discarding error
	//Send your user to "authUrl"
	//Once the user grant permission, then authorize the token.
	api.LoginWithToken(lfmtoken) //discarding error
	response, err := api.Track.GetInfo(lastfm.P{"track": songName, "artist": artist})
	if err != nil {
		return err
	} else {
		artist := response.Artist.Name
		song := response.Name
		album := response.Album.Title
		var albumImage = ""
		if len(response.Album.Images) > 0 {
			albumImage = response.Album.Images[0].Url
		}
		var trackGenre = ""
		for _, tag := range response.TopTags {
			log.Println("Tags: " + tag.Name)
			for _, genre := range genres {
				g := genre
				if strings.Compare(strings.ToLower(tag.Name), strings.ToLower(genre)) == 0 {
					log.Println("SET TAG:" + tag.Name + ", GENRE:" + genre)
					trackGenre = g
					break
				}
			}
			if trackGenre != "" {
				break
			}
			for _, genre := range genres {
				g := genre
				if strings.Contains(strings.ToLower(tag.Name), strings.ToLower(genre)) || strings.Contains(strings.ToLower(genre), strings.ToLower(tag.Name)) {
					trackGenre = g
					break
				}
			}
			if trackGenre != "" {
				break
			}
		}
		outMeta := Meta{albumImage, album, "", artist, song, "", trackGenre, "", "", "lastfm"}
		resultMeta = []Meta{}
		resultMeta = append(resultMeta, outMeta)
		return nil
	}

}
func getMetaFromSongAndArtistMusicBrainz(song string, artist string) error {
	response, err := musicbrainz.SearchRecordingsByTitleAndArtist(song, artist)
	if err != nil {
		return err
	}
	resultMeta = []Meta{}
	for _, recording := range response {
		album := ""
		if len(recording.Releases) > 0 {
			album = recording.Releases[0].Title
		}
		artist := ""
		if len(recording.ArtistCredit) > 0 {
			artist = recording.ArtistCredit[0].Name
		}
		outMeta := Meta{"", album, "", artist, recording.Title, "", "", "", "", "musicbrainz"}
		resultMeta = append(resultMeta, outMeta)
	}
	return nil
}
func (m *Meta) getAlbumMeta() {
	ctx := context.Background()
	config := &clientcredentials.Config{
		ClientID:     spotifyClientID,
		ClientSecret: spotifySecret,
		TokenURL:     spotifyauth.TokenURL,
	}
	token, err := config.Token(ctx)
	if err != nil {
		log.Fatalf("couldn't get token: %v", err)
	}
	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient)
	// search for playlists and albums containing "holiday"
	r := spotify.Market("US")
	if m.albumID != "" && m.trackID != "" {
		results, err := client.GetAlbum(ctx, m.albumID, r)
		if err == nil {
			m.year = results.ReleaseDate[:4]
			if err != nil {
				log.Fatal(err)
			}
		}
		aresults, err := client.GetAudioAnalysis(ctx, m.trackID)
		if err == nil {
			m.bpm = strconv.Itoa(int(aresults.Track.Tempo))
		}
	}
	if len(strings.Split(m.artist, ",")) > 1 {
		singleArtist = strings.Split(m.artist, ",")[0]
	} else {
		singleArtist = m.artist
	}
	api := lastfm.New(lastFmApiKey, lastFmSecret) //lastfm creds
	lfmtoken, _ := api.GetToken()                 //discarding error
	//Send your user to "authUrl"
	//Once the user grant permission, then authorize the token.
	api.LoginWithToken(lfmtoken) //discarding error

	response, err := api.Track.GetInfo(lastfm.P{"track": m.song, "artist": singleArtist})
	if err != nil {
		log.Println("Error:")
		log.Println(err)
		return
	}
	var trackGenre string
	for _, tag := range response.TopTags {
		log.Println("Tags: " + tag.Name)
		for _, genre := range genres {
			g := genre
			if strings.Compare(strings.ToLower(tag.Name), strings.ToLower(genre)) == 0 {
				trackGenre = g
				break
			}
		}
		if trackGenre != "" {
			break
		}
		for _, genre := range genres {
			g := genre
			if strings.Contains(strings.ToLower(tag.Name), strings.ToLower(genre)) || strings.Contains(strings.ToLower(genre), strings.ToLower(tag.Name)) {
				trackGenre = g
				break
			}
		}
		if trackGenre != "" {
			break
		}
	}
	m.genre = trackGenre
}
func processMeta(results *spotify.SearchResult) {
	resultMeta = []Meta{}
	for _, track := range results.Tracks.Tracks {
		var albumImage = ""
		if len(track.Album.Images) > 0 {
			albumImage = track.Album.Images[0].URL
		}
		album := track.Album.Name
		albumID := track.Album.ID
		artist := ""
		for _, art := range track.Artists {
			artist += art.Name + ", "
		}
		artist = artist[:(strings.LastIndex(artist, ", "))] + strings.Replace(artist[(strings.LastIndex(artist, ", ")):], ", ", "", 1)
		song := track.Name
		trackID := track.ID

		outMeta := Meta{albumImage, album, albumID, artist, song, trackID, "", "", "", "spotify"}
		resultMeta = append(resultMeta, outMeta)
	}
}
