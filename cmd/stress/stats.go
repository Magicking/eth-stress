// Copyright (C) 2019 Sylvain 6120 Laurent
// This file is part of eth-stress <https://github.com/Magicking/eth-stress>.
//
// eth-stress is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// eth-stress is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with eth-stress.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
)

func TxWatcher(txChan <-chan string, startPill chan interface{}) (err error) {
	txMap := make(map[string]bool)
	ticker := time.NewTicker(1 * time.Second)
	sent := 0
	seen := 0
	lastSeen := 0
	lastSent := 0
	var txAvg []float64
	var txpsSeen float64
	var txpsSent float64
	var client *ethclient.Client
	var maxBlock uint64
	lastCount := time.Now()

	for retry := 0; retry < Ethopts.Retry; retry++ {
		time.Sleep((2 << uint(retry%3)) * time.Second)
		client, err = ethclient.Dial(Ethopts.RPCURL)
		if err == nil {
			break
		}
		if (retry + 1) == Ethopts.Retry {
			return err
		}
	}
	defer client.Close()
	bChan := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(context.TODO(), bChan)
	if err != nil {
		log.Fatalf("Could not register for event: %v", err)
	}
	log.Println("Start time", lastCount)
	close(startPill)
	for {
		select {
		case tx := <-txChan:
			sent++
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
					txMap[tx.Hash().Hex()] = false
					seen++
				}
			}
			log.WithFields(log.Fields{
				"block number": blk.Number(),
				"hash":         blk.Hash().TerminalString(),
				"difficulty":   blk.Difficulty(),
				//				"extra":      hex.EncodeToString(b.Extra),
				"gasLimit": blk.GasLimit(),
				"gasUsed":  blk.GasUsed(),
				"nTx":      blk.Transactions().Len(),
				//				"cb":         blk.Coinbase().Hex(),
				"block time": time.Unix(blk.Time().Int64(), 0),
			}).Info("new block")
			timeSpent := time.Since(lastCount).Seconds()
			lastCount = time.Now()
			txpsSeen = float64(seen-lastSeen) / timeSpent
			lastSeen = seen
			txpsSent = float64(sent-lastSent) / timeSpent
			lastSent = sent
		case err := <-sub.Err():
			log.Println(err)
			Done <- true
		case <-ticker.C:
			b, err := client.BlockByNumber(context.TODO(), nil) // TODO: Maximum 1 second of network context
			if err != nil {
				log.Println(err) //TODO
				continue
			}
			maxBlock = b.NumberU64()
			txAvg = append(txAvg, txpsSeen)
			var diff float64
			for _, e := range txAvg {
				diff += e
			}
			diff /= float64(len(txAvg))
			if len(txAvg) > 10 {
				txAvg = txAvg[1:10]
			}
			log.WithFields(log.Fields{
				"seen tx/s avg": fmt.Sprintf("%.02f", diff),
				"seen tx/s":     fmt.Sprintf("%.02f", txpsSeen),
				"sent tx/s":     fmt.Sprintf("%.02f", txpsSent),
				"block number":  maxBlock,
				"connection":    OpenedConnection,
				"seen tx":       seen,
				"sent tx":       sent,
			}).Info()
		}
	}
}
