package main

import (
	. "app/fabric"
	"log"
	"fmt"
	_ "app/coin"
	coin "app/coin"
	"strings"
	"strconv"
	"sync"
)

func main() {
	InitCA("config.yaml")
	myca := new(CA)
	err := myca.InitCaServer("peerorg1", "enroll_user_peerorg1")
	if err != nil {
		log.Fatalf("Init CA FAILT: ",err.Error())
	} else {
		fmt.Println("Init CA SUCCESS")
	}

	userNameA := GenerateRandomID()
	_,_,err = myca.RegisterAndEnrollUser(userNameA,"userAW", "org1.department1")
	if err != nil {
		fmt.Println("RegisterAndEnrollUser FAILT",err)
		return
	} else {
		fmt.Println("RegisterAndEnrollUser User",userNameA)
	}

	userNameB := GenerateRandomID()
	_,_,err = myca.RegisterAndEnrollUser(userNameB,"userBW", "org1.department1")
	if err != nil {
		fmt.Println("RegisterAndEnrollUser FAILT",err)
		return
	} else {
		fmt.Println("RegisterAndEnrollUser User",userNameB)
	}

	userNameC := GenerateRandomID()
	_,_,err = myca.RegisterAndEnrollUser(userNameC,"userBW", "org1.department1")
	if err != nil {
		fmt.Println("RegisterAndEnrollUser FAILT",err)
		return
	} else {
		fmt.Println("RegisterAndEnrollUser User",userNameC)
	}



	server := new(FabricServer)
	err = server.Init("peerorg1", "enroll_user_peerorg1")
	if err != nil {
		log.Fatalf("Init Fabric Service failed %v", err)
	}

	chaincodeName := "mychaincodev5"

	//install chaincode

	/*
	err = server.InitAsset("mychannel", "github.com/chaincodetest", chaincodeName,  "v1", "", "")
	if err != nil {
		fmt.Println("InitAsset CC  FAILT",err)
		return
	} else {
		fmt.Println("InitAsset CC SUCCESS")
	}
	*/

	//create account A
	fmt.Println("____________________________________________________create account A______________________________________________________________________")
	txID, err := server.InvokeRegister("mychannel", chaincodeName, userNameA)
	if err != nil {
		fmt.Println("InvokeInit  FAILT",err)
	}
	fmt.Println("Create Account A InvokeInit:",userNameA, txID)
	return

	//create account B
	fmt.Println("____________________________________________________create account B______________________________________________________________________")
	txID, err = server.InvokeRegister("mychannel", chaincodeName, userNameB)
	if err != nil {
		fmt.Println("InvokeInit  FAILT",err)
	}
	fmt.Println("Create Account B InvokeInit:",userNameB, txID)

	//create account C
	fmt.Println("____________________________________________________create account C______________________________________________________________________")
	txID, err = server.InvokeRegister("mychannel", chaincodeName, userNameC)
	if err != nil {
		fmt.Println("InvokeInit  FAILT",err)
	}
	fmt.Println("Create Account C InvokeInit:",userNameC, txID)


	//baseCoin user A
	fmt.Println("____________________________________________________baseCoin user A______________________________________________________________________")
	tx := coin.NewTransaction("")
	TxoutA := coin.NewTxOut(100, userNameA, -1)
	tx.Txout = append(tx.Txout, TxoutA)

	txString := coin.TxToString(tx)
	//txID, err = server.InvokeCoinbase("mychannel", chaincodeName, "admin", txString)
	txID, err = server.InvokeCoinbase("mychannel", chaincodeName, userNameA, txString)
	if err != nil {
		fmt.Println("InvokeInit  FAILT",err)
	}
	fmt.Println("Init Account A InvokeCoinbase:",userNameA, txID)


	//baseCoin user B
	fmt.Println("____________________________________________________baseCoin user B______________________________________________________________________")
	tx = coin.NewTransaction("")
	TxoutB := coin.NewTxOut(100, userNameB, -1)
	tx.Txout = append(tx.Txout, TxoutB)

	txString = coin.TxToString(tx)
	txID, err = server.InvokeCoinbase("mychannel", chaincodeName, userNameB, txString)
	if err != nil {
		fmt.Println("InvokeInit  FAILT",err)
	}
	fmt.Println("Init Account B InvokeCoinbase:",userNameB, txID)


	//baseCoin user C
	fmt.Println("____________________________________________________baseCoin user C______________________________________________________________________")
	tx = coin.NewTransaction("")
	TxoutC := coin.NewTxOut(100, userNameC, -1)
	tx.Txout = append(tx.Txout, TxoutC)

	txString = coin.TxToString(tx)
	txID, err = server.InvokeCoinbase("mychannel", chaincodeName, userNameC, txString)
	if err != nil {
		fmt.Println("InvokeInit  FAILT",err)
	}
	fmt.Println("Init Account C InvokeCoinbase:",userNameC, txID)



	//query user A
	fmt.Println("____________________________________________________query user A______________________________________________________________________")
	var txHashA []string
	addrResultsA, err := server.QueryAddrs("mychannel", chaincodeName, userNameA)
	if err != nil {
		fmt.Println("QueryAddrs  FAILT",err)
	}

	fmt.Println("QueryAddrs Account:", addrResultsA)
	for name, account := range addrResultsA.Accounts {
		fmt.Println("user:", name)
		fmt.Println("Account.Addr:",account.Addr)
		fmt.Println("Account.Balance:",account.Balance)

		for key, TxOut := range account.Txouts {
			fmt.Println("key=",key)
			fmt.Println("TxOut.Value:", TxOut.Value)
			fmt.Println("TxOut.Addr:",TxOut.Addr)
			fmt.Println("TxOut.Until:",TxOut.Until)

			txHashA = strings.Split(key,":")
			fmt.Println("txHash:",txHashA)
		}
	}

	txDate, err := server.QueryTx("mychannel", chaincodeName, txHashA[0])
	if err != nil {
		fmt.Println("QueryTx  FAILT",err)
	}
	fmt.Println("Query TX:",txDate)


	//query user B
	fmt.Println("____________________________________________________query user B______________________________________________________________________")
	var txHashB []string
	addrResultsB, err := server.QueryAddrs("mychannel", chaincodeName, userNameB)
	if err != nil {
		fmt.Println("QueryAddrs  FAILT",err)
	}

	fmt.Println("QueryAddrs Account:", addrResultsB)
	for name, account := range addrResultsB.Accounts {
		fmt.Println("user:", name)
		fmt.Println("Account.Addr:",account.Addr)
		fmt.Println("Account.Balance:",account.Balance)

		for key, TxOut := range account.Txouts {
			fmt.Println("key=",key)
			fmt.Println("TxOut.Value:", TxOut.Value)
			fmt.Println("TxOut.Addr:",TxOut.Addr)
			fmt.Println("TxOut.Until:",TxOut.Until)

			txHashB = strings.Split(key,":")
			fmt.Println("txHash:",txHashB)
		}
	}

	txDate, err = server.QueryTx("mychannel", chaincodeName, txHashB[0])
	if err != nil {
		fmt.Println("QueryTx  FAILT",err)
	}
	fmt.Println("Query TX:",txDate)


	fmt.Println("____________________________________________________QueryCoin______________________________________________________________________")
	coinDate, err :=server.QueryCoin("mychannel", chaincodeName)
	if err != nil {
		fmt.Println("QueryCoin  FAILT",err)
	}
	fmt.Println("Query Coin:",coinDate)


	var wg sync.WaitGroup

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		//A transfer coin
		fmt.Println("____________________________________________________A transfer C coin______________________________________________________________________")
		Ix, err := strconv.ParseInt(txHashA[1], 10, 64)
		if err != nil {
			panic(err)
		}
		tx = coin.NewTransaction(userNameA)
		TxIn := coin.NewTxIn(userNameA, txHashA[0], uint32(Ix))
		TxoutC = coin.NewTxOut(100, userNameC, 0)
		tx.Txin = append(tx.Txin, TxIn)
		tx.Txout = append(tx.Txout, TxoutC)

		txString = coin.TxToString(tx)
		txID, err = server.InvokeTransfer("mychannel", chaincodeName, userNameA, txString)
		if err != nil {
			fmt.Println("InvokeTransfer  FAILT", err)
		}
		fmt.Println("Transfer txID:", txID)

		fmt.Println("1 done")
		wg.Done()
	}(&wg)

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		//B transfer coin
		fmt.Println("____________________________________________________B transfer C coin______________________________________________________________________")
		Ix, err := strconv.ParseInt(txHashB[1], 10, 64)
		if err != nil {
			panic(err)
		}
		tx = coin.NewTransaction(userNameB)
		TxIn := coin.NewTxIn(userNameB, txHashB[0], uint32(Ix))
		TxoutC = coin.NewTxOut(100, userNameC, 0)
		tx.Txin = append(tx.Txin, TxIn)
		tx.Txout = append(tx.Txout, TxoutC)

		txString = coin.TxToString(tx)
		txID, err = server.InvokeTransfer("mychannel", chaincodeName, userNameB, txString)
		if err != nil {
			fmt.Println("InvokeTransfer  FAILT", err)
		}
		fmt.Println("Transfer txID:", txID)
		fmt.Println("1 done")
		wg.Done()
	}(&wg)

	wg.Wait()
}






