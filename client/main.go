package main

import (
	"embed"
	"log"
	"net/http"

	"github.com/vjerci/gig/internal/settings"
)

//go:embed static
var staticFiles embed.FS

func main() {
	settings.Load()
	var staticFS = http.FS(staticFiles)
	fs := http.FileServer(staticFS)

	http.Handle("/static/", fs)

	log.Printf("Listening on %s...", settings.Http.Address)
	err := http.ListenAndServe(settings.Http.Address, nil)
	if err != nil {
		log.Fatal(err)
	}
}
