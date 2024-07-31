package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/PuerkitoBio/goquery"
)

const (
	SONGSTERR_METADATA_API = "https://www.songsterr.com/api/meta"
)

type SongsterrStatedata struct {
	Route struct {
		SongID int `json:"songId"`
	} `json:"route"`
}

type SongsterrMetadata struct {
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Source string `json:"source"`
}

type Result struct {
	SongsterrMetadata
	SongID   int    `json:"songId"`
	Filename string `json:"filename"`
}

func main() {
	targetUrl := flag.String("u", "", "Songsterr URL")
	flag.Parse()
	if targetUrl == nil || *targetUrl == "" {
		flag.Usage()
		os.Exit(1)
	}
	// Fetch SongID
	res, err := http.Get(*targetUrl)
	must(err)
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	must(err)
	res.Body.Close()
	jsonResp := doc.Find("#state").Text()
	must(err)
	jsonResp, err = url.QueryUnescape(jsonResp)
	must(err)
	state := &SongsterrStatedata{}
	must(json.Unmarshal([]byte(jsonResp), state))
	if state.Route.SongID == 0 {
		log.Fatal("No song ID found")
	}

	// Fetch Tab file
	resApi, err := http.Get(fmt.Sprintf("%s/%d", SONGSTERR_METADATA_API, state.Route.SongID))
	must(err)
	if resApi.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
	bodyBytes, err := io.ReadAll(resApi.Body)
	must(err)
	resApi.Body.Close()
	metadata := &SongsterrMetadata{}
	must(json.Unmarshal(bodyBytes, metadata))

	// Write Tab file
	result := Result{
		SongsterrMetadata: *metadata,
		SongID:            state.Route.SongID,
		Filename:          fmt.Sprintf("%s - %s%s", metadata.Title, metadata.Artist, filepath.Ext(metadata.Source)),
	}
	out, err := os.Create(result.Filename)
	must(err)
	defer out.Close()
	respfile, err := http.Get(metadata.Source)
	must(err)
	_, err = io.Copy(out, respfile.Body)
	must(err)
	respfile.Body.Close()
	jsonResult, err := json.Marshal(result)
	must(err)
	fmt.Println(string(jsonResult))
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
