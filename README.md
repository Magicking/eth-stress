# Ethereum transaction stresser

## Description

A CLI ethereum client transaction stresser

## Usage

```
Usage:
  stress [flags]

Flags:
      --abi string                    ABI to enable events watching
      --async                         Sending unsigned transaction with Quorum Async RPC
      --async-addr string             Listening address of Async RPC callback server (default ":18547")
      --async-advertised-url string   ASync Callback URL (default "http://localhost:18547/sendTransactionAsync")
      --from string                   Address of the emiter
  -h, --help                          help for stress
      --max-open-conn int             Maximum opened connection to ethereum client (default 1)
      --max-tx int                    Maximum transaction to send (default 1)
      --payload string                Transaction payload (default "00")
      --pkey string                   Hex encoded private key
      --privateFor strings            Base64 encoded public key
      --retry int                     Max connection retry (default 3)
      --rpc-url string                Ethereum client WebSocket RPC URL (default "ws://127.0.0.1:8546")
      --to string                     Address to send the payload
```
## Install

 - [Golang](https://golang.org/doc/install)
or
 - [Docker](https://docs.docker.com/compose/install/)

Then

```
go get github.com/Magicking/quorum-stress/cmd/stress
stress --help
```

## Example

```
docker-compose run --rm stress --from 0x6120a30e955b6dd99c9adc6e1ece6dcc6d48a53f --to 0x6120a7e00f3b2937362dfdac9f80b79f5b55f165 --rpc-url ws://192.168.0.1:8546 --max-tx 20000 --max-open-conn 200

INFO[0014] Start time 2019-02-08 02:24:38.872591 +0100 CET m=+0.014778258 
INFO[0121] new block                                     block number=111 block time="2019-02-15 02:26:48 +0100 CET" difficulty=2 gasLimit=70748612822601 gasUsed=64295712 hash="1921b2…5130ee" nTx=3042
INFO[0121]                                               block number=111 connection=100 seen tx=92159 seen tx/s=1265.67 seen tx/s avg=1011.68 sent tx=94343 sent tx/s=907.86
INFO[0122]                                               block number=111 connection=100 seen tx=92159 seen tx/s=1265.67 seen tx/s avg=1034.77 sent tx=95418 sent tx/s=907.86
INFO[0123] new block                                     block number=112 block time="2019-02-15 02:26:50 +0100 CET" difficulty=2 gasLimit=70679522474576 gasUsed=46139888 hash="7beba7…495a21" nTx=2183
INFO[0123]                                               block number=112 connection=100 seen tx=94342 seen tx/s=1252.77 seen tx/s avg=1046.70 sent tx=95419 sent tx/s=617.49
INFO[0124]                                               block number=112 connection=100 seen tx=94342 seen tx/s=1252.77 seen tx/s avg=1065.44 sent tx=97418 sent tx/s=617.49
INFO[0124] new block                                     block number=113 block time="2019-02-15 02:26:52 +0100 CET" difficulty=2 gasLimit=70610499570998 gasUsed=22742336 hash="93deed…dab874" nTx=1076
INFO[0125]                                               block number=113 connection=100 seen tx=95418 seen tx/s=669.31 seen tx/s avg=985.07 sent tx=98400 sent tx/s=1244.07
```

## Clients

 - [ ] [Pantheon](https://github.com/PegaSysEng/pantheon/blob/master/docs/index.md#what-is-pantheon)
 - [x] [go-ethereum](https://github.com/ethereum/go-ethereum/wiki/Command-Line-Options)
 - [x] [Quorum](https://github.com/jpmorganchase/quorum/wiki/Using-Quorum)
 - [x] [Parity](https://wiki.parity.io/Basic-Usage) (to be tested)
 - [ ] Ganache
 - [x] Infura (to be tested)
 - [ ] Ethereum 2.0 client

## Transaction type

 - [x] Public
 - [x] Private
 - [ ] Contract

## Endpoint type

 - [x] Websocket
 - [ ] HTTP (legacy)

## RPC

 - [x] [sendTransaction](https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_sendtransaction)
 - [x] [sendTransactionAsync](https://github.com/jpmorganchase/quorum/blob/master/docs/api.md#eth_sendtransactionasync) (quorum)
 - [x] [sendRawTransaction](https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_sendrawtransaction)
 - [ ] [sendRawPrivateTransaction](https://github.com/jpmorganchase/quorum/blob/master/docs/api.md#ethsendrawprivatetransaction) (quorum)
 - [ ] [storeraw](https://github.com/jpmorganchase/tessera/wiki/Interface-&-API#third-party-http-public-api) (quorum/tessera)

## Status report

 - [x] CLI
 - [ ] JSON
 - [ ] CSV

## Output

 * `seen tx/s` is number of transaction seen per second since last block
 * `seen tx/s avg` is average of last 10 secs of `seen tx/s`
 * `sent tx/s` is number of transaction sent per second since last block
 * `block number` is last processed block number
 * `connection` is number of transacting connection open
 * `seen tx` is number of seen tx
 * `sent tx` is number of sent tx
