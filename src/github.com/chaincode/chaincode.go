package main

import (
	"fmt"
	"strconv"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type Chaincode struct{

}

func  (t *Chaincode) Init(stub shim.ChaincodeStubInterface) pb.Response{
	_, args := stub.GetFunctionAndParameters()
	var appid string
	var value string 
	var err error

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	
	appid = args[0]
	value = args[1]

	err = stub.PutState(appid, []byte(value))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}



func (t *Chaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("ex02 Invoke")
	function, args := stub.GetFunctionAndParameters()

	if function != "invoke" {
		return shim.Error("Unknown function call")
	}

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting at least 2")
	}

	if args[0] == "funcinit" {
		/*
		if err := stub.SetEvent("testEvent", []byte("Test Payload")); err != nil {
			return shim.Error("Unable to set CC event: testEvent. Aborting transaction ...")
		}
		*/
		return t.funcinit(stub, args)
	}

	if args[0] == "functransaction" {
		return t.functransaction(stub, args)
	}

	if args[0] == "funcdelete" {
		// Deletes an entity from its state
		return t.funcdelete(stub, args)
	}

	if args[0] == "funcquery" {
		// queries an entity state
		return t.funcquery(stub, args)
	}

	return shim.Error("Unknown action, check the first argument, must be one of 'delete', 'query', or 'move'")
}

func (t *Chaincode) funcinit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3{
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	var appid string
	var value string // Asset holdings
	var err error

	appid = args[1]
	value = args[2]
	fmt.Printf("appid=%s,value = %s\n", appid,value)
	// Write the state to the ledger
	err = stub.PutState(appid, []byte(value))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *Chaincode) funcdelete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	appid := args[1]

	// Delete the key from the state in ledger
	err := stub.DelState(appid)
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	return shim.Success(nil)
}

func (t *Chaincode) functransaction(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// must be an invoke
	var A, B string    // Entities
	var Aval, Bval int64 // Asset holdings
	var X int64          // Transaction value
	var err error
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4, function followed by 2 names and 1 value")
	}

	A = args[1]
	B = args[2]

	// Get the state from the ledger
	// TODO: will be nice to have a GetAllState call to ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Avalbytes == nil {
		err = stub.PutState(A, []byte("0"))
		if err != nil {
			return shim.Error("Entity not found")
			//return shim.Error(err.Error())
		}
		Avalbytes = []byte("0")
	}
	Aval, _ = strconv.ParseInt(string(Avalbytes), 10, 64)

	Bvalbytes, err := stub.GetState(B)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Bvalbytes == nil {
		err = stub.PutState(B, []byte("0"))
		if err != nil {
			return shim.Error("Entity not found")
			//return shim.Error(err.Error())
		}
		Bvalbytes = []byte("0")
	}
	Bval, _ = strconv.ParseInt(string(Bvalbytes), 10, 64)

	// Perform the execution
	X, err = strconv.ParseInt(args[3], 10, 64)
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}

	Aval = Aval - X
	if Aval < 0 {
		return shim.Error("score is not enough to use")
	}

	Bval = Bval + X
	fmt.Printf("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state back to the ledger
	err = stub.PutState(A, []byte(strconv.FormatInt(Aval,10)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.FormatInt(Bval,10)))
	if err != nil {
		return shim.Error(err.Error())
	}
	/*
	if transientMap, err := stub.GetTransient(); err == nil {
		if transientData, ok := transientMap["result"]; ok {
			fmt.Printf("Transient data in 'move' : %s\n", transientData)
			return shim.Success(transientData)
		}
	}
	*/
	return shim.Success(nil)
}


func (t *Chaincode) funcquery(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var appid string // Entities
	var err error

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	appid = args[1]

	// Get the state from the ledger
	value, err := stub.GetState(appid)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + string(value) + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(value)
}

func main() {
	err := shim.Start(new(Chaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
