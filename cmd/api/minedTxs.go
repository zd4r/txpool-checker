package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
)

type MinedTx struct {
	Removed     bool `json:"removed"`
	Transaction struct {
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
	} `json:"transaction"`
}

func (p *txpoolConfig) minedTxsSubscribe(wg *sync.WaitGroup, conn *websocket.Conn, mt int) {
	defer wg.Done()

	method := "alchemy_minedTransactions"
	txsChan := make(chan MinedTx, 1024)

	sub, err := p.client.EthSubscribe(context.Background(), txsChan,
		method,
		map[string][]map[string]string{
			"addresses": {
				{
					"to": p.toAddress,
				},
			},
		},
	)
	if err != nil {
		log.Printf("[%v] %v\n", p.uuid, fmt.Errorf("error ethSubscribe: %v", err))
		return
	}
	defer sub.Unsubscribe()
	log.Printf("[%v] Subscription \"%v\" registered, waiting for responses. toAddress: %v\n", p.uuid, method, p.toAddress)

	for {
		select {
		case tx := <-txsChan:
			log.Printf("[%v] New mined tx: %v\n", p.uuid, tx.Transaction.Hash)
			p.txpool.Delete(tx.Transaction.Hash)
			log.Printf("[%v] Pending txs count: %v | MAP: %#v\n", p.uuid, p.txpool.Count(), p.txpool)

			// write txsPool.Count() to ws
			message := []byte(strconv.Itoa(p.txpool.Count()))
			err = conn.WriteMessage(mt, message)
			if err != nil {
				log.Printf("[%v] %v\n", p.uuid, fmt.Errorf("write error: %v", err))
				return
			}
		case err = <-sub.Err():
			if err != nil {
				log.Printf("[%v] %v\n", p.uuid, fmt.Errorf("subscription error: %v\n", err))
				return
			}
		default:
			continue
		}
	}
}
