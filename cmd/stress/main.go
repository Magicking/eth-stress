package main

import (
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var Done chan bool
var OpenedConnection int64

var Ethopts struct {
	RPCURL            string   `long:"rpc-url" env:"RPC_URL" description:"Quorum RPC URL (e.g: http://kaleido.io/...)"`
	Retry             int      `long:"retry" env:"RETRY" description:"Max connection retry"`
	From              string   `long:"from" env:"FROM" description:"Address of the emiter"`
	To                string   `long:"to" env:"TO" description:"Address to send the payload"`
	Payload           string   `long:"payload" default:"00" env:"PAYLOAD" description:"Transaction payload"`
	PrivateFor        []string `long:"privateFor" env:"PRIVATE_FOR" description:"Base64 encoded public key"`
	PrivateKey        string   `long:"pkey" env:"PRIVATE_KEY" description:"Hex encoded private key"`
	MaxOpenConnection int64    `long:"max-open-conn" default:"1" env:"MAX_OPEN_CONNECTION" description:"Maximum opened connection to Quorum"`
	MaxTransaction    int64    `long:"max-tx" default:"1" env:"MAX_TRANSACTION" description:"Maximum transaction to send"`
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
	PrivateFor  []string `json:"privateFor"`
}

func SendUnsignedTransaction(ec *rpc.Client, txArgs *TransactionArgsPrivate) (string, error) {
	var ret string
	return ret, ec.Call(&ret /*TODO*/, "eth_sendTransaction", txArgs)
}

func sendTransaction(counter *int64, c chan string) error {
	if atomic.LoadInt64(counter) >= Ethopts.MaxTransaction {
		return nil
	}
	// craft transaction
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
			ret, err := SendUnsignedTransaction(client, transactArg)
			if err != nil {
				atomic.AddInt64(counter, -1)
				log.Println(err)
				break
			}
			c <- ret
		}
	}
	return nil
}

var rootCmd = &cobra.Command{
	Use: "stress",
	Run: func(cmd *cobra.Command, args []string) {
		var wg sync.WaitGroup
		var counter int64

		c := make(chan string)
		go func() {
			for i := int64(0); i < Ethopts.MaxOpenConnection && i < Ethopts.MaxTransaction; i++ {
				wg.Add(1)
				go func() {
					if err := sendTransaction(&counter, c); err != nil {
						log.Errorln(err)
					}
					defer wg.Done()
				}()
			}
			wg.Wait()
		}()
		if err := TXWatcher(c); err != nil {
			log.Fatal(err)
		}
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

	rootCmd.PersistentFlags().StringVar(&Ethopts.RPCURL, "rpc-url", "http://127.0.0.1:8545", "Quorum RPC URL (e.g: http://kaleido.io/...)")
	rootCmd.PersistentFlags().StringVar(&Ethopts.From, "from", "", "Address of the emiter")
	rootCmd.PersistentFlags().StringVar(&Ethopts.To, "to", "", "Address to send the payload")
	rootCmd.PersistentFlags().StringVar(&Ethopts.Payload, "payload", "00", "Transaction payload")
	rootCmd.PersistentFlags().StringVar(&Ethopts.PrivateKey, "pkey", "", "Hex encoded private key")
	rootCmd.PersistentFlags().StringSliceVar(&Ethopts.PrivateFor, "privateFor", nil, "Base64 encoded public key")
	rootCmd.PersistentFlags().IntVar(&Ethopts.Retry, "retry", 3, "Max connection retry")
	rootCmd.PersistentFlags().Int64Var(&Ethopts.MaxOpenConnection, "max-open-conn", 1, "Maximum opened connection to Quorum")
	rootCmd.PersistentFlags().Int64Var(&Ethopts.MaxTransaction, "max-tx", 1, "Maximum transaction to send")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	<-Done
}
