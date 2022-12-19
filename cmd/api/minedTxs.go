package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
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

func (p *txpoolConfig) minedTxsSubscribe(wg *sync.WaitGroup, conn *websocket.Conn, mt int) {
	defer wg.Done()

	method := "newHeads"
	blocksChan := make(chan Block, 1024)

	// Create subscription to new blocks
	sub, err := p.client.EthSubscribe(context.Background(), blocksChan, method)
	if err != nil {
		log.Printf("[%v] %v\n", p.uuid, fmt.Errorf("error ethSubscribe: %v", err))
		return
	}
	defer sub.Unsubscribe()
	log.Printf("[%v] Subscription \"%v\" registered, waiting for responses. toAddress: %v \n", p.uuid, method, p.toAddress)

	for {
		select {
		case blockHeader := <-blocksChan:
			// Get block by its number
			blockNumber, err := hexutil.DecodeBig(blockHeader.Number)
			if err != nil {
				log.Printf("[%v] %v\n", p.uuid, fmt.Errorf("blockNumber decode error: %v", err))
			}
			log.Printf("[%v] New block: %v\n", p.uuid, blockNumber)

			block, err := p.ethclient.BlockByNumber(context.Background(), blockNumber)
			if err != nil {
				log.Printf("[%v] %v\n", p.uuid, fmt.Errorf("block by number error: %v", err))
			}

			// Remove mined txs from txpool
			for _, tx := range block.Transactions() {
				if tx.To() != nil && tx.To().String() == p.toAddress.String() {
					p.txpool.Delete(tx.Hash().String())
				}
			}
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
