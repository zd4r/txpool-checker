package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/gorilla/websocket"
)

// initial variables
var addr = flag.String("addr", "localhost:8081", "http service address")
var nodeUrl = flag.String("node", "wss://eth-mainnet.g.alchemy.com/v2/ONdYH5RobUhuIa963-uQcVqT1R1DBKiM", "node ws url")

// WebSocket upgrader set up
var origins = []string{"http://localhost:8080"}
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
	}}

type Config struct {
	NodeClient *rpc.Client
}

func main() {
	flag.Parse()

	app := Config{}

	// WebSocket connection to node
	client, err := rpc.Dial(*nodeUrl)
	if err != nil {
		log.Println("rpc dial:", fmt.Errorf("error: %v", err))
		return
	}
	defer client.Close()

	// Starting web server
	app.NodeClient = client
	log.Println("Node connection established")

	http.HandleFunc("/txpool", app.listenTxpool)

	log.Println("Server started localhost:8081")
	log.Fatal(http.ListenAndServe(*addr, nil))
}
