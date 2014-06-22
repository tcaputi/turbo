package main

import (
	"flag"
	"go/build"
	"log"
	"net/http"
	"path/filepath"
	"text/template"
)

var (
	addr      = flag.String("addr", ":4000", "http service address")
	assets    = flag.String("assets", defaultAssetPath(), "path to assets")
	homeTempl *template.Template
)

func defaultAssetPath() string {
	p, err := build.Default.Import("gary.burd.info/go-websocket-chat", "", build.FindOnly)
	if err != nil {
		return "."
	}
	return p.Dir
}

func homeHandler(c http.ResponseWriter, req *http.Request) {
	homeTempl.Execute(c, req.Host)
}

func jsHandler(res http.ResponseWriter, req *http.Request) {
	http.ServeFile(res, req, "turbo.js")
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	flag.Parse()
	homeTempl = template.Must(template.ParseFiles(filepath.Join(*assets, "home.html")))
	go h.run()
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/turbo.js", jsHandler)
	http.HandleFunc("/ws", wsHandler)
	log.Println("Turbo sandbox server is now listening on 127.0.0.1:4000")
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
