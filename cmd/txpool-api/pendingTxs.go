package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gorilla/websocket"
)

type PendingTx struct {
	BlockHash        interface{} `json:"blockHash"`
	BlockNumber      interface{} `json:"blockNumber"`
	From             string      `json:"from"`
	Gas              string      `json:"gas"`
	GasPrice         string      `json:"gasPrice"`
	Hash             string      `json:"hash"`
	Input            string      `json:"input"`
	Nonce            string      `json:"nonce"`
	To               string      `json:"to"`
	TransactionIndex interface{} `json:"transactionIndex"`
	Value            string      `json:"value"`
	Type             string      `json:"type"`
	V                string      `json:"v"`
	R                string      `json:"r"`
	S                string      `json:"s"`
}

func (p *txpoolGlobal) pendingTxsSubscribe(wg *sync.WaitGroup, conn *websocket.Conn, mt int) {
	defer wg.Done()

	method := "alchemy_pendingTransactions"
	txsChan := make(chan PendingTx, 1024)

	// Create subscription to new pending txs
	sub, err := p.config.rpcClient.EthSubscribe(context.Background(), txsChan,
		method,
		map[string]string{
			"toAddress": p.config.toAddress.String(),
		},
	)
	if err != nil {
		log.Printf("[%v] %v\n", p.config.uuid, fmt.Errorf("error ethSubscribe: %v", err))
		return
	}
	defer sub.Unsubscribe()
	log.Printf("[%v] Subscription \"%v\" registered, waiting for responses. toAddress: %v\n", p.config.uuid, method, p.config.toAddress.String())

	for {
		select {
		case tx := <-txsChan:
			// Add new pending txs to txpool
			accNonce, err := p.config.ethClientHTTPS.NonceAt(context.Background(), common.HexToAddress(tx.From), nil)
			if err != nil {
				log.Printf("[%v] %v\n", p.config.uuid, fmt.Errorf("nonce error: %v", err))
				continue
			}
			txNonce, err := hexutil.DecodeBig(tx.Nonce)
			if err != nil {
				log.Printf("[%v] %v\n", p.config.uuid, fmt.Errorf("txNonce DecodeBig error: %v", err))
				continue
			}
			log.Printf("[%v] New pending tx: %v\n", p.config.uuid, tx.Hash)

			if txNonce.Uint64() > accNonce {
				p.txpoolQueued.Store(tx.Hash, tx.From)
			} else {
				p.txpoolPending.Store(tx.Hash, tx.From)
			}

			pending := p.txpoolPending.Count()
			queued := p.txpoolQueued.Count()
			total := p.txpoolPending.Count() + p.txpoolQueued.Count()
			log.Printf("[%v] Pending txs total count: %v \n txpoolPending: %#v\n txpoolQueued: %#v\n", p.config.uuid, total, p.txpoolPending, p.txpoolQueued)

			// write txPool stats to ws
			txpoolStats := struct {
				TxpoolPending int `json:"pending"`
				TxpoolQueued  int `json:"queued"`
				TxpoolTotal   int `json:"total"`
			}{
				TxpoolPending: pending,
				TxpoolQueued:  queued,
				TxpoolTotal:   total,
			}
			err = conn.WriteJSON(txpoolStats)
			if err != nil {
				log.Printf("[%v] %v\n", p.config.uuid, fmt.Errorf("write error: %v", err))
				return
			}
		case err := <-sub.Err():
			if err != nil {
				log.Printf("[%v] %v\n", p.config.uuid, fmt.Errorf("subscription error: %v\n", err))
				return
			}
		case <-p.config.cancelChanP:
			log.Printf("[%v] PendingTxsSubscribe address update\n", p.config.uuid)
			return
		default:
			continue
		}
	}
}
