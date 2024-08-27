package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os/exec"
	"text/template"
	"time"
)

//go:embed index.html
var tpl embed.FS

const command = `
	tell application "Spotify"
    set ctrack to "{"
    set ctrack to ctrack & "\"artist\": \"" & current track's artist & "\""
    set ctrack to ctrack & ",\"album\": \"" & current track's album & "\""
    set ctrack to ctrack & ",\"discNumber\": " & current track's disc number
    set ctrack to ctrack & ",\"duration\": " & current track's duration
    set ctrack to ctrack & ",\"playedCount\": " & current track's played count
    set ctrack to ctrack & ",\"trackNumber\": " & current track's track number
    set ctrack to ctrack & ",\"popularity\": " & current track's popularity
    set ctrack to ctrack & ",\"id\": \"" & current track's id & "\""
    set ctrack to ctrack & ",\"position\": " & (player position as integer)
    set ctrack to ctrack & ",\"name\": \"" & current track's name & "\""
    set ctrack to ctrack & ",\"albumArtist\": \"" & current track's album artist & "\""
    set ctrack to ctrack & ",\"artworkUrl\": \"" & current track's artwork url & "\""
    set ctrack to ctrack & ",\"spotifyUrl\": \"" & current track's spotify url & "\""
    set ctrack to ctrack & "}"

		return ctrack
	end tell`

type SpotifyCurrentTrack struct {
	Artist      string `json:"artist"`
	Album       string `json:"album"`
	DiscNumber  int    `json:"discNumber"`
	Duration    int    `json:"duration"`
	PlayedCount int    `json:"playedCount"`
	TrackNumber int    `json:"trackNumber"`
	Popularity  int    `json:"popularity"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Position    int    `json:"position"`
	AlbumArtist string `json:"albumArtist"`
	ArtworkURL  string `json:"artworkUrl"`
	SpotifyURL  string `json:"spotifyUrl"`
}

type Response struct {
	Track   SpotifyCurrentTrack
	Refresh int64
	Port    string
}

func getCurrentTrack() SpotifyCurrentTrack {
	currentTrack := SpotifyCurrentTrack{}

	cmd := exec.Command("osascript", "-e", command)

	output, err := cmd.CombinedOutput()

	if err != nil {
		return currentTrack
	}

	if err := json.Unmarshal(output, &currentTrack); err != nil {
		return currentTrack
	}

	return currentTrack
}

func main() {
	port := flag.String("port", "5783", "http port")
	flag.Parse()

	tmpl, err := template.ParseFS(tpl, "index.html")

	if err != nil {
		panic(err)
	}

	http.HandleFunc(
		"/", func(w http.ResponseWriter, req *http.Request) {
			response := Response{
				Track:   getCurrentTrack(),
				Refresh: time.Second.Milliseconds(),
				Port:    *port,
			}

			refreshQueryParam := req.URL.Query().Get("refresh")

			if refreshQueryParam != "" {
				if duration, err := time.ParseDuration(refreshQueryParam); err == nil {
					response.Refresh = duration.Milliseconds()
				}
			}

			if err := tmpl.Execute(w, response); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		},
	)

	http.HandleFunc(
		"/track", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")

			resp, _ := json.Marshal(getCurrentTrack())

			fmt.Fprint(w, string(resp))
		},
	)

	fmt.Printf("Add Overlay to OBS Studio with URL: http://localhost:%s/", *port)

	http.ListenAndServe(fmt.Sprintf(":%s", *port), nil)
}
