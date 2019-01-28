package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"

	flags "github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var done chan bool

var ethopts struct {
	RPCURL            string   `long:"ws-uri" env:"RPC_URL" description:"Quorum RPC URL (e.g: http://infura.io/...)"`
	Retry             int      `long:"retry" env:"RETRY" description:"Max connection retry"`
	To                string   `long:"to" env:"TO" description:"Address to send the payload"`
	Payload           string   `long:"payload" default:"00" env:"PAYLOAD" description:"Transaction payload"`
	PrivateFor        []string `long:"privateFor" env:"PRIVATE_FOR" description:"Base64 encoded public key"`
	PrivateKey        string   `long:"pkey" env:"PRIVATE_KEY" description:"Hex encoded private key"`
	MaxOpenConnection uint64   `long:"max-open-conn" default:"1" env:"MAX_OPEN_CONNECTION" description:"Maximum opened connection to Quorum"`
	MaxTransaction    uint64   `long:"max-tx" default:"1" env:"MAX_TRANSACTION" description:"Maximum transaction to send"`
}

func sendTransaction(counter *uint64) error {
	if atomic.LoadUint64(counter) >= ethopts.MaxTransaction {
		return nil
	}
	// craft transaction
	// connect with at least X retry
	// while counter < MaxTransaction
	//   send transaction
	//   atomic.AddUint64(counter, 1)
	return nil
}

var rootCmd = &cobra.Command{
	Use: "stress",
	Run: func(cmd *cobra.Command, args []string) {
		var wg sync.WaitGroup
		var counter uint64
		for i := uint64(0); i < ethopts.MaxOpenConnection && i < ethopts.MaxTransaction; i++ {
			wg.Add(1)
			go func() {
				if err := sendTransaction(&counter); err != nil {
					log.Errorln(err)
				}
				defer wg.Done()
			}()
		}
		wg.Wait()
	},
}

func main() {
	sigs := make(chan os.Signal, 1)
	done = make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		// Do exit stuff
		<-done
		log.Fatal("Exiting")
	}()
	go func() {
		sig := <-sigs
		log.WithFields(log.Fields{
			"signal": sig,
		}).Warning("Signal caught")
		done <- true
	}()

	_, err := flags.Parse(&ethopts)
	if err != nil {
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
