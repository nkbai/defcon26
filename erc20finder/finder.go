package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/nkbai/defcon26/proxytoken"
)

const rpcserver = "http://127.0.0.1:8545"

func main2() {
	ctx := context.Background()
	fmt.Printf("eth-rpc-endpoint:%s\n", rpcserver)
	// Create an IPC based RPC connection to a remote node and an authorized transactor
	conn, err := ethclient.Dial(rpcserver)
	if err != nil {
		log.Fatalf(fmt.Sprintf("Failed to connect to the Ethereum client: %v", err))
	}
	h, err := conn.HeaderByNumber(ctx, nil)
	if err != nil {
		log.Fatalf("failed to header by number :%s ", err)
	}
	wg := &sync.WaitGroup{}
	wg.Add(8)
	go f(4800000, 4900000, wg)
	go f(4900000, 5000000, wg)
	go f(5000000, 5100000, wg)
	go f(5100000, 5200000, wg)
	go f(5200000, 5300000, wg)
	go f(5300000, 5400000, wg)
	go f(5400000, 5500000, wg)
	go f(5500000, h.Number.Int64(), wg)
	wg.Wait()
}

func f(from, to int64, wg2 *sync.WaitGroup) {

	ctx := context.Background()
	fmt.Printf("eth-rpc-endpoint:%s\n", rpcserver)
	// Create an IPC based RPC connection to a remote node and an authorized transactor
	conn, err := ethclient.Dial(rpcserver)
	if err != nil {
		log.Fatalf(fmt.Sprintf("Failed to connect to the Ethereum client: %v", err))
	}
	h, err := conn.HeaderByNumber(ctx, nil)
	if err != nil {
		log.Fatalf("failed to header by number :%s ", err)
	}
	log.Printf("total block=%d\n", h.Number.Int64())
	wg := sync.WaitGroup{}
	var big0 = big.NewInt(0)
	task := func(from, to int64) {
		var i int64
		wg.Add(1)
		defer wg.Done()
		log.Printf("%d-%d start.., number=%d\n", from, to, to-from)
		i = from - 1
		for i <= to {
		nxt:
			i++
			b, err := conn.BlockByNumber(ctx, big.NewInt(i))
			if err != nil {
				log.Printf("get block err %s for %d\n", err, i)
				goto nxt
			}
			if b.Number().Int64()%1000 == 0 {
				log.Printf("%d-%d process block:%s\n", from, to, b.Number())
			}

			txs := b.Transactions()
			for _, tx := range txs {
				if tx.To() == nil {
					//log.Printf(" new contract @%s\n", b.Number())
					r, err := conn.TransactionReceipt(ctx, tx.Hash())
					if err != nil {
						log.Printf("get receipient err %s, for tx:%s\n", err, tx.Hash().String())
						goto nxt
					}
					t, err := ugt.NewUGToken(r.ContractAddress, conn)
					if err != nil {
						log.Printf("new token err %s for address %s\n", err, r.ContractAddress.String())
						goto nxt
					}
					s, err := t.TotalSupply(nil)
					if err != nil || s.Cmp(big0) <= 0 {
						goto nxt
					}
					name, err := t.Name(nil)
					if err != nil || len(name) <= 0 {
						//没有名字,肯定不正规,忽略.
						goto nxt
					}
					log.Printf("find a erc20 contract :addr=%s,name=%s,total=%s,block=%d,txhash=%s\n", r.ContractAddress.String(), name, s, b.Number().Int64(), tx.Hash().String())
				}
			}

		}
		log.Printf("%d-%d complete\n", from, to)
	}

	//total := h.Number.Int64()
	//oneshare := total / 2
	//var i int64
	//for i = 0; i < 2; i++ {
	//	go task(i*oneshare, i*oneshare+oneshare)
	//}
	go task(from, to)
	time.Sleep(time.Second)
	wg.Wait()
	wg2.Done()
}
