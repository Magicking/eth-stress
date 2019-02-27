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
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
)

type Nonce struct {
	c   chan uint64
	nv  chan uint64
	v   uint64
	end chan interface{}
}

func NewNonce(value uint64, deathPill chan interface{}) *Nonce {
	return &Nonce{
		c:   make(chan uint64),
		nv:  make(chan uint64),
		v:   value,
		end: deathPill,
	}
}

func (n *Nonce) Run() {
	for {
		select {
		case <-n.end:
			return
		default:
			n.c <- n.v
			ok := true
			for ok {
				select {
				case newValue := <-n.nv:
					if newValue < n.v {
						n.v = newValue - 1
					}
					continue
				default:
					ok = false
				}
			}
			n.v++
		}
	}
}

func (n *Nonce) Refresh(value uint64) bool {
	if value >= n.v {
		return false
	}
	n.nv <- value
	return true
}

func (n *Nonce) Next() uint64 {
	return <-n.c
}

type NonceManager struct {
	from      map[common.Address]*Nonce
	deathPill chan interface{}
	client    *ethclient.Client
	NetworkId *big.Int
}

func NewNonceManager(retry int, rpcurl string) (nm *NonceManager, err error) {
	var client *ethclient.Client
	for i := 0; i < retry; i++ {
		time.Sleep((2 << uint(i)) * time.Second)
		client, err = ethclient.Dial(rpcurl)
		if err == nil {
			break
		}
		if (i + 1) >= retry {
			return nil, err
		}
	}
	networkId, err := client.NetworkID(context.TODO())
	if err != nil {
		return nil, err
	}
	log.Println("Starting nonce manager on networkId:", networkId)
	return &NonceManager{
		from:      make(map[common.Address]*Nonce),
		client:    client,
		NetworkId: networkId,
		deathPill: make(chan interface{}),
	}, nil
}

func (nm *NonceManager) Close() {
	close(nm.deathPill)
	nm.client.Close()
}

func (nm *NonceManager) Add(from common.Address) error {
	balance, err := nm.client.BalanceAt(context.TODO(), from, nil)
	if err != nil {
		return err
	}
	nonce, err := nm.client.NonceAt(context.TODO(), from, nil)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"address": from.String(),
		"nonce":   nonce,
		"balance": balance,
	}).Info("Added account to NonceManager")
	nm.from[from] = NewNonce(nonce, nm.deathPill)
	go nm.from[from].Run()
	return nil
}

func (nm *NonceManager) RefreshNonce(from common.Address) error {
	nonce, err := nm.client.NonceAt(context.TODO(), from, nil)
	if err != nil {
		return err
	}
	if nm.from[from].Refresh(nonce) {
		log.WithFields(log.Fields{
			"address": from.String(),
			"nonce":   nonce,
		}).Info("Refresh nonce")
	}
	return nil
}

func (nm *NonceManager) NextNonce(from common.Address) uint64 {
	return nm.from[from].Next()
}
