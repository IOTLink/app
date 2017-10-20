package main

import (
	"fmt"
	"strconv"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/op/go-logging"
	"bytes"
	"encoding/pem"
	"crypto/x509"

)

type Chaincode struct{

}

var (
	logger = logging.MustGetLogger("hydruscoin")
)


func  (t *Chaincode) Init(stub shim.ChaincodeStubInterface) pb.Response{
	logger.Debug("deploy Chaincode successfully")
	return shim.Success(nil)
}



func (t *Chaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("ex02 Invoke")
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
		logger.Debug("Invoke funcinit")
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

	//get creator
	//获取交易用户的证书
	creatorByte,_:= stub.GetCreator()
	certStart := bytes.IndexAny(creatorByte, "-----BEGIN")
	if certStart == -1 {
		logger.Debug("No certificate found")
	}
	certText := creatorByte[certStart:]
	bl, _ := pem.Decode(certText)
	if bl == nil {
		logger.Debug("Could not decode the PEM structure")
	}

	cert, err := x509.ParseCertificate(bl.Bytes)
	if err != nil {
		logger.Debug("ParseCertificate failed")
	}
	uname:=cert.Subject.CommonName
	org := cert.Subject.Organization
	cn := cert.Subject.Country
	serialnum := cert.Subject.SerialNumber
	extraname := cert.Subject.ExtraNames
	logger.Debug("Name:"+uname," Organization:",org," Country:",cn, " SerialNumber:",serialnum, " ExtraNames:",extraname)

	signedPro ,_:= stub.GetSignedProposal() //获取签名数据，在chiancode中验证，获取验证数据，需要数据加密算法，暂时未作
	binding, err := stub.GetBinding()
	if err != nil {
		return shim.Success([]byte("Failed getting binding"))
	}

	//logger.Debug("passed certificate:", cert)
	logger.Debug("passed sigma:", signedPro.Signature)
	logger.Debug("passed payload:", args[2])
	logger.Debug("passed binding:", binding)

	/*
	err = cert.CheckSignature(x509.SHA256WithRSA, signedPro.Signature, []byte(args[2]))
	if err != nil {
		logger.Debug("Failed checking signature",err.Error())
	} else {
		logger.Debug("success checking signature")
	}
	*/

	/*
	//https://gist.github.com/jedy/5963633
	//https://www.socketloop.com/references/golang-crypto-x509-certificate-checksignature-function-example

	ok, err := stub.VerifySignature(
		cert,
		signedPro.Signature,
		append(args[2], binding...),
	)


	err = cert.CheckSignature(
		cert,
		signedPro.Signature,
		append(args[2], binding...),
	)
	if err != nil {
		logger.Debug("Failed checking signature",err.Error())
	}
	if !ok {
		logger.Error("Invalid signature")
	}
	*/
	//_____________________________


	times,_ := stub.GetTxTimestamp()
	logger.Debug("TxID", stub.GetTxID(), " TxTimeStamp:",times.String())

	var appid string
	var value string // Asset holdings
	//var err error

	appid = args[1]
	value = args[2]
	logger.Debug("key-appid:",appid, "value:", value)
	// Write the state to the ledger
	err = stub.PutState(appid, []byte(value))
	if err != nil {
		return shim.Error(err.Error())
	}

	if transientMap, err := stub.GetTransient(); err == nil {
		for k, v := range transientMap {
			logger.Debug(k, string(v))
		}
		//return shim.Success([]byte("test success"))
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
