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

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
)

func TxWatcher(txChan <-chan string, startPill chan interface{}) (err error) {
	txMap := make(map[string]bool)
	txTimestamps := make(map[string]time.Time) // Track when transactions were sent
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

	log.WithFields(log.Fields{
		"time": lastCount,
		"url":  Ethopts.RPCURL,
	}).Info("Starting transaction watcher")
	close(startPill)

	for {
		select {
		case tx := <-txChan:
			sent++
			txMap[tx] = true
			txTimestamps[tx] = time.Now()
			log.WithFields(log.Fields{
				"txHash":    tx,
				"totalSent": sent,
				"timestamp": time.Now(),
			}).Debug("Transaction sent")
		case <-Done:
			Done <- true
			return
		case <-ticker.C:
			// Check pending transactions
			for txHash, pending := range txMap {
				if !pending {
					continue
				}

				tx, isPending, err := client.TransactionByHash(context.TODO(), common.HexToHash(txHash))
				if err != nil {
					if err.Error() != "not found" {
						log.WithError(err).WithField("txHash", txHash).Error("Failed to get transaction")
					}
					continue
				}

				if !isPending {
					txMap[txHash] = false
					seen++
					if sentTime, ok := txTimestamps[txHash]; ok {
						confirmationTime := time.Since(sentTime)
						log.WithFields(log.Fields{
							"txHash":             txHash,
							"confirmationTimeMs": confirmationTime.Milliseconds(),
							"gasUsed":            tx.Gas(),
						}).Info("Transaction confirmed")
						delete(txTimestamps, txHash)
					}
				}
			}

			// Get current block number for stats
			currentBlock, err := client.BlockNumber(context.TODO())
			if err != nil {
				log.WithError(err).Error("Failed to get current block number")
				continue
			}

			timeSpent := time.Since(lastCount).Seconds()
			lastCount = time.Now()
			txpsSeen = float64(seen-lastSeen) / timeSpent
			lastSeen = seen
			txpsSent = float64(sent-lastSent) / timeSpent
			lastSent = sent

			// Update stats
			maxBlock = currentBlock
			txAvg = append(txAvg, txpsSeen)
			var diff float64
			for _, e := range txAvg {
				diff += e
			}
			diff /= float64(len(txAvg))
			if len(txAvg) > 10 {
				txAvg = txAvg[1:10]
			}

			// Check for long-pending transactions
			now := time.Now()
			for txHash, sentTime := range txTimestamps {
				pendingDuration := now.Sub(sentTime)
				if pendingDuration > 5*time.Minute {
					log.WithFields(log.Fields{
						"txHash":             txHash,
						"pendingDurationMin": pendingDuration.Minutes(),
					}).Warning("Transaction pending for extended period")
				}
			}

			log.WithFields(log.Fields{
				"seenTxPerSecAvg": fmt.Sprintf("%.02f", diff),
				"seenTxPerSec":    fmt.Sprintf("%.02f", txpsSeen),
				"sentTxPerSec":    fmt.Sprintf("%.02f", txpsSent),
				"blockNumber":     maxBlock,
				"connections":     OpenedConnection,
				"seenTx":          seen,
				"sentTx":          sent,
				"pendingTx":       len(txTimestamps),
			}).Info("Stats update")
		}
	}
}
