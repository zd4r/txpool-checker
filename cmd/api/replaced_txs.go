package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

func (p *txpoolGlobal) dropReplacedPendingTxs(wg *sync.WaitGroup) {
	defer wg.Done()

	f := func(txHash, fromAddress string) bool {
		accNonce, err := p.config.ethClientHTTPS.NonceAt(context.Background(), common.HexToAddress(fromAddress), nil)
		if err != nil {
			log.Printf("[%v] %v\n", p.config.uuid, fmt.Errorf("account nonce error: %v", err))
			return true
		}

		tx, _, err := p.config.ethClientHTTPS.TransactionByHash(context.Background(), common.HexToHash(txHash))
		if err != nil {
			log.Printf("[%v] %v\n", p.config.uuid, fmt.Errorf("transactionByHash [%v] error: %v", txHash, err))
			return true
		}

		if accNonce > tx.Nonce() {
			log.Printf("[%v] %v\n", p.config.uuid, fmt.Sprintf("replaced tx drop: %v", txHash))
			p.txpoolPending.Delete(txHash)
		}
		return true
	}

	for {
		select {
		case <-p.config.cancelChanDropP:
			log.Printf("[%v] Stop drop pending txs for current pool due to address update\n", p.config.uuid)
			return
		default:
			p.txpoolPending.Range(f)
		}
	}
}

func (p *txpoolGlobal) dropReplacedQueuedTxs(wg *sync.WaitGroup) {
	defer wg.Done()

	f := func(txHash, fromAddress string) bool {
		accNonce, err := p.config.ethClientHTTPS.NonceAt(context.Background(), common.HexToAddress(fromAddress), nil)
		if err != nil {
			log.Printf("[%v] %v\n", p.config.uuid, fmt.Errorf("account nonce error: %v", err))
			return true
		}

		tx, _, err := p.config.ethClientHTTPS.TransactionByHash(context.Background(), common.HexToHash(txHash[:2]))
		if err != nil {
			log.Printf("[%v] %v\n", p.config.uuid, fmt.Errorf("transactionByHash error: %v", err))
			return true
		}

		if accNonce > tx.Nonce() {
			log.Printf("[%v] %v\n", p.config.uuid, fmt.Sprintf("replaced tx drop: %v", txHash))
			p.txpoolQueued.Delete(txHash)
		}
		return true
	}

	for {
		select {
		case <-p.config.cancelChanDropQ:
			log.Printf("[%v] Stop drop queued txs for current pool due to address update\n", p.config.uuid)
			return
		default:
			p.txpoolQueued.Range(f)
		}
	}
}
