package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"zd4rova/txpool-checker/internal/txpool"

	"github.com/ethereum/go-ethereum/rpc"
	uuid2 "github.com/google/uuid"
)

type txpoolConfig struct {
	client    *rpc.Client
	txpool    *txpool.Txpool
	toAddress string
	uuid      uuid2.UUID
}

func (app *Config) listenTxpool(w http.ResponseWriter, r *http.Request) {
	uuid := uuid2.New()
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[%v] %v\n", uuid, fmt.Errorf("error upgrade: %v", err))
		return
	}
	log.Printf("[%v] New connection\n", uuid)
	defer log.Printf("[%v] Connections closed\n", uuid)
	defer conn.Close()

	pool := txpoolConfig{
		client: app.NodeClient,
		uuid:   uuid,
	}

	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[%v] %v\n", uuid, fmt.Errorf("error read: %v", err))
			return
		}

		pool.toAddress = string(message)
		log.Printf("[%v] New toAddress set: %v\n", uuid, pool.toAddress)

		// Pending txs monitoring
		pool.txpool = txpool.New()
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go pool.pendingTxsSubscribe(wg, conn, mt)
		wg.Add(1)
		go pool.minedTxsSubscribe(wg, conn, mt)
		wg.Wait()
	}
}
