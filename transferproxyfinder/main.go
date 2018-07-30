package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"

	"bytes"
	"encoding/hex"

	"io"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const rpcserver = "http://127.0.0.1:7731"

/*
查找所有的 transfer proxy
*/

//所有相关地址
var contractAddress = map[common.Address]bool{
	common.HexToAddress("0x72E70ad1FD8Bb22318d31ED9e9FB85DC56BAD19c"): true,
	common.HexToAddress("0xb80463F354b54b958e88b4D5385693bB554cbd8E"): true,
	common.HexToAddress("0x7f2a6FB65bcb31c56872D6E9Bf48016292d7E198"): true,
	common.HexToAddress("0x2E5fb91975C3fB6F6D61C5859314E7c53AF07912"): true,
	common.HexToAddress("0xAf038b0290C35A3210D5Cfda9f39ffcEEa1c1Ea5"): true,
	common.HexToAddress("0x61c1a8A9524044f4d2FA3cB51eAa166e2aA59F3A"): true,
	common.HexToAddress("0xa2B6d9e310893C24FA61129e92070b9b40d02016"): true,
	common.HexToAddress("0x4730fB1463A6F1F44AEB45F6c5c422427f37F4D0"): true,
	common.HexToAddress("0x72efF9e6d24e9078A87f51e569aD5560bFB3Fd40"): true,
	common.HexToAddress("0xeB4844a87937C9ee3ae6C89268806C0AB5bd5DfF"): true,
	common.HexToAddress("0x67119bbdc1683df2ad85CCEa4304619F5D62e000"): true,
	common.HexToAddress("0x1Bf63Aca0124c9617d99C13Ec3C279FF3e76F467"): true,
	common.HexToAddress("0xdeDd02DD8f7Af7418d587C3c2ae67E127a27eC3c"): true,
	common.HexToAddress("0x8309Dd14df991B8f26954B3a2C0d03AAD6cC0547"): true,
	common.HexToAddress("0x454e3a25E70a5aEdd0620337671fE6Df8826636F"): true,
	common.HexToAddress("0xd51C109110E7ccB7f6B84DdD2485c2F8E2753415"): true,
	common.HexToAddress("0x84c17fC312cb0e9b00330675e643B340e3303fe6"): true,
	common.HexToAddress("0xF20b76Ed9d5467fDcDc1444455e303257d2827c7"): true,
	common.HexToAddress("0xE39C277f81d347Af8aA8c8B8B20061b7Dc78bc64"): true,
	common.HexToAddress("0x818CA1fa9F88F76D061cddB4169C287E7DbcC6D8"): true,
	common.HexToAddress("0x41fe87D16d60d069d533e2Dc67462cfe4eEeAdEE"): true,
	common.HexToAddress("0x360E51857242661dE8F3Ec4E6C684B45b3c0DE87"): true,
	common.HexToAddress("0x98D178Fffb226CC6d53157663573d7B3D66c1446"): true,
	common.HexToAddress("0xAf30D2a7E90d7DC361c8C4585e9BB7D2F6f15bc7"): true,
	common.HexToAddress("0x43eE79e379e7b78D871100ed696e803E7893b644"): true,
	common.HexToAddress("0x03AF37073258B08FfFF303e9E07E8a0B7bfc4fd9"): true,
	common.HexToAddress("0x9a4FE70Bb7b39127f4772acaF0d000578644b39d"): true,
	common.HexToAddress("0x44033F223F8d1c1BC40a0bCCa30C37f9737425a3"): true,
	common.HexToAddress("0xa3A07912E74AcF57F3362d8b4CdA85ce652fC887"): true,
	common.HexToAddress("0xd7Dc42B78b5CA37Ff5493598D5B6978DC98c3b38"): true,
	common.HexToAddress("0x9E88770DA20ebea0Df87aD874c2F5cf8ab92f605"): true,
	common.HexToAddress("0xf7c3B0ea78e1a65DFf35E7F20701Ac0709c9B86F"): true,
	common.HexToAddress("0x405C5e9804206CFe4F982310c48CDA4E3B8471A3"): true,
	common.HexToAddress("0x5Ba49Fcf7C9Dce67d6a9FC92Cc51c8E77CbA5b7a"): true,
	common.HexToAddress("0xE8599783e272ec087FaDA183b05516304Fa2c7Ff"): true,
	common.HexToAddress("0x55F93985431Fc9304077687a35A1BA103dC1e081"): true,
	common.HexToAddress("0x4c251de85Ad3498c5b89388d8efc696ddD0b0fE7"): true,
	common.HexToAddress("0xd780Ae2Bf04cD96E577D3D014762f831d97129d0"): true,
	common.HexToAddress("0x3AC6cb00f5a44712022a51fbace4C7497F56eE31"): true,
	common.HexToAddress("0xe780C49fe4B9022a0781B2DFcD34BBB337D946E7"): true,
	common.HexToAddress("0x8feBf7551EeA6Ce499F96537Ae0e2075c5A7301a"): true,
	common.HexToAddress("0xE3050dAEcA9ef42E2549BA8D9cfb89d9080846d5"): true,
	common.HexToAddress("0xD73bE539d6B2076BaB83CA6Ba62DfE189aBC6Bbe"): true,
	common.HexToAddress("0xB363A3C584b1f379c79fBF09df015DA5529d4dac"): true,
	common.HexToAddress("0xdfdc0D82d96F8fd40ca0CFB4A288955bECEc2088"): true,
	common.HexToAddress("0x825c1d61a5b170808a31b895c26ee4f24B690d59"): true,
	common.HexToAddress("0xcF9fBFfEC9e0e5BbC62E79bf1965f5Db76955661"): true,
	common.HexToAddress("0xDdCAf6c604592b37f775F3d303A02013BBD7AD93"): true,
	common.HexToAddress("0x4C54a332e3c962D5E513C3cc927c4cCf3452C79C"): true,
	common.HexToAddress("0xC8110527a0b803c0ba15456861d720864915b6ED"): true,
	common.HexToAddress("0x01F2AcF2914860331C1Cb1a9AcecDa7475e06Af8"): true,
	common.HexToAddress("0x286292C0BC3fa5af45E7ad9F0864CcD79F8346ef"): true,
}

var transferProxyID []byte
var approveProxyID []byte

func init() {
	var err error
	transferProxyID, err = hex.DecodeString("eb502d45000000000000000000000000")
	if err != nil {
		log.Fatal("DecodeString err %s", err)
	}
	approveProxyID, err = hex.DecodeString("7f5dfd16000000000000000000000000")
	if err != nil {
		log.Fatal("DecodeString err %s", err)
	}
}
func main() {
	var stopBlockNumber int64 = 5600000
	var i int64
	number := make(chan int64, 10)
	wg := &sync.WaitGroup{}
	//十个线程同时进行
	threadNumber := 10
	wg.Add(threadNumber)
	for j := 0; j < threadNumber; j++ {
		go f(number, wg)
	}
	for i = 1; i < stopBlockNumber; i++ {
		number <- i
	}
	close(number)
	wg.Wait()
	log.Printf("complete ...")
}
func ReadBigInt(reader io.Reader) *big.Int {
	bi := new(big.Int)
	tmpbuf := make([]byte, 32)
	_, err := reader.Read(tmpbuf)
	if err != nil {
		log.Printf("read BigInt error %s\n", err)
	}
	bi.SetBytes(tmpbuf)
	return bi
}

func f(number chan int64, wg *sync.WaitGroup) {
	ctx := context.Background()
	fmt.Printf("eth-rpc-endpoint:%s\n", rpcserver)
	// Create an IPC based RPC connection to a remote node and an authorized transactor
	conn, err := ethclient.Dial(rpcserver)
	if err != nil {
		log.Fatalf(fmt.Sprintf("Failed to connect to the Ethereum client: %v", err))
	}
	v := big.NewInt(0)
	for {
		i, ok := <-number
		if !ok {
			break //channel closed
		}
		b, err := conn.BlockByNumber(ctx, big.NewInt(i))
		if err != nil {
			log.Printf("get block err %s for %d\n", err, i)
			continue
		}
		if b.Number().Int64()%1000 == 0 {
			log.Printf("process block:%s\n", b.Number())
		}

		txs := b.Transactions()
		for index, tx := range txs {
			d := tx.Data()
			//log.Printf("data=%s", hex.EncodeToString(d))
			if bytes.Index(d, transferProxyID) == 0 {
				if contractAddress[*tx.To()] {
					valueb := d[4+32*2 : 4+32*3]
					value := ReadBigInt(bytes.NewReader(valueb))
					log.Printf("transferproxytxhash=%s,blocknumber=%s,index=%d,value=%s,token=%s", tx.Hash().String(), b.Number(), index, value, tx.To().String())
					v = v.Add(v, value)
				}
			}
			if bytes.Index(d, approveProxyID) == 0 {
				if contractAddress[*tx.To()] {
					valueb := d[4+32*2 : 4+32*3]
					value := ReadBigInt(bytes.NewReader(valueb))
					log.Printf("approveproxytxhash=%s,blocknumber=%s,index=%d,value=%s,token=%s", tx.Hash().String(), b.Number(), index, value, tx.To().String())
					v = v.Add(v, value)
				}
			}
		}

	}
	wg.Done()
	log.Printf("thread total=%s\n", v)
}
