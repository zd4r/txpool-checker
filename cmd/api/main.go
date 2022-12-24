package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// initial variables
var addr = flag.String("addr", ":8081", "http service address")
var ethclientWS = flag.String("nodeWs", "wss://eth-mainnet.g.alchemy.com/v2/ONdYH5RobUhuIa963-uQcVqT1R1DBKiM", "node ws url")
var ethclientHTTPS = flag.String("nodeHTTPS", "https://eth-mainnet.g.alchemy.com/v2/ONdYH5RobUhuIa963-uQcVqT1R1DBKiM", "node ws url")

// WebSocket upgrader set up
var origins = []string{"http://localhost:8080", "http://217.151.231.111:8080/"}
var upgrader = websocket.Upgrader{
	// Resolve cross-domain problems
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("origin")
		for _, allowOrigin := range origins {
			if origin == allowOrigin {
				return true
			}
		}
		return false
	},
}

type Config struct {
}

func main() {
	flag.Parse()

	app := Config{}

	// Starting web server
	http.HandleFunc("/txpool", app.listenTxpool)

	log.Println("Server started localhost:8081")
	log.Fatal(http.ListenAndServe(*addr, nil))
}
