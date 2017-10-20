package fabric

import (
	"fmt"
	//"os"
	//"path"
	"time"
	"log"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	ca "github.com/hyperledger/fabric-sdk-go/api/apifabca"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	deffab "github.com/hyperledger/fabric-sdk-go/def/fabapi"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/opt"
	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/orderer"
	//fabricTxn "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn"
	admin "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/admin"
	//"github.com/hyperledger/fabric/common/cauthdsl"
	pb "github.com/hyperledger/fabric/protos/peer"
	//"github.com/cloudflare/cfssl/log"
	"encoding/hex"
)

// BaseSetupImpl implementation of BaseTestSetup
// BaseSetupImpl implementation of BaseTestSetup
type BaseSetupImpl struct {
	Client          fab.FabricClient
	FabricSDK       *deffab.FabricSDK
	Channel         fab.Channel
	EventHub        fab.EventHub
	ConnectEventHub bool
	ConfigFile      string
	OrgID           string
	EnrollUserDir   string
	ChannelID       string
	ChainCodeID     string
	Initialized     bool
	ChannelConfig   string
	AdminUser       ca.User
	NormalUser      ca.User
}

var org map[string]string
func init() {
	org = make(map[string]string)
	org["peerorg1"] = "org1"
	org["peerorg2"] = "org2"
}

// Initialize reads configuration from file and sets up client, channel and event hub
func (setup *BaseSetupImpl) Initialize(enroll_user_dir string) error {
	// Create SDK setup for the integration tests
	sdkOptions := deffab.Options{
		ConfigFile: setup.ConfigFile,
		StateStoreOpts: opt.StateStoreOpts{
			//Path: "enroll_user",
			Path: enroll_user_dir,
		},
	}
	sdk, err := deffab.NewSDK(sdkOptions)
	if err != nil {
		return fmt.Errorf("Error initializing SDK: %s", err)
	}
	setup.FabricSDK = sdk

	context, err := sdk.NewContext(setup.OrgID)
	if err != nil {
		return fmt.Errorf("Error getting a context for org: %s", err)
	}
	log.Println("Init Fabric Server OrgID:", setup.OrgID)

	user, err := deffab.NewUser(sdk.ConfigProvider(), context.MSPClient(), "admin", "adminpw", setup.OrgID)
	//user, err := deffab.NewUser(sdk.ConfigProvider(), context.MSPClient(), "admin", "adminpw", setup.MspID)
	if err != nil {
		return fmt.Errorf("NewUser returned error: %v", err)
	}

	session1, err := sdk.NewSession(context, user)
	if err != nil {
		return fmt.Errorf("NewSession returned error: %v", err)
	}
	sc, err := sdk.NewSystemClient(session1)
	if err != nil {
		return fmt.Errorf("NewSystemClient returned error: %v", err)
	}


	err = sc.SaveUserToStateStore(user, false)
	if err != nil {
		return fmt.Errorf("@ client.SaveUserToStateStore returned error: %v", err)
	}
	setup.Client = sc

	prikey := session1.Identity().PrivateKey().SKI()
	log.Println("===> session1 user.name:",session1.Identity().Name(), " user.prikey:",hex.EncodeToString(prikey[:]), prikey)


	/*
	org1Admin, err := GetAdmin(sc, "org1", setup.OrgID)
	if err != nil {
		return fmt.Errorf("Error getting org admin user: %v", err)
	}

	org1User, err := GetUser(sc, "org1", setup.OrgID)
	if err != nil {
		return fmt.Errorf("Error getting org user: %v", err)
	}
	*/
	org1Admin, err := GetAdmin(sc, org[setup.OrgID], setup.OrgID)
	if err != nil {
		return fmt.Errorf("Error getting org admin user: %v", err)
	}

	org1User, err := GetUser(sc, org[setup.OrgID], setup.OrgID)
	if err != nil {
		return fmt.Errorf("Error getting org user: %v", err)
	}

	setup.AdminUser = org1Admin
	setup.NormalUser = org1User

	channel, err := setup.GetChannel(setup.Client, setup.ChannelID, []string{setup.OrgID})
	if err != nil {
		return fmt.Errorf("Create channel (%s) failed: %v", setup.ChannelID, err)
	}
	setup.Channel = channel

	ordererAdmin, err := GetOrdererAdmin(sc, setup.OrgID)
	if err != nil {
		return fmt.Errorf("Error getting orderer admin user: %v", err)
	}

	// Check if primary peer has joined channel
	alreadyJoined, err := HasPrimaryPeerJoinedChannel(sc, org1Admin, channel)
	if err != nil {
		return fmt.Errorf("Error while checking if primary peer has already joined channel: %v", err)
	}

	if !alreadyJoined {
		// Create, initialize and join channel
		if err = admin.CreateOrUpdateChannel(sc, ordererAdmin, org1Admin, channel, setup.ChannelConfig); err != nil {
			return fmt.Errorf("CreateChannel returned error: %v", err)
		}
		time.Sleep(time.Second * 3)

		sc.SetUserContext(org1Admin)
		if err = channel.Initialize(nil); err != nil {
			return fmt.Errorf("Error initializing channel: %v", err)
		}

		if err = admin.JoinChannel(sc, org1Admin, channel); err != nil {
			return fmt.Errorf("JoinChannel returned error: %v", err)
		}
	}

	//by default client's user context should use regular user, for admin actions, UserContext must be set to AdminUser
	sc.SetUserContext(org1User)
	//sc.SetUserContext(org1Admin)
	if err := setup.setupEventHub(sc); err != nil {
		return err
	}

	setup.Initialized = true

	return nil
}


// CreateAndSendTransactionProposal ... TODO duplicate
func (setup *BaseSetupImpl) CreateAndSendTransactionProposal(channel fab.Channel, chainCodeID string,
	fcn string, args []string, targets []apitxn.ProposalProcessor, transientData map[string][]byte) ([]*apitxn.TransactionProposalResponse, apitxn.TransactionID, error) {

	request := apitxn.ChaincodeInvokeRequest{
		Targets:      targets,
		Fcn:          fcn,
		Args:         args,
		TransientMap: transientData,
		ChaincodeID:  chainCodeID,
	}
	transactionProposalResponses, txnID, err := channel.SendTransactionProposal(request)
	if err != nil {
		return nil, txnID, err
	}

	for _, v := range transactionProposalResponses {
		if v.Err != nil {
			return nil, txnID, fmt.Errorf("invoke Endorser %s returned error: %v", v.Endorser, v.Err)
		}
	}

	return transactionProposalResponses, txnID, nil
}

// CreateAndSendTransaction ...
func (setup *BaseSetupImpl) CreateAndSendTransaction(channel fab.Channel, resps []*apitxn.TransactionProposalResponse) (*apitxn.TransactionResponse, error) {

	tx, err := channel.CreateTransaction(resps)
	if err != nil {
		return nil, fmt.Errorf("CreateTransaction return error: %v", err)
	}

	transactionResponse, err := channel.SendTransaction(tx)
	if err != nil {
		return nil, fmt.Errorf("SendTransaction return error: %v", err)

	}

	if transactionResponse.Err != nil {
		return nil, fmt.Errorf("Orderer %s return error: %v", transactionResponse.Orderer, transactionResponse.Err)
	}

	return transactionResponse, nil
}

func (setup *BaseSetupImpl) setupEventHub(client fab.FabricClient) error {
	eventHub, err := setup.getEventHub(client)
	if err != nil {
		return err
	}

	if setup.ConnectEventHub {
		if err := eventHub.Connect(); err != nil {
			return fmt.Errorf("Failed eventHub.Connect() [%s]", err)
		}
	}
	setup.EventHub = eventHub

	return nil
}

// InitConfig ...
func (setup *BaseSetupImpl) InitConfig() (apiconfig.Config, error) {
	configImpl, err := config.InitConfig(setup.ConfigFile)
	if err != nil {
		return nil, err
	}
	return configImpl, nil
}


// GetChannel initializes and returns a channel based on config
func (setup *BaseSetupImpl) GetChannel(client fab.FabricClient, channelID string, orgs []string) (fab.Channel, error) {

	channel, err := client.NewChannel(channelID)
	if err != nil {
		return nil, fmt.Errorf("NewChannel return error: %v", err)
	}

	ordererConfig, err := client.Config().RandomOrdererConfig()
	if err != nil {
		return nil, fmt.Errorf("RandomOrdererConfig() return error: %s", err)
	}

	orderer, err := orderer.NewOrderer(fmt.Sprintf("%s:%d", ordererConfig.Host,
		ordererConfig.Port), ordererConfig.TLS.Certificate,
		ordererConfig.TLS.ServerHostOverride, client.Config())
	if err != nil {
		return nil, fmt.Errorf("NewOrderer return error: %v", err)
	}
	err = channel.AddOrderer(orderer)
	if err != nil {
		return nil, fmt.Errorf("Error adding orderer: %v", err)
	}

	for _, org := range orgs {
		peerConfig, err := client.Config().PeersConfig(org)
		if err != nil {
			return nil, fmt.Errorf("Error reading peer config: %v", err)
		}
		for _, p := range peerConfig {
			endorser, err := deffab.NewPeer(fmt.Sprintf("%s:%d", p.Host, p.Port),
				p.TLS.Certificate, p.TLS.ServerHostOverride, client.Config())
			if err != nil {
				return nil, fmt.Errorf("NewPeer return error: %v", err)
			}
			err = channel.AddPeer(endorser)
			if err != nil {
				return nil, fmt.Errorf("Error adding peer: %v", err)
			}
			if p.Primary {
				channel.SetPrimaryPeer(endorser)
			}
		}
	}

	return channel, nil
}



// RegisterTxEvent registers on the given eventhub for the give transaction
// returns a boolean channel which receives true when the event is complete
// and an error channel for errors
// TODO - Duplicate
func (setup *BaseSetupImpl) RegisterTxEvent(txID apitxn.TransactionID, eventHub fab.EventHub) (chan bool, chan error) {
	done := make(chan bool)
	fail := make(chan error)

	eventHub.RegisterTxEvent(txID, func(txId string, errorCode pb.TxValidationCode, err error) {
		if err != nil {
			fmt.Printf("Received error event for txid(%s)\n", txId)
			fail <- err
		} else {
			fmt.Printf("Received success event for txid(%s)\n", txId)
			done <- true
		}
	})

	return done, fail
}

// getEventHub initilizes the event hub
func (setup *BaseSetupImpl) getEventHub(client fab.FabricClient) (fab.EventHub, error) {
	eventHub, err := events.NewEventHub(client)
	if err != nil {
		return nil, fmt.Errorf("Error creating new event hub: %v", err)
	}
	foundEventHub := false
	peerConfig, err := client.Config().PeersConfig(setup.OrgID)
	if err != nil {
		return nil, fmt.Errorf("Error reading peer config: %v", err)
	}
	for _, p := range peerConfig {
		if p.EventHost != "" && p.EventPort != 0 {
			fmt.Printf("******* EventHub connect to peer (%s:%d) *******\n", p.EventHost, p.EventPort)
			eventHub.SetPeerAddr(fmt.Sprintf("%s:%d", p.EventHost, p.EventPort),
				p.TLS.Certificate, p.TLS.ServerHostOverride)
			foundEventHub = true
			break
		}
	}

	if !foundEventHub {
		return nil, fmt.Errorf("No EventHub configuration found")
	}

	return eventHub, nil
}


