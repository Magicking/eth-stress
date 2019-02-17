# Ethereum transaction stresser

## Description

A WIP CLI ethereum client transaction stresser

## Usage

```
Usage:
  stress [flags]

Flags:
      --from string          Address of the emiter
  -h, --help                 help for stress
      --max-open-conn int    Maximum opened connection to Quorum (default 1)
      --max-tx int           Maximum transaction to send (default 1)
      --payload string       Transaction payload (default "00")
      --pkey string          Hex encoded private key
      --privateFor strings   Base64 encoded public key
      --retry int            Max connection retry (default 3)
      --rpc-url string       Quorum RPC URL (e.g: http://kaleido.io/...) (default "http://127.0.0.1:8545")
      --to string            Address to send the payload
```
## Install

 - [Golang](https://golang.org/doc/install)

Then

```
go get github.com/Magicking/quorum-stress/cmd/stress
stress --help
```

## Example

```
go get github.com/Magicking/quorum-stress/cmd/stress

stress --from 0x6120a30e955b6dd99c9adc6e1ece6dcc6d48a53f --to 0x6120a7e00f3b2937362dfdac9f80b79f5b55f165 --rpc-url ws://127.0.0.1:8546 --max-tx 20000 --max-open-conn 200

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

 - [x] go-ethereum
 - [x] Quorum
 - [x] Parity (to be tested)
 - [ ] Ganache
 - [ ] Pantheon
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

 - [x] sendTransaction
 - [x] sendTransactionAsync (quorum)
 - [x] sendRawTransaction
 - [ ] sendRawTransaction (quorum)
 - [ ] sendRawPrivateTransaction (quorum)

## Status report

 - [x] CLI
 - [ ] JSON
 - [ ] CSV
