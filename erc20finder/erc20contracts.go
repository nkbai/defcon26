package main

import (
	"github.com/astaxie/beego/orm"

	"sync"

	"fmt"

	"io/ioutil"
	"strings"

	"log"

	"errors"
	"regexp"
	"strconv"

	"context"
	"encoding/hex"

	"github.com/PuerkitoBio/goquery"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/net/html"
)

type Contract struct {
	Id           int    // 主键
	Address      string `orm:"unique"`
	Name         string
	TotalAmount  string
	BlockNumer   int64
	TxHash       string
	Code         string `orm:"type(text)"`
	BinCode      string `orm:"type(text)"`
	HasEcrecover int    //0 没有处理,1没有,2 源码有,3 反汇编疑似有
	CodeStatus   int    //0 没获取过,1获取到了,2,获取了不成功.
	TxNumber     int64  //发生在这个合约上的交易,可以从 getsourcecode 获取.
}

var o orm.Ormer
var lock sync.Mutex

func init() {
	orm.RegisterModel(new(Contract))
	orm.RegisterDriver("sqlite", orm.DRSqlite)
	orm.RegisterDataBase("default", "sqlite3", "contracts.db")
	orm.RunSyncdb("default", false, true)

	o = orm.NewOrm()
	o.Using("default")
}
func AddContract(Name string, Address common.Address, total string) {
	lock.Lock()
	defer lock.Unlock()
	c := &Contract{
		Name:        Name,
		Address:     Address.String(),
		TotalAmount: total,
	}
	_, err := o.Insert(c)
	if err != nil {
		fmt.Printf("insert error :%s\n", err)
	}
}
func collectContracts() {
	data, _ := ioutil.ReadFile("erc20_2.txt")
	lines := strings.Split(string(data), "\n")
	for _, l := range lines {
		ai := strings.Index(l, "addr=")
		if ai != 0 {
			log.Printf("skip %s", l)
			continue
		}
		ll := strings.Split(l, ",")
		if len(ll) < 3 {
			log.Printf("skip %s", l)
			continue
		}
		addr := common.HexToAddress(strings.Split(ll[0], "=")[1])
		l = l[len(ll[0])+1:]
		ai = strings.Index(l, "name=")
		if ai != 0 {
			log.Printf("skip %s", l)
			continue
		}
		ti := strings.Index(l, "total=")
		name := l[5:ti]
		total := l[ti+6:]
		log.Printf("addr=%s,name=%s,t=%s", addr.String(), name, total)
		AddContract(name, addr, total)
	}
	o.Commit()
}

var ErrNotFound = errors.New("not found")

func GetSourceCode(tokenAddr common.Address) (int, string, error) {
	doc, err := goquery.NewDocument(fmt.Sprintf("https://etherscan.io/address/%s", tokenAddr.String()))
	if err != nil {
		return 0, "", err
	}
	txNumber := 0
	str, _ := doc.Html()
	reg := regexp.MustCompile(`Latest (\d+) txn`)
	q := reg.FindStringSubmatch(str)
	if len(q) != 2 {
		err := fmt.Errorf("cannot get txs for %s", tokenAddr.String())
		return txNumber, "", err
	}
	txNumber, _ = strconv.Atoi(q[1])
	if txNumber < 1 { //at least one transaction ,for creation
		err := fmt.Errorf("tx number err %q", q)
		return txNumber, "", err
	}
	txNumber -= 1
	var node *html.Node
	sel := doc.Find("#editor")
	if len(sel.Nodes) > 0 {
		node = sel.Get(0)
	} else {
		return txNumber, "", ErrNotFound
	}
	s := node.FirstChild.Data
	s = html.UnescapeString(s)
	return txNumber, s, nil
}

func collectSourceCode() {
	var contracts []*Contract
	qs := o.QueryTable(&Contract{})
	_, err := qs.Limit(100000).All(&contracts)
	if err != nil {
		log.Printf("all err %s", err)
		return
	}
	for i, c := range contracts {
		if c.CodeStatus != 0 {
			continue
		}
		log.Printf("%d process %s %s", i, c.Name, c.Address)
		tx, s, err := GetSourceCode(common.HexToAddress(c.Address))
		//log.Printf("%s tx=%d,code=%d", c.Address, tx, len(s))
		if err == nil {
			c.CodeStatus = 1
			c.Code = s
			c.TxNumber = int64(tx)
			o.Update(c, "Code", "CodeStatus", "TxNumber")
		} else /*if err == funcs.ErrNotFound*/ {
			c.CodeStatus = 2
			c.TxNumber = int64(tx)
			o.Update(c, "CodeStatus", "TxNumber")
		}
	}
}
func collectCode() {
	const rpcserver = "http://10.0.0.21:8545"
	ctx := context.Background()
	conn, err := ethclient.Dial(rpcserver)
	if err != nil {
		log.Fatalf(fmt.Sprintf("Failed to connect to the Ethereum client: %v", err))
	}
	var contracts []*Contract
	qs := o.QueryTable(&Contract{})
	_, err = qs.Limit(100000).All(&contracts)
	if err != nil {
		log.Printf("all err %s", err)
		return
	}
	for _, c := range contracts {
		raw, err := conn.CodeAt(ctx, common.HexToAddress(c.Address), nil)
		if err == nil && len(raw) > 0 {
			c.BinCode = hex.EncodeToString(raw)
			o.Update(c, "BinCode")
			log.Printf("%s update  code", c.Address)
		}
	}
}
func main() {
	collectSourceCode()
}
