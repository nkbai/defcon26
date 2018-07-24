package main

import (
	"context"
	"errors"
	"path/filepath"

	"os"

	"fmt"

	"math/big"

	"log"

	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/node"
	"github.com/nkbai/defcon26/accounts"
	"github.com/nkbai/defcon26/proxytoken"
	"github.com/nkbai/goice/utils"
	"github.com/slonzok/getpass"
	"github.com/urfave/cli"
)

//transfer from `from` to `to` `amount` tokens by proxy
//ugt: 0x3772cedc564cf1e9ae23d761ff8a79cf764ea58b
//mtc:0x0913358c40bf86af448bb3af5a20d9da1b7475da
func main() {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "to",
			Usage: "transfer to this address",
		},
		cli.StringFlag{
			Name:  "keystore-path",
			Usage: "If you have a non-standard path for the ethereum keystore directory provide it using this argument. ",
			Value: filepath.Join(node.DefaultDataDir(), "keystore"),
		},
		cli.StringFlag{
			Name:  "proxy",
			Usage: "use this proxy's service to transfer,this account must have enough ether for transaction gas",
		},
		cli.StringFlag{
			Name:  "from",
			Usage: "transfer token from this address",
		},
		cli.StringFlag{
			Name:  "token",
			Usage: "which token to transfer by proxy",
		},
		cli.Int64Flag{
			Name:  "amount",
			Usage: "how many tokens to transfer",
		},
		cli.StringFlag{
			Name: "eth-rpc-endpoint",
			Usage: `"host:port" address of ethereum JSON-RPC server.\n'
	           'Also accepts a protocol prefix (ws:// or ipc channel) with optional port',`,
			Value: node.DefaultIPCEndpoint("geth"),
		},
	}
	app.Action = mainCtx
	app.Name = "trproxy"
	app.Version = "0.1"
	err := app.Run(os.Args)
	if err != nil {
		log.Fatalf("run err %s", err)
	}
}
func promptAccount(addr common.Address, keystorePath string, hint string) (key *ecdsa.PrivateKey, err error) {
	am := accounts.NewAccountManager(keystorePath)
	if len(am.Accounts) == 0 {
		err = fmt.Errorf("No Ethereum accounts found in the directory %s", keystorePath)
		return
	}
	for i := 0; i < 3; i++ {
		var keybin []byte
		//retries three times
		password := getpass.Prompt(fmt.Sprintf("Enter the password to unlock %s account", hint))
		keybin, err = am.GetPrivateKey(addr, password)
		if err != nil && i == 3 {
			log.Printf(fmt.Sprintf("Exhausted passphrase unlock attempts for %s. Aborting ...", addr))
			utils.SystemExit(1)
		}
		if err != nil {
			log.Printf(fmt.Sprintf("password incorrect\n Please try again or kill the process to quit.\nUsually Ctrl-c."))
			continue
		}
		key, err = crypto.ToECDSA(keybin)
		return
	}
	err = errors.New("must specified password")
	return
}

func transferproxy(conn *ethclient.Client, fromkey, proxyKey *ecdsa.PrivateKey, to, tokenAddr common.Address, amount *big.Int) {
	token, err := ugt.NewUGToken(tokenAddr, conn)
	if err != nil {
		panic(err)
	}
	var from common.Address
	from = crypto.PubkeyToAddress(fromkey.PublicKey) //1000
	//1token=10^18wei
	value := amount.Mul(amount, big.NewInt(1000000000000000000))
	fee := big.NewInt(0)
	nonce, err := token.GetNonce(nil, from)
	if err != nil {
		panic(err)
	}
	valueraw := math.PaddedBigBytes(math.U256(value), 32)
	feeraw := math.PaddedBigBytes(math.U256(fee), 32)
	nonceraw := math.PaddedBigBytes(math.U256(nonce), 32)
	msg := crypto.Keccak256Hash(from[:], to[:], valueraw[:], feeraw[:], nonceraw[:])
	sig, err := crypto.Sign(msg[:], fromkey)
	if err != nil {
		log.Printf("sig err=%s\n", err)
		return
	}

	r := sig[0:32]
	s := sig[32:64]
	v := sig[64]
	var r1, s1 [32]byte
	copy(r1[:], r)
	copy(s1[:], s)
	v += 27
	auth := bind.NewKeyedTransactor(proxyKey)
	tx, err := token.TransferProxy(auth, from, to, value, fee, uint8(v), r1, s1)
	if err != nil {
		log.Fatalf("Failed to Transfer: %v", err)
	}
	ctx := context.Background()
	_, err = bind.WaitMined(ctx, conn, tx)
	if err != nil {
		log.Fatalf("failed to Transfer when mining :%v", err)
	}
	log.Printf("Transfer complete...\ntxhash=%s", tx.Hash().String())
}
func mainCtx(ctx *cli.Context) (err error) {
	log.Printf("welcome to trproxy,version %s\n", ctx.App.Version)
	from := common.HexToAddress(ctx.String("from"))
	to := common.HexToAddress(ctx.String("to"))
	token := common.HexToAddress(ctx.String("token"))
	proxy := common.HexToAddress(ctx.String("proxy"))
	amount := ctx.Int64("amount")
	log.Printf("transfer from %s to %s %d tokens, token is %s, proxy is %s\n",
		from.String(), to.String(), amount, token.String(), proxy.String())
	fromkey, err := promptAccount(from, ctx.String("keystore-path"), "from")
	if err != nil {
		return err
	}
	proxykey, err := promptAccount(proxy, ctx.String("keystore-path"), "proxy")
	conn, err := ethclient.Dial(ctx.String("eth-rpc-endpoint"))
	if err != nil {
		return err
	}
	transferproxy(conn, fromkey, proxykey, to, token, big.NewInt(amount))
	return nil
}
