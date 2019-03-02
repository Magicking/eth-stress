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
	"bytes"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var Done chan bool
var OpenedConnection int64
var TransactionKind int
var NM *NonceManager

const (
	kindUnsigned = 1 << iota
	kindSigned
	kindAsync
	kindRawPrivate
)

var Ethopts struct {
	RPCURL                 string   `long:"rpc-url" env:"RPC_URL" description:"Ethereum client WebSocket RPC URL"`
	Retry                  int      `long:"retry" env:"RETRY" description:"Max connection retry"`
	From                   string   `long:"from" env:"FROM" description:"Address of the emiter"`
	To                     string   `long:"to" env:"TO" description:"Address to send the payload"`
	Payload                string   `long:"payload" default:"00" env:"PAYLOAD" description:"Transaction payload"`
	PrivateFrom            string   `long:"privateFrom" env:"PRIVATE_FROM" description:"Base64 Quorum privateFrom encoded public key (from)"`
	PrivateFor             []string `long:"privateFor" env:"PRIVATE_FOR" description:"Base64 Quorum privateFor encoded public keys (to)"`
	PrivateKey             string   `long:"pkey" env:"PRIVATE_KEY" description:"Hex encoded private key"`
	MaxOpenConnection      int64    `long:"max-open-conn" default:"1" env:"MAX_OPEN_CONNECTION" description:"Maximum opened connection to ethereum client"`
	MaxTransaction         int64    `long:"max-tx" default:"1" env:"MAX_TRANSACTION" description:"Maximum transaction to send"`
	ABI                    string   `long:"abi" env:"ABI" description:"ABI to enable events watching"`
	ASync                  bool     `long:"async" env:"ASYNC" description:"Sending unsigned transaction with Quorum Async RPC"`
	ASyncAddr              string   `long:"async-addr" env:"ASYNC_ADDR" description:"Listening address of Async RPC callback server"`
	ASyncAdvertisedUrl     string   `long:"async-advertised-url" env:"ASYNC_ADVERTISED_URL" description:"ASync Callback URL"`
	TransactionManagerUrls []string `long:"enclave-manager-url" env:"TRANSACTION_MANAGER_URLS" description:"Transaction Manager Public HTTP API"`
}

// TransactionArgs represents the arguments for a transaction.
type TransactionArgs struct {
	From     common.Address  `json:"from"`
	To       *common.Address `json:"to"`
	Gas      hexutil.Big     `json:"gas"`
	GasPrice hexutil.Big     `json:"gasPrice"`
	Value    hexutil.Big     `json:"value"`
	Data     hexutil.Bytes   `json:"data"`
}

// TransactionArgsPrivate represents the arguments for private transaction
type TransactionArgsPrivate struct {
	TransactionArgs
	PrivateFrom string   `json:"privateFrom,omitempty"`
	PrivateFor  []string `json:"privateFor,omitempty"`
	CallbackUrl string   `json:"callbackUrl,omitempty"`
}

func (tx *TransactionArgsPrivate) SignedTransaction(transactor *bind.TransactOpts) []byte {
	nonce := NM.NextNonce(tx.From)
	_tx := types.NewTransaction(nonce, *tx.To, tx.Value.ToInt(), tx.Gas.ToInt().Uint64(), tx.GasPrice.ToInt(), tx.Data)
	signedTx, err := transactor.Signer(types.NewEIP155Signer(big.NewInt(NM.NetworkId.Int64())), transactor.From, _tx)
	if err != nil {
		log.Fatal(err)
	}
	// TODO Quorum private
	buf := bytes.NewBuffer(nil)
	err = signedTx.EncodeRLP(buf)
	if err != nil {
		log.Fatal(err)
	}
	return buf.Bytes()
}

/*
func (tx *TransactionArgsPrivate) PrivatePayload() {
gasPrice: 0,
gasLimit: 4300000,
to: "",
value: 0,
data: initializedDeploy,
from: decryptedAccount,
privateFrom: TM1_PUBLIC_KEY,
privateFor: TM2_PUBLIC_KEY,
isPrivate: true
}*/

func SendUnsignedTransaction(ec *rpc.Client, txArgs *TransactionArgsPrivate) (ret string, err error) {
	if txArgs.CallbackUrl == "" {
		return ret, ec.Call(&ret, "eth_sendTransaction", txArgs)
	}
	return ret, ec.Call(&ret, "eth_sendTransactionAsync", txArgs)
}

func SendSignedTransaction(ec *rpc.Client, txArgs *TransactionArgsPrivate, transactor *bind.TransactOpts) (ret string, err error) {
	//setPrivate V byte to 0x2X

	//
	// storeRaw w/ Async
	// sendRawTransaction w/ payload to new private hash
	rawtx := txArgs.SignedTransaction(transactor)
	return ret, ec.Call(&ret, "eth_sendRawTransaction", "0x"+common.Bytes2Hex(rawtx))
}

func sendTransaction(counter *int64, c chan string, startPill <-chan interface{}) error {
	<-startPill
	if atomic.LoadInt64(counter) >= Ethopts.MaxTransaction {
		return nil
	}
	var transactor *bind.TransactOpts
	if TransactionKind == kindSigned {
		key, err := crypto.HexToECDSA(Ethopts.PrivateKey)
		if err != nil {
			log.Fatal(err)
		}
		transactor = bind.NewKeyedTransactor(key)
	}
	// connect with at least X retry
	for retry := 0; retry < Ethopts.Retry && atomic.LoadInt64(counter) < Ethopts.MaxTransaction; retry++ {
		time.Sleep((2 << uint(retry)) * time.Second)
		client, err := rpc.Dial(Ethopts.RPCURL)
		if err != nil {
			if (retry + 1) >= Ethopts.Retry {
				return err
			}
			continue
		}
		atomic.AddInt64(&OpenedConnection, 1)
		defer func() {
			client.Close()
			atomic.AddInt64(&OpenedConnection, -1)
		}()
		transactArg := &TransactionArgsPrivate{
			TransactionArgs: TransactionArgs{
				From:     common.HexToAddress(Ethopts.From),
				Gas:      hexutil.Big(*big.NewInt(90000)),
				GasPrice: hexutil.Big{},
				Value:    hexutil.Big{},
				Data:     []byte("00")},
			PrivateFor: Ethopts.PrivateFor,
		}
		if to := common.HexToAddress(Ethopts.To); Ethopts.To != "" {
			transactArg.To = &to
		}
		for atomic.LoadInt64(counter) < Ethopts.MaxTransaction {
			atomic.AddInt64(counter, 1)
			var txHash string
			switch TransactionKind {
			case kindUnsigned:
				txHash, err = SendUnsignedTransaction(client, transactArg)
			case kindAsync:
				transactArg.CallbackUrl = Ethopts.ASyncAdvertisedUrl
				txHash, err = SendUnsignedTransaction(client, transactArg)
			case kindSigned:
				txHash, err = SendSignedTransaction(client, transactArg, transactor)
			}
			if err != nil {
				atomic.AddInt64(counter, -1)
				log.Println("SendTransaction", err)
				break
			}
			if TransactionKind == kindAsync {
				continue
			}
			c <- txHash
		}
	}
	return nil
}

var rootCmd = &cobra.Command{
	Use: "stress",
	Run: func(cmd *cobra.Command, args []string) {
		var wg sync.WaitGroup
		var counter int64
		var err error

		c := make(chan string)
		startPill := make(chan interface{})

		NM, err = NewNonceManager(Ethopts.Retry, Ethopts.RPCURL)
		if err != nil {
			log.Fatal(err)
		}
		defer func() { NM.Close() }()
		if Ethopts.PrivateKey != "" {
			key, err := crypto.HexToECDSA(Ethopts.PrivateKey)
			if err != nil {
				log.Fatal(err)
			}
			_from := crypto.PubkeyToAddress(key.PublicKey)
			err = NM.Add(_from)
			if err != nil {
				log.Fatal(err)
			}
			Ethopts.From = _from.String()
			TransactionKind = kindSigned
		}
		if Ethopts.From != "" && TransactionKind != kindSigned {
			TransactionKind = kindUnsigned
		}
		if Ethopts.ASync {
			if Ethopts.PrivateKey != "" {
				log.Fatal("Cannot send Quorum ASync signed transaction")
			}
			ascServer := NewASyncCallbackServer(Ethopts.ASyncAddr, c)
			go func() {
				log.Fatal(ascServer.Run())
			}()
			TransactionKind = kindAsync
		}
		log.Println("From", Ethopts.From)
		go func() {
			for i := int64(0); i < Ethopts.MaxOpenConnection && i < Ethopts.MaxTransaction; i++ {
				wg.Add(1)
				go func() {
					if err := sendTransaction(&counter, c, startPill); err != nil {
						log.Errorln(err)
					}
					defer wg.Done()
				}()
			}
			wg.Wait()
		}()
		if err := TxWatcher(c, startPill); err != nil {
			log.Fatal(err)
		}
		Done <- true
	},
}

func main() {
	sigs := make(chan os.Signal, 1)
	Done = make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		log.WithFields(log.Fields{
			"signal": sig,
		}).Warning("Signal caught")
		Done <- true
		time.Sleep(2 * time.Second)
		log.Fatal("Forced exit")
	}()

	rootCmd.PersistentFlags().StringVar(&Ethopts.RPCURL, "rpc-url", "ws://127.0.0.1:8546", "Ethereum client WebSocket RPC URL")
	rootCmd.PersistentFlags().StringVar(&Ethopts.From, "from", "", "Address of the emiter")
	rootCmd.PersistentFlags().StringVar(&Ethopts.To, "to", "", "Address to send the payload")
	rootCmd.PersistentFlags().StringVar(&Ethopts.Payload, "payload", "00", "Transaction payload")
	rootCmd.PersistentFlags().StringVar(&Ethopts.PrivateKey, "pkey", "", "Hex encoded private key")
	rootCmd.PersistentFlags().StringVar(&Ethopts.PrivateFrom, "privateFrom", "", "Base64 Quorum privateFrom encoded public key (from)")
	rootCmd.PersistentFlags().StringSliceVar(&Ethopts.PrivateFor, "privateFor", nil, "Base64 Quorum privateFor encoded public keys (to)")
	rootCmd.PersistentFlags().IntVar(&Ethopts.Retry, "retry", 3, "Max connection retry")
	rootCmd.PersistentFlags().Int64Var(&Ethopts.MaxOpenConnection, "max-open-conn", 1, "Maximum opened connection to ethereum client")
	rootCmd.PersistentFlags().Int64Var(&Ethopts.MaxTransaction, "max-tx", 1, "Maximum transaction to send")
	rootCmd.PersistentFlags().StringVar(&Ethopts.ABI, "abi", "", "ABI to enable events watching")
	rootCmd.PersistentFlags().BoolVar(&Ethopts.ASync, "async", false, "Sending unsigned transaction with Quorum Async RPC")
	rootCmd.PersistentFlags().StringVar(&Ethopts.ASyncAddr, "async-addr", ":18547", "Listening address of Async RPC callback server")
	rootCmd.PersistentFlags().StringVar(&Ethopts.ASyncAdvertisedUrl, "async-advertised-url", "http://localhost:18547/sendTransactionAsync", "ASync Callback URL")
	rootCmd.PersistentFlags().StringSliceVar(&Ethopts.TransactionManagerUrls, "transaction-manager-urls", []string{"http://127.0.0.1:9080"}, "Transaction Managers Public HTTP API")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
