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

type txpoolGlobal struct {
	txpoolPending *txpool.Txpool
	txpoolQueued  *txpool.Txpool
	config        txpoolConfig
}

type txpoolConfig struct {
	uuid            uuid2.UUID
	toAddress       common.Address
	ethClientHTTPS  *ethclient.Client
	rpcClient       *rpc.Client
	cancelChanP     chan any
	cancelChanM     chan any
	cancelChanDropP chan any
	cancelChanDropQ chan any
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

	// WebSocket connection to eth node
	ethClientWS, err := rpc.Dial(*ethclientWS)
	if err != nil {
		log.Println("rpc dial:", fmt.Errorf("error: %v", err))
		return
	}
	defer ethClientWS.Close()

	// HTTPS connection to eth node
	ethClientHTTPS, err := ethclient.Dial(*ethclientHTTPS)
	if err != nil {
		log.Printf("[%v] %v\n", uuid, fmt.Errorf("error ethclient.Dial: %v", err))
	}
	defer ethClientHTTPS.Close()

	poolGlobal := txpoolGlobal{
		config: struct {
			uuid            uuid2.UUID
			toAddress       common.Address
			ethClientHTTPS  *ethclient.Client
			rpcClient       *rpc.Client
			cancelChanP     chan any
			cancelChanM     chan any
			cancelChanDropP chan any
			cancelChanDropQ chan any
		}{uuid: uuid, ethClientHTTPS: ethClientHTTPS, rpcClient: ethClientWS, cancelChanP: make(chan any), cancelChanM: make(chan any)},
	}

	wg := &sync.WaitGroup{}
	// Reading messages from client
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[%v] %v\n", uuid, fmt.Errorf("error read: %v", err))
			return
		}

		if poolGlobal.config.toAddress.String() != "0x0000000000000000000000000000000000000000" {
			poolGlobal.config.cancelChanP <- struct{}{}
			poolGlobal.config.cancelChanM <- struct{}{}
			poolGlobal.config.cancelChanDropP <- struct{}{}
			poolGlobal.config.cancelChanDropQ <- struct{}{}
		}

		poolGlobal.txpoolPending = txpool.New()
		poolGlobal.txpoolQueued = txpool.New()
		poolGlobal.config.toAddress = common.HexToAddress(string(message))
		log.Printf("[%v] New toAddress set: %v\n", uuid, poolGlobal.config.toAddress)

		// Pending txs monitoring
		wg.Add(1)
		go poolGlobal.pendingTxsSubscribe(wg, conn)
		wg.Add(1)
		go poolGlobal.minedTxsSubscribe(wg, conn)
		wg.Add(1)
		go poolGlobal.dropReplacedPendingTxs(wg)
		wg.Add(1)
		go poolGlobal.dropReplacedQueuedTxs(wg)
		// TODO: add method to clean up dropped txs (probably by comparing tx nonce and acc nonce)
	}
	wg.Wait()
}
