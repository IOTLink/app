package fabric

import (
	"log"
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"errors"
	coin "app/coin"
	"github.com/golang/protobuf/proto"
)

type FabricServer struct {
	SetupMap *BaseSetupImpl
}

func (fabric *FabricServer) Init(OrgID string, enroll_user_dir string) error {
	fabric.SetupMap = &BaseSetupImpl{
		ConfigFile:      "config.yaml",
		ChannelID:       "mychannel",
		OrgID:           OrgID, //"peerorg1",
		ChannelConfig:   "channel.tx",
		ChainCodeID:     "utxo",
		ConnectEventHub: true,
	}

	if err := fabric.SetupMap.Initialize(enroll_user_dir); err != nil {
		log.Println("fabric server init abort %s",err.Error())
		return err
	}
	return nil
}

//install chaincode
func (fabric *FabricServer)InitAsset(channelid string, chaincodepath, chaincode, chaincodeversion string ,key string, payload string) error{
	if channelid == "" || chaincode == "" || chaincodeversion == "" {
		return errors.New("parameter can not be empty")
	}

	baseSetup := fabric.SetupMap
	if err := baseSetup.InitCC(chaincodepath, chaincode, chaincodeversion, key, payload); err != nil {
		log.Println("install chaincode return error: %v", err)
		return err
	}

	return nil
}

func (fabric *FabricServer) InvokeInit(channelid string, chaincode string, appid string, key string, payload string) (string,error) {
	if channelid == "" || chaincode == "" || appid == "" || key == "" || payload == "" {
		return fmt.Sprintf("%s","parameter can not be empty"), errors.New("parameter can not be empty")
	}
	fmt.Println("Channel Nmae:",channelid)
	var txID string
	var err error
	fcn := "invoke"

	var args []string
	args = append(args, "funcinit")
	args = append(args, key)
	args = append(args, payload)

	transientDataMap := make(map[string][]byte)
	transientDataMap["result"] = []byte("Transient data in move funds...")

	baseSetup := fabric.SetupMap

	sdk  := baseSetup.FabricSDK
	client, nil := sdk.NewSystemClient(nil)
	_, err = client.LoadUserFromStateStore(appid)
	if err != nil {
		log.Println(appid, "LoadUserFromStateStore ERROR:",err.Error())
	} else {
		log.Println("success load user appid:", appid)
	}

	channel, err := baseSetup.GetChannel(client, channelid, []string{baseSetup.OrgID})
	if err != nil {
		return fmt.Sprintf("Create channel %s failed: %v", channelid, err), err
	}
	txID, err = baseSetup.InvokeFunc(client, channel, []apitxn.ProposalProcessor{channel.PrimaryPeer()}, baseSetup.EventHub, chaincode, fcn, args, transientDataMap)
	//txID, err = baseSetup.InvokeFunc(baseSetup.Client, baseSetup.Channel, []apitxn.ProposalProcessor{baseSetup.Channel.PrimaryPeer()}, baseSetup.EventHub, chaincode, fcn, args, transientDataMap)

	return 	txID, err
}

func (fabric *FabricServer) InvokeTransaction(channelid string, chaincode string , ownerid string, receiverid string, payload string) (string, error) {
	var txID string
	var err error
	fcn := "invoke"

	var args []string
	args = append(args, "functransaction")
	args = append(args, ownerid)
	args = append(args, receiverid)
	args = append(args, payload)

	transientDataMap := make(map[string][]byte)
	transientDataMap["result"] = []byte("Transient data in move funds...")

	baseSetup := fabric.SetupMap
	sdk  := baseSetup.FabricSDK
	client, nil := sdk.NewSystemClient(nil)
	_, err = client.LoadUserFromStateStore(ownerid)
	if err != nil {
		log.Println(ownerid, "LoadUserFromStateStore ERROR:",err.Error())
	} else {
		log.Println("success load user appid:", ownerid)
	}

	channel, err := baseSetup.GetChannel(client, channelid, []string{baseSetup.OrgID})
	if err != nil {
		return fmt.Sprintf("Create channel %s failed: %v", channelid, err), err
	}
	txID, err = baseSetup.InvokeFunc(client, channel, []apitxn.ProposalProcessor{channel.PrimaryPeer()}, baseSetup.EventHub, chaincode, fcn, args, transientDataMap)

	return 	txID, err
}

func (fabric *FabricServer) InvokeQuery(channelid string, chaincode string, appid string, key string) (string,error) {
	var payload string
	var err error
	fcn := "invoke"

	var args []string
	args = append(args, "funcquery")
	args = append(args, key)

	baseSetup := fabric.SetupMap
	Client  := baseSetup.Client
	Channel := baseSetup.Channel
	payload, err = baseSetup.InvokeFuncQuery(Client, Channel, chaincode, fcn, args)

	return 	payload, err
}

/////////////////////////////////////////////////////////////////////////////////

func (fabric *FabricServer) InvokeRegister(channelid string, chaincode string, appid string) (string,error) {
	if channelid == "" || chaincode == "" || appid == "" {
		return fmt.Sprintf("%s","parameter can not be empty"), errors.New("parameter can not be empty")
	}
	fmt.Println("Channel Nmae:",channelid)
	var txID string
	var err error
	fcn := "invoke"

	var args []string
	args = append(args, "funcinit")
	args = append(args, appid)
	args = append(args, "1000")

	transientDataMap := make(map[string][]byte)
	transientDataMap["key1"] = []byte("value1")
	transientDataMap["key2"] = []byte("value2")
	transientDataMap["key3"] = []byte("value3")

	baseSetup := fabric.SetupMap

	sdk  := baseSetup.FabricSDK
	client, nil := sdk.NewSystemClient(nil)
	_, err = client.LoadUserFromStateStore(appid)
	if err != nil {
		log.Println(appid, "LoadUserFromStateStore ERROR:",err.Error())
	} else {
		log.Println("success load user appid:", appid)
	}

	channel, err := baseSetup.GetChannel(client, channelid, []string{baseSetup.OrgID})
	if err != nil {
		return fmt.Sprintf("Create channel %s failed: %v", channelid, err), err
	}
	txID, err = baseSetup.InvokeFunc(client, channel, []apitxn.ProposalProcessor{channel.PrimaryPeer()}, baseSetup.EventHub, chaincode, fcn, args, transientDataMap)
	//txID, err = baseSetup.InvokeFunc(client, channel, []apitxn.ProposalProcessor{channel.PrimaryPeer()}, baseSetup.EventHub, chaincode, fcn, args, transientDataMap)
	//txID, err = baseSetup.InvokeFunc(baseSetup.Client, baseSetup.Channel, []apitxn.ProposalProcessor{baseSetup.Channel.PrimaryPeer()}, baseSetup.EventHub, chaincode, fcn, args, transientDataMap)

	return 	txID, err
}

func (fabric *FabricServer) InvokeCoinbase(channelid string, chaincode string, appid string, tx string) (string,error) {
	if channelid == "" || chaincode == "" || appid == "" {
		return fmt.Sprintf("%s","parameter can not be empty"), errors.New("parameter can not be empty")
	}
	fmt.Println("Channel Nmae:",channelid)
	var txID string
	var err error
	fcn := "invoke"

	var args []string
	args = append(args, "invoke_coinbase")
	args = append(args, tx)

	transientDataMap := make(map[string][]byte)
	//transientDataMap["result"] = []byte("Transient data in move funds...")

	baseSetup := fabric.SetupMap

	sdk  := baseSetup.FabricSDK
	client, nil := sdk.NewSystemClient(nil)
	_, err = client.LoadUserFromStateStore(appid)
	if err != nil {
		log.Println(appid, "LoadUserFromStateStore ERROR:",err.Error())
	} else {
		log.Println("success load user appid:", appid)
	}

	channel, err := baseSetup.GetChannel(client, channelid, []string{baseSetup.OrgID})
	if err != nil {
		return fmt.Sprintf("Create channel %s failed: %v", channelid, err), err
	}
	txID, err = baseSetup.InvokeFunc(client, channel, []apitxn.ProposalProcessor{channel.PrimaryPeer()}, baseSetup.EventHub, chaincode, fcn, args, transientDataMap)
	//txID, err = baseSetup.InvokeFunc(client, channel, []apitxn.ProposalProcessor{channel.PrimaryPeer()}, baseSetup.EventHub, chaincode, fcn, args, transientDataMap)
	//txID, err = baseSetup.InvokeFunc(baseSetup.Client, baseSetup.Channel, []apitxn.ProposalProcessor{baseSetup.Channel.PrimaryPeer()}, baseSetup.EventHub, chaincode, fcn, args, transientDataMap)

	return 	txID, err
}



func (fabric *FabricServer) InvokeTransfer(channelid string, chaincode string, appid string, tx string) (string,error) {
	if channelid == "" || chaincode == "" || appid == "" {
		return fmt.Sprintf("%s","parameter can not be empty"), errors.New("parameter can not be empty")
	}
	fmt.Println("Channel Nmae:",channelid)
	var txID string
	var err error
	fcn := "invoke"

	var args []string
	args = append(args, "invoke_transfer")
	args = append(args, tx)

	transientDataMap := make(map[string][]byte)
	//transientDataMap["result"] = []byte("Transient data in move funds...")

	baseSetup := fabric.SetupMap

	sdk  := baseSetup.FabricSDK
	client, nil := sdk.NewSystemClient(nil)
	_, err = client.LoadUserFromStateStore(appid)
	if err != nil {
		log.Println(appid, "LoadUserFromStateStore ERROR:",err.Error())
	} else {
		log.Println("success load user appid:", appid)
	}

	channel, err := baseSetup.GetChannel(client, channelid, []string{baseSetup.OrgID})
	if err != nil {
		return fmt.Sprintf("Create channel %s failed: %v", channelid, err), err
	}
	txID, err = baseSetup.InvokeFunc(client, channel, []apitxn.ProposalProcessor{channel.PrimaryPeer()}, baseSetup.EventHub, chaincode, fcn, args, transientDataMap)
	//txID, err = baseSetup.InvokeFunc(client, channel, []apitxn.ProposalProcessor{channel.PrimaryPeer()}, baseSetup.EventHub, chaincode, fcn, args, transientDataMap)
	//txID, err = baseSetup.InvokeFunc(baseSetup.Client, baseSetup.Channel, []apitxn.ProposalProcessor{baseSetup.Channel.PrimaryPeer()}, baseSetup.EventHub, chaincode, fcn, args, transientDataMap)
	return 	txID, err
}


func (fabric *FabricServer) QueryAddrs(channelid string, chaincode string, appid string) (*coin.QueryAddrResults, error) {
	var payload string
	var err error
	fcn := "invoke"

	var args []string
	args = append(args, "query_addrs")
	args = append(args, appid)

	baseSetup := fabric.SetupMap
	Client  := baseSetup.Client
	Channel := baseSetup.Channel
	payload, err = baseSetup.InvokeFuncQuery(Client, Channel, chaincode, fcn, args)
	if err != nil {
		return nil,err
	}

	AddrResults := new(coin.QueryAddrResults)
	if err := proto.Unmarshal([]byte(payload), AddrResults); err != nil {
		return nil, err
	}
	return 	AddrResults, err
}

func (fabric *FabricServer) QueryTx(channelid string, chaincode string, txHash string) (*coin.TX, error) {
	var payload string
	var err error
	fcn := "invoke"

	var args []string
	args = append(args, "query_tx")
	args = append(args, txHash)

	baseSetup := fabric.SetupMap
	Client  := baseSetup.Client
	Channel := baseSetup.Channel
	payload, err = baseSetup.InvokeFuncQuery(Client, Channel, chaincode, fcn, args)
	if err != nil {
		return nil,err
	}

	tx := new(coin.TX)
	if err := proto.Unmarshal([]byte(payload), tx); err != nil {
		return nil, err
	}
	return 	tx, err
}

func (fabric *FabricServer) QueryCoin(channelid string, chaincode string) (*coin.HydruscoinInfo, error) {
	var payload string
	var err error
	fcn := "invoke"

	var args []string
	args = append(args, "query_coin")
	args = append(args, "")

	baseSetup := fabric.SetupMap
	Client  := baseSetup.Client
	Channel := baseSetup.Channel
	payload, err = baseSetup.InvokeFuncQuery(Client, Channel, chaincode, fcn, args)
	if err != nil {
		return nil,err
	}

	coinInfo := new(coin.HydruscoinInfo)
	if err := proto.Unmarshal([]byte(payload), coinInfo); err != nil {
		return nil, err
	}
	return 	coinInfo, err
}



