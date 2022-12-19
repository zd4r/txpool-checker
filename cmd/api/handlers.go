package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"zd4rova/txpool-checker/internal/txpool"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	uuid2 "github.com/google/uuid"
)

type txpoolConfig struct {
	ethclient *ethclient.Client
	client    *rpc.Client
	txpool    *txpool.Txpool
	toAddress common.Address
	uuid      uuid2.UUID
}

func (app *Config) listenTxpool(w http.ResponseWriter, r *http.Request) {
	uuid := uuid2.New()

	// Upgrade connection to ws protocol
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[%v] %v\n", uuid, fmt.Errorf("error upgrade: %v", err))
		return
	}
	log.Printf("[%v] New connection\n", uuid)
	defer log.Printf("[%v] Client connections closed\n", uuid)
	defer conn.Close()

	// WebSocket connection to node
	client, err := rpc.Dial(*nodeWSUrl)
	if err != nil {
		log.Println("rpc dial:", fmt.Errorf("error: %v", err))
		return
	}
	defer client.Close()

	// ethclient Dial
	ethClient, err := ethclient.Dial(*nodeHTTPSUrl)
	if err != nil {
		log.Printf("[%v] %v\n", uuid, fmt.Errorf("error: %v", err))
	}

	pool := txpoolConfig{
		client:    client,
		ethclient: ethClient,
		uuid:      uuid,
	}

	// Reading messages from client
	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[%v] %v\n", uuid, fmt.Errorf("error read: %v", err))
			return
		}

		pool.toAddress = common.HexToAddress(string(message))
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
