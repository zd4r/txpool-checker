package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gorilla/websocket"
)

type Block struct {
	Difficulty       string `json:"difficulty"`
	ExtraData        string `json:"extraData"`
	GasLimit         string `json:"gasLimit"`
	GasUsed          string `json:"gasUsed"`
	LogsBloom        string `json:"logsBloom"`
	Miner            string `json:"miner"`
	Nonce            string `json:"nonce"`
	Number           string `json:"number"`
	ParentHash       string `json:"parentHash"`
	ReceiptRoot      string `json:"receiptRoot"`
	Sha3Uncles       string `json:"sha3Uncles"`
	StateRoot        string `json:"stateRoot"`
	Timestamp        string `json:"timestamp"`
	TransactionsRoot string `json:"transactionsRoot"`
}

func (p *txpoolGlobal) minedTxsSubscribe(wg *sync.WaitGroup, conn *websocket.Conn, mt int) {
	defer wg.Done()

	method := "newHeads"
	blocksChan := make(chan Block, 8)

	// Create subscription to new blocks
	sub, err := p.config.rpcClient.EthSubscribe(context.Background(), blocksChan, method)
	if err != nil {
		log.Printf("[%v] %v\n", p.config.uuid, fmt.Errorf("error ethSubscribe: %v", err))
		return
	}
	defer sub.Unsubscribe()
	log.Printf("[%v] Subscription \"%v\" registered, waiting for responses. toAddress: %v \n", p.config.uuid, method, p.config.toAddress)

	for {
		select {
		case blockHeader := <-blocksChan:
			// Get block by its number
			blockNumber, err := hexutil.DecodeBig(blockHeader.Number)
			if err != nil {
				log.Printf("[%v] %v\n", p.config.uuid, fmt.Errorf("blockNumber decode error: %v", err))
			}
			log.Printf("[%v] New block: %v\n", p.config.uuid, blockNumber)

			block, err := p.config.ethClientHTTPS.BlockByNumber(context.Background(), blockNumber)
			if err != nil {
				log.Printf("[%v] %v\n", p.config.uuid, fmt.Errorf("block by number error: %v", err))
			}

			// Remove mined txs from txpool
			for _, tx := range block.Transactions() {
				if tx.To() != nil && tx.To().String() == p.config.toAddress.String() {
					p.txpoolPending.Delete(tx.Hash().String())
					p.txpoolQueued.Delete(tx.Hash().String())
				}
			}

			pending := p.txpoolPending.Count()
			queued := p.txpoolQueued.Count()
			total := p.txpoolPending.Count() + p.txpoolQueued.Count()
			log.Printf("[%v] Pending txs total count: %v \n txpoolPending: %#v\n txpoolQueued: %#v\n", p.config.uuid, total, p.txpoolPending, p.txpoolQueued)

			// write txsPool.Count() to ws
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
		case err = <-sub.Err():
			if err != nil {
				log.Printf("[%v] %v\n", p.config.uuid, fmt.Errorf("subscription error: %v\n", err))
				return
			}
		case <-p.config.cancelChanM:
			log.Printf("[%v] MinedTxsSubscribe address update\n", p.config.uuid)
			return
		default:
			continue
		}
	}
}
