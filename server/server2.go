package main

import (
	. "app/fabric"
	"fmt"
	_ "app/coin"
)

import (
	"sync"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"

	"github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric/protos/peer"
	"os"
)

const (
	eventTimeout = time.Second * 5
)

func main() {
	testSetup := initializeTests()

	//testReconnectEventHub(testSetup)


	testFailedTx(testSetup) //chaincode中的event有助于加快交易函数的反馈 返回

	return
	//testFailedTx2(testSetup) //提议失败则反馈

//	testFailedTxErrorCode(testSetup) //读冲突


	testMultipleBlockEventCallbacks(testSetup)
}


func initializeTests() BaseSetupImpl {
	testSetup := BaseSetupImpl{
		ConfigFile:      "config.yaml",
		ChannelID:       "mychannel",
		OrgID:           "peerorg1",
		ChannelConfig:   "channel.tx",
		ConnectEventHub: true,
	}


	if err := testSetup.Initialize("enroll_user"); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	testSetup.ChainCodeID = "eventv5"

	/*
	// Install and Instantiate Events CC
	if err := testSetup.InstallCC(testSetup.ChainCodeID, "github.com/events_cc", "v0", nil); err != nil {
		fmt.Printf("installCC return error: %v", err)
	}

	if err := testSetup.InstantiateCC(testSetup.ChainCodeID, "github.com/events_cc", "v0", nil); err != nil {
		fmt.Printf("instantiateCC return error: %v", err)
	}
	*/


	return testSetup
}

func testReconnectEventHub(testSetup BaseSetupImpl) {
	// Test disconnect event hub
	testSetup.EventHub.Disconnect()
	if testSetup.EventHub.IsConnected() {
		fmt.Printf("Failed to disconnect event hub")
	}
	// Reconnect event hub
	if err := testSetup.EventHub.Connect(); err != nil {
		fmt.Printf("Failed to connect event hub")
	}
}

func monitorFailedTx(testSetup BaseSetupImpl, done1 chan bool, fail1 chan error, done2 chan bool, fail2 chan error) {
	rcvDone := false
	rcvFail := false
	timeout := time.After(eventTimeout)

Loop:
	for !rcvDone || !rcvFail {
		select {
		case <-done1:
			rcvDone = true
		case <-fail1:
			fmt.Printf("Received fail for first invoke")
			//os.Exit(0)
		case <-done2:
			fmt.Printf("Received success for second invoke")
			//os.Exit(0)
		case <-fail2:
			rcvFail = true
		case <-timeout:
			fmt.Printf("Timeout: Didn't receive events")
			break Loop
		}
	}

	if !rcvDone || !rcvFail {
		fmt.Printf("Didn't receive events (done: %t; fail %t)", rcvDone, rcvFail)
	}
}

func testFailedTx(testSetup BaseSetupImpl) {
	fcn := "invoke"

	// Arguments for events CC
	var args []string
	args = append(args, "invoke")
	args = append(args, "SEVERE")

	tpResponses1, tx1, err := testSetup.CreateAndSendTransactionProposal(testSetup.Channel, testSetup.ChainCodeID, fcn, args, []apitxn.ProposalProcessor{testSetup.Channel.PrimaryPeer()}, nil)
	if err != nil {
		fmt.Printf("CreateAndSendTransactionProposal return error: %v", err)
	}

	tpResponses2, tx2, err := testSetup.CreateAndSendTransactionProposal(testSetup.Channel, testSetup.ChainCodeID, fcn, args, []apitxn.ProposalProcessor{testSetup.Channel.PrimaryPeer()}, nil)
	if err != nil {
		fmt.Printf("CreateAndSendTransactionProposal return error: %v", err)
	}

	// Register tx1 and tx2 for commit/block event(s)
	done1, fail1 := testSetup.RegisterTxEvent(tx1, testSetup.EventHub)
	defer testSetup.EventHub.UnregisterTxEvent(tx1)

	done2, fail2 := testSetup.RegisterTxEvent(tx2, testSetup.EventHub)
	defer testSetup.EventHub.UnregisterTxEvent(tx2)

	// Setup monitoring of events


	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		monitorFailedTx(testSetup, done1, fail1, done2, fail2)
	}()


	// Test invalid transaction: create 2 invoke requests in quick succession that modify
	// the same state variable which should cause one invoke to be invalid

	//不执行下面的代码，则event 无法受到txid
	_, err = testSetup.CreateAndSendTransaction(testSetup.Channel, tpResponses1)
	if err != nil {
		fmt.Printf("First invoke failed err: %v\n", err)
	}else {
		fmt.Printf("CreateAndSendTransaction invoke SUCCESS: %v\n", tpResponses1)
	}
	_, err = testSetup.CreateAndSendTransaction(testSetup.Channel, tpResponses2)
	if err != nil {
		fmt.Printf("Second invoke failed err: %v\n", err)
	}else {
		fmt.Printf("CreateAndSendTransaction invoke SUCCESS: %v\n", tpResponses2)
	}


	wg.Wait()
}


func testFailedTx2(testSetup BaseSetupImpl) {
	var wg sync.WaitGroup
	wg.Add(2)

	for i:=0; i < 2; i++ {
		go func(num int, testSetup BaseSetupImpl) {
			defer wg.Done()
			fcn := "invoke"

			// Arguments for events CC
			var args []string
			args = append(args, "invoke")
			args = append(args, "SEVERE")

			tpResponses1, tx, err := testSetup.CreateAndSendTransactionProposal(testSetup.Channel, testSetup.ChainCodeID, fcn, args, []apitxn.ProposalProcessor{testSetup.Channel.PrimaryPeer()}, nil)
			if err != nil {
				fmt.Printf("CreateAndSendTransactionProposal return error: %v", err)
			}
			// Register tx1 and tx2 for commit/block event(s)
			done, fail := testSetup.RegisterTxEvent(tx, testSetup.EventHub)
			defer testSetup.EventHub.UnregisterTxEvent(tx)

			// Test invalid transaction: create 2 invoke requests in quick succession that modify
			// the same state variable which should cause one invoke to be invalid
			_, err = testSetup.CreateAndSendTransaction(testSetup.Channel, tpResponses1)
			if err != nil {
				fmt.Printf("First invoke failed err: %v", err)
			}

			select {
			case <-done:
				fmt.Println(num,"invoke success")
			case err := <-fail:
				fmt.Println(num, "invoke Error received from eventhub for txid(%s) Nonce(%s), error(%s)", tx.ID, string(tx.Nonce), err.Error())
			case <-time.After(time.Second * 30):
				fmt.Println(num, "invoke Didn't receive block event for txid(%s)", tx)
			}
		}(i, testSetup)
	}
	wg.Wait()
}




func monitorFailedTxErrorCode(testSetup BaseSetupImpl, done chan bool, fail chan pb.TxValidationCode, done2 chan bool, fail2 chan pb.TxValidationCode) {
	rcvDone := false
	rcvFail := false
	timeout := time.After(eventTimeout)

Loop:
	for !rcvDone || !rcvFail {
		select {
		case <-done:
			rcvDone = true
			fmt.Printf("success invoke 1 \n")
		case <-fail:
			fmt.Printf("Received fail for first invoke")
		case <-done2:
			fmt.Printf("Received success for second invoke 2 \n")
		case errorValidationCode := <-fail2:
			if errorValidationCode.String() != "MVCC_READ_CONFLICT" {
				fmt.Printf("Expected error code MVCC_READ_CONFLICT. Got %s\n", errorValidationCode.String())
			} else {
				fmt.Printf("Expected  code MVCC_READ_CONFLICT. Got %s\n", errorValidationCode.String())
			}
			rcvFail = true
		case <-timeout:
			fmt.Printf("Timeout: Didn't receive events")
			break Loop
		}
	}

	if !rcvDone || !rcvFail {
		fmt.Printf("Didn't receive events (done: %t; fail %t)", rcvDone, rcvFail)
	}
}
func testFailedTxErrorCode( testSetup BaseSetupImpl) {
	fcn := "invoke"

	// Arguments for events CC
	var args []string
	args = append(args, "invoke")
	args = append(args, "SEVERE")

	tpResponses1, tx1, err := testSetup.CreateAndSendTransactionProposal(testSetup.Channel, testSetup.ChainCodeID, fcn, args, []apitxn.ProposalProcessor{testSetup.Channel.PrimaryPeer()}, nil)

	if err != nil {
		fmt.Printf("CreateAndSendTransactionProposal return error: %v", err)
	}

	tpResponses2, tx2, err := testSetup.CreateAndSendTransactionProposal(testSetup.Channel, testSetup.ChainCodeID, fcn, args, []apitxn.ProposalProcessor{testSetup.Channel.PrimaryPeer()}, nil)
	if err != nil {
		fmt.Printf("CreateAndSendTransactionProposal return error: %v", err)
	}

	done := make(chan bool)
	fail := make(chan pb.TxValidationCode)

	testSetup.EventHub.RegisterTxEvent(tx1, func(txId string, errorCode pb.TxValidationCode, err error) {
		if err != nil {
			fail <- errorCode
		} else {
			done <- true
		}
	})

	defer testSetup.EventHub.UnregisterTxEvent(tx1)

	done2 := make(chan bool)
	fail2 := make(chan pb.TxValidationCode)

	testSetup.EventHub.RegisterTxEvent(tx2, func(txId string, errorCode pb.TxValidationCode, err error) {
		if err != nil {
			fail2 <- errorCode
		} else {
			done2 <- true
		}
	})

	defer testSetup.EventHub.UnregisterTxEvent(tx2)

	// Setup monitoring of events
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		monitorFailedTxErrorCode(testSetup, done, fail, done2, fail2)
	}()

	// Test invalid transaction: create 2 invoke requests in quick succession that modify
	// the same state variable which should cause one invoke to be invalid
	_, err = testSetup.CreateAndSendTransaction(testSetup.Channel, tpResponses1)
	if err != nil {
		fmt.Printf("First invoke failed err: %v", err)
	}
	_, err = testSetup.CreateAndSendTransaction(testSetup.Channel, tpResponses2)
	if err != nil {
		fmt.Printf("Second invoke failed err: %v", err)
	}

	wg.Wait()
}


func testMultipleBlockEventCallbacks(testSetup BaseSetupImpl) {
	fcn := "invoke"

	// Arguments for events CC
	var args []string
	args = append(args, "invoke")
	args = append(args, "SEVERE")

	// Create and register test callback that will be invoked upon block event
	test := make(chan bool)
	testSetup.EventHub.RegisterBlockEvent(func(block *common.Block) {  //监听block消息
		fmt.Printf("Received test callback on block event")
		test <- true
	})

	tpResponses, tx, err := testSetup.CreateAndSendTransactionProposal(testSetup.Channel, testSetup.ChainCodeID, fcn, args, []apitxn.ProposalProcessor{testSetup.Channel.PrimaryPeer()}, nil)
	if err != nil {
		fmt.Printf("CreateAndSendTransactionProposal returned error: %v", err)
	}
	//_ = tpResponses

	// Register tx for commit/block event(s)
	done, fail := testSetup.RegisterTxEvent(tx, testSetup.EventHub)
	defer testSetup.EventHub.UnregisterTxEvent(tx)

	// Setup monitoring of events
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		monitorMultipleBlockEventCallbacks(testSetup, done, fail, test)
	}()

	//不执行下面的代码，则服务envet受到orderer确认的txid
	_, err = testSetup.CreateAndSendTransaction(testSetup.Channel, tpResponses)
	if err != nil {
		fmt.Printf("CreateAndSendTransaction failed with error: %v", err)
	}

	wg.Wait()
}

func monitorMultipleBlockEventCallbacks(testSetup BaseSetupImpl, done chan bool, fail chan error, test chan bool) {
	rcvTxDone := false
	rcvTxEvent := false
	timeout := time.After(eventTimeout)

Loop:
	for !rcvTxDone || !rcvTxEvent {
		select {
		case <-done:
			rcvTxDone = true
			fmt.Printf("rcvTxDone = true \n")
		case <-fail:
			fmt.Printf("Received tx failure")
		case <-test:
			rcvTxEvent = true
			fmt.Printf("rcvTxEvent = true \n")
		case <-timeout:
			fmt.Printf("Timeout while waiting for events")
			break Loop
		}
	}

	if !rcvTxDone || !rcvTxEvent {
		fmt.Printf("Didn't receive events (tx event: %t; tx done %t)", rcvTxEvent, rcvTxDone)
	}
}



/*

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

	chaincodeName := "mychaincodeeventv0"

	//install chaincode

	err = server.InitAsset("mychannel", "github.com/events_cc", chaincodeName,  "v1", "", "")
	if err != nil {
		fmt.Println("InitAsset CC  FAILT",err)
		return
	} else {
		fmt.Println("InitAsset CC SUCCESS")
	}


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

*/





