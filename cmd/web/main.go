package main

import (
	"flag"
	"log"
	"net/http"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func home(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./app/index.html")
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/", home)
	log.Println("Server started localhost:8080")
	log.Fatal(http.ListenAndServe(*addr, nil))
}
