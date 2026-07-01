package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
)

//go:embed ui/dist
var uiFiles embed.FS

func uiHandler() http.Handler {
	dist, err := fs.Sub(uiFiles, "ui/dist")
	if err != nil {
		log.Fatalf("Failed to load embedded UI: %v", err)
	}
	return http.FileServer(http.FS(dist))
}
