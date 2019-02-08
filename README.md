# Ethereum stresser

## Description

A CLI ethereum client stresser with emphasis on privacy enabled clients

## Example

```
stress --from 0x6120a30e955b6dd99c9adc6e1ece6dcc6d48a53f --to 0x6120a7e00f3b2937362dfdac9f80b79f5b55f165 --rpc-url ws://127.0.0.1:8546 --max-tx 20000 --max-open-conn 200

INFO[0014] Start time 2019-02-08 02:24:38.872591 +0100 CET m=+0.014778258 
INFO[0014] block number=652 connection=200 seen tx=0 sent tx=4 tx/s=0.00 tx/s avg=0.00
INFO[0014] Hash="156cbe…13d374" cb=0x0000000000000000000000000000000000000000 chain time="2019-02-08 02:24:51 +0100 CET" difficulty=2 gasLimit=2606389269495 gasUsed=0 n=652 nTx=0
INFO[0015] block number=653 connection=200 seen tx=0 sent tx=1871 tx/s=0.00 tx/s avg=0.00
INFO[0016] block number=653 connection=200 seen tx=0 sent tx=4019 tx/s=0.00 tx/s avg=0.00
INFO[0016] Hash="47a0b2…7bce0e" cb=0x0000000000000000000000000000000000000000 chain time="2019-02-08 02:24:53 +0100 CET" difficulty=2 gasLimit=2603843967476 gasUsed=0 n=653 nTx=0
INFO[0017] block number=654 connection=200 seen tx=0 sent tx=5406 tx/s=0.00 tx/s avg=0.00
INFO[0018] block number=654 connection=200 seen tx=0 sent tx=5634 tx/s=0.00 tx/s avg=0.00
INFO[0018] Hash="9b89ac…66a4c4" cb=0x0000000000000000000000000000000000000000 chain time="2019-02-08 02:24:55 +0100 CET" difficulty=2 gasLimit=2601301151103 gasUsed=169088 n=654 nTx=8
INFO[0019] block number=655 connection=200 seen tx=8 sent tx=7354 tx/s=3.93 tx/s avg=0.66
INFO[0020] block number=655 connection=200 seen tx=8 sent tx=8632 tx/s=3.93 tx/s avg=1.12
INFO[0021] Hash="cdd46a…7dc885" cb=0x0000000000000000000000000000000000000000 chain time="2019-02-08 02:24:57 +0100 CET" difficulty=2 gasLimit=2598760818196 gasUsed=86150336 n=655 nTx=4076
INFO[0021] block number=656 connection=200 seen tx=4084 sent tx=8636 tx/s=1193.20 tx/s avg=150.13
INFO[0022] block number=656 connection=200 seen tx=4084 sent tx=9496 tx/s=1193.20 tx/s avg=266.03
INFO[0022] Hash="0d1eb0…e19127" cb=0x0000000000000000000000000000000000000000 chain time="2019-02-08 02:24:59 +0100 CET" difficulty=2 gasLimit=2596223092032 gasUsed=29315632 n=656 nTx=1387
INFO[0023] block number=657 connection=200 seen tx=5471 sent tx=10804 tx/s=1393.48 tx/s avg=378.77
INFO[0024] block number=657 connection=200 seen tx=5471 sent tx=12340 tx/s=1393.48 tx/s avg=471.02
INFO[0025] Hash="e50506…533ed9" cb=0x0000000000000000000000000000000000000000 chain time="2019-02-08 02:25:01 +0100 CET" difficulty=2 gasLimit=2593687760862 gasUsed=66810896 n=657 nTx=3161

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
 - [ ] Private

## Endpoint type

 - [x] Websocket
 - [ ] HTTP (legacy)

## RPC

 - [x] sendTransaction
 - [ ] sendTransactionAsync
 - [ ] sendRawTransaction
 - [ ] sendRawPrivateTransaction

## Status report

 - [x] CLI
 - [ ] JSON
 - [ ] CSV
