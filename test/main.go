package main

import (
	"flag"
	"github.com/logmein3546/turbo"
	"github.com/skratchdot/open-golang/open"
	"log"
	"net/http"
	"path/filepath"
)

var (
	addr = flag.String("addr", ":4000", "http service address")
)

func main() {
	// Log config
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	// Make a turbo instance
	err, turboServer := turbo.New("mongodb://bitbeam.info:27017", "test", "entries")
	if err != nil {
		return
	}
	staticPath, err := filepath.Abs("./static")
	if err != nil {
		log.Println("Could not make a path to the static folder", err)
		return
	}
	log.Println("Referencing static turbo files at path", staticPath)
	indexPath, err := filepath.Abs("./test/test.html")
	if err != nil {
		log.Println("Could not make a path to test.html", err)
		return
	}
	// Register turbo handler
	http.HandleFunc("/ws", turboServer.Handler)
	// Register the static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath))))
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		http.ServeFile(res, req, indexPath)
	})
	// Open le browser
	open.Start("http://localhost:4000/")
	// Start le server
	log.Println("Turbo test server is now listening on 127.0.0.1:4000/ws")
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
