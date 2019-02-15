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
	v   uint64
	end chan interface{}
}

func NewNonce(value uint64, deathPill chan interface{}) *Nonce {
	return &Nonce{
		c:   make(chan uint64),
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
			n.v++
		}
	}
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
		if err != nil {
			if (i + 1) >= retry {
				return nil, err
			}
			continue
		}
	}
	networkId, err := client.NetworkID(context.TODO())
	if err != nil {
		return nil, err
	}
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
	nonce, err := nm.client.NonceAt(context.TODO(), from, nil)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"address": from.String(),
		"nonce":   nonce,
	}).Info("Added account to NonceManager")
	nm.from[from] = NewNonce(nonce, nm.deathPill)
	go nm.from[from].Run()
	return nil
}

func (nm *NonceManager) NextNonce(from common.Address) uint64 {
	return nm.from[from].Next()
}
