/*
Copyright Hydrusio Labs Inc. 2016 All Rights Reserved.
Written by mint.zhao.chiu@gmail.com. github.com: https://www.github.com/mintzhao

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package coin

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/op/go-logging"
	"fmt"
)

var (
	logger = logging.MustGetLogger("hydruscoin")
)

// Hydruscoin
type Hydruscoin struct{}

// Init deploy chaincode into vp
func (coin *Hydruscoin) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("hydruscoin")

	// construct a new store
	store := MakeChaincodeStore(stub)

	// deploy hydruscoin chaincode only need to set coin stater
	if err := store.InitCoinInfo(); err != nil {
		return shim.Error(err.Error())
	}

	logger.Debug("deploy Hydruscoin successfully")
	return shim.Success(nil)
}


// Invoke function
const (
	IF_REGISTER string = "invoke_register"
	IF_COINBASE string = "invoke_coinbase"
	IF_TRANSFER string = "invoke_transfer"
)

// Query function
const (
	QF_ADDRS = "query_addrs"
	QF_TX    = "query_tx"
	QF_COIN  = "query_coin"
)


// Invoke
func (coin *Hydruscoin) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	var err error
	// construct a new store
	store := MakeChaincodeStore(stub)

	function, args := stub.GetFunctionAndParameters()
	if function != "invoke" {
		return shim.Error("Unknown function call")
	}
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting at least 2")
	}

	fmt.Println(function,args)

	switch args[0] {
	case IF_REGISTER:
		err = coin.registerAccount(store, args)
		if err != nil {
			return shim.Error(err.Error())
		} else {
			return shim.Success(nil)
		}
	case IF_COINBASE:
		info, err := coin.coinbase(store, args)
		if err != nil {
			return shim.Error(err.Error())
		} else {
			return shim.Success(info)
		}
	case IF_TRANSFER:
		info, err := coin.transfer(store, args)
		if err != nil {
			return shim.Error(err.Error())
		} else {
			return shim.Success(info)
		}
		// Query
	case QF_ADDRS:
		info, err :=  coin.queryAddrs(store, args)
		if err != nil {
			return shim.Error(err.Error())
		} else {
			return shim.Success(info)
		}
	case QF_TX:
		info, err :=  coin.queryTx(store, args)
		if err != nil {
			return shim.Error(err.Error())
		} else {
			return shim.Success(info)
		}
	case QF_COIN:
		info, err :=  coin.queryCoin(store, args)
		if err != nil {
			return shim.Error(err.Error())
		} else {
			return shim.Success(info)
		}
	default:
		return shim.Error(ErrUnsupportedOperation.Error())
	}
	return shim.Error("Invalid invoke function name. Expecting \"invoke\"")
}
