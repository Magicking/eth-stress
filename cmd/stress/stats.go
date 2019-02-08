package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
)

func TXWatcher(txChan <-chan string) (err error) {
	txMap := make(map[string]bool)
	ticker := time.NewTicker(1 * time.Second)
	counter := 0
	seen := 0
	lastSeen := 0
	var txAvg []float64
	var txpersec float64
	var client *ethclient.Client
	var maxBlock uint64
	lastCount := time.Now()

	for retry := 0; retry < Ethopts.Retry; retry++ {
		time.Sleep((2 << uint(retry)) * time.Second)
		client, err = ethclient.Dial(Ethopts.RPCURL)
		if err != nil {
			if (retry + 1) == Ethopts.Retry {
				return err
			}
			continue
		}
	}
	defer client.Close()
	bChan := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(context.TODO(), bChan)
	if err != nil {
		log.Fatalf("Could not register for event: %v", err)
	}
	log.Println("Start time", lastCount)
	for {
		select {
		case tx := <-txChan:
			counter++
			txMap[tx] = true
		case <-Done:
			Done <- true
			return
		case b := <-bChan:
			blk, err := client.BlockByHash(context.TODO(), b.ParentHash)
			if err != nil {
				log.Println(err)
				continue
			}
			txs := blk.Transactions()
			for _, tx := range txs {
				if txMap[tx.Hash().Hex()] {
					seen++
				}
			}
			log.WithFields(log.Fields{
				"Hash":       blk.Hash().TerminalString(),
				"cb":         blk.Coinbase().Hex(),
				"difficulty": blk.Difficulty(),
				//				"extra":      hex.EncodeToString(b.Extra),
				"gasLimit":   blk.GasLimit(),
				"gasUsed":    blk.GasUsed(),
				"n":          blk.Number(),
				"nTx":        blk.Transactions().Len(),
				"chain time": time.Unix(blk.Time().Int64(), 0),
			}).Info()
			timeSpent := time.Since(lastCount).Seconds()
			lastCount = time.Now()
			txpersec = float64(seen-lastSeen) / timeSpent
			lastSeen = seen
		case err := <-sub.Err():
			log.Println(err)
			Done <- true
		case <-ticker.C:
			b, err := client.BlockByNumber(context.TODO(), nil)
			if err != nil {
				log.Println(err) //TODO
				continue
			}
			// for each last to max
			// get lastblock
			maxBlock = b.NumberU64()
			txAvg = append(txAvg, txpersec)
			var diff float64
			for _, e := range txAvg {
				diff += e
			}
			diff /= float64(len(txAvg))
			if len(txAvg) > 10 {
				txAvg = txAvg[1:10]
			}
			log.WithFields(log.Fields{
				"tx/s avg":     fmt.Sprintf("%.02f", diff),
				"tx/s":         fmt.Sprintf("%.02f", txpersec),
				"block number": maxBlock,
				"connection":   OpenedConnection,
				"seen tx":      seen,
				"sent tx":      counter,
			}).Info()
		}
	}
}
