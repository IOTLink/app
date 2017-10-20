package fabric

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"log"
	"math/rand"
	ca "github.com/hyperledger/fabric-sdk-go/api/apifabca"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	deffab "github.com/hyperledger/fabric-sdk-go/def/fabapi"

	"github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/admin"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	fabricTxn "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn"

	"time"
)

// GetOrdererAdmin returns a pre-enrolled orderer admin user
func GetOrdererAdmin(c fab.FabricClient, orgName string) (ca.User, error) {
	keyDir := "ordererOrganizations/example.com/users/Admin@example.com/msp/keystore"
	certDir := "ordererOrganizations/example.com/users/Admin@example.com/msp/signcerts"
	return getDefaultImplPreEnrolledUser(c, keyDir, certDir, "ordererAdmin", orgName)
}

// GetAdmin returns a pre-enrolled org admin user
func GetAdmin(c fab.FabricClient, orgPath string, orgName string) (ca.User, error) {
	keyDir := fmt.Sprintf("peerOrganizations/%s.example.com/users/Admin@%s.example.com/msp/keystore", orgPath, orgPath)
	certDir := fmt.Sprintf("peerOrganizations/%s.example.com/users/Admin@%s.example.com/msp/signcerts", orgPath, orgPath)
	username := fmt.Sprintf("peer%sAdmin", orgPath)
	return getDefaultImplPreEnrolledUser(c, keyDir, certDir, username, orgName)
}

// GetUser returns a pre-enrolled org user
func GetUser(c fab.FabricClient, orgPath string, orgName string) (ca.User, error) {
	keyDir := fmt.Sprintf("peerOrganizations/%s.example.com/users/User1@%s.example.com/msp/keystore", orgPath, orgPath)
	certDir := fmt.Sprintf("peerOrganizations/%s.example.com/users/User1@%s.example.com/msp/signcerts", orgPath, orgPath)
	username := fmt.Sprintf("peer%sUser1", orgPath)
	return getDefaultImplPreEnrolledUser(c, keyDir, certDir, username, orgName)
}


// GetDefaultImplPreEnrolledUser ...
func getDefaultImplPreEnrolledUser(client fab.FabricClient, keyDir string, certDir string, username string, orgName string) (ca.User, error) {
	privateKeyDir := filepath.Join(client.Config().CryptoConfigPath(), keyDir)
	privateKeyPath, err := getFirstPathFromDir(privateKeyDir)
	if err != nil {
		return nil, fmt.Errorf("Error finding the private key path: %v", err)
	}

	enrollmentCertDir := filepath.Join(client.Config().CryptoConfigPath(), certDir)
	enrollmentCertPath, err := getFirstPathFromDir(enrollmentCertDir)
	if err != nil {
		return nil, fmt.Errorf("Error finding the enrollment cert path: %v", err)
	}
	mspID, err := client.Config().MspID(orgName)
	if err != nil {
		return nil, fmt.Errorf("Error reading MSP ID config: %s", err)
	}
	return deffab.NewPreEnrolledUser(client.Config(), privateKeyPath, enrollmentCertPath, username, mspID, client.CryptoSuite())
}


// Gets the first path from the dir directory
func getFirstPathFromDir(dir string) (string, error) {

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("Could not read directory %s, err %s", err, dir)
	}

	for _, p := range files {
		if p.IsDir() {
			continue
		}

		fullName := filepath.Join(dir, string(filepath.Separator), p.Name())
		fmt.Printf("Reading file %s\n", fullName)
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		fullName := filepath.Join(dir, string(filepath.Separator), f.Name())
		return fullName, nil
	}

	return "", fmt.Errorf("No paths found in directory: %s", dir)
}

// HasPrimaryPeerJoinedChannel checks whether the primary peer of a channel
// has already joined the channel. It returns true if it has, false otherwise,
// or an error
func HasPrimaryPeerJoinedChannel(client fab.FabricClient, orgUser ca.User, channel fab.Channel) (bool, error) {
	foundChannel := false
	primaryPeer := channel.PrimaryPeer()

	currentUser := client.UserContext()
	defer client.SetUserContext(currentUser)

	client.SetUserContext(orgUser)
	response, err := client.QueryChannels(primaryPeer)
	if err != nil {
		return false, fmt.Errorf("Error querying channel for primary peer: %s", err)
	}
	log.Println("channel num: ",len(response.Channels))
	for _, responseChannel := range response.Channels {
		log.Println("channel ----------> : ",responseChannel.ChannelId)
		if responseChannel.ChannelId == channel.Name() {
			foundChannel = true
		}
	}
	return foundChannel, nil
}


// InstallAndInstantiateExampleCC ..
func (setup *BaseSetupImpl) InitCC(chaincodepath string, chaincode string , version string, key string, payload string) error {

	chainCodePath := chaincodepath //"github.com/chaincode"
	chainCodeVersion := version    //"v1"
	setup.ChainCodeID = chaincode


	if err := setup.InstallCC(setup.ChainCodeID, chainCodePath, chainCodeVersion, nil); err != nil {
		return err
	}

	var args []string
	args = append(args, "init")
	args = append(args,  key)
	args = append(args, payload)

	return setup.InstantiateCC(setup.ChainCodeID, chainCodePath, chainCodeVersion, args)
}

// InstallCC ...
func (setup *BaseSetupImpl) InstallCC(chainCodeID string, chainCodePath string, chainCodeVersion string, chaincodePackage []byte) error {

	// installCC requires AdminUser privileges so setting user context with Admin User
	setup.Client.SetUserContext(setup.AdminUser)

	// must reset client user context to normal user once done with Admin privilieges
	defer setup.Client.SetUserContext(setup.NormalUser)

	//fmt.Println("success InstallCC")
	if err := admin.SendInstallCC(setup.Client, chainCodeID, chainCodePath, chainCodeVersion, chaincodePackage, setup.Channel.Peers(), setup.GetDeployPath()); err != nil {
		return fmt.Errorf("SendInstallProposal return error: %v", err)
	}

	return nil
}

// InstantiateCC ...
func (setup *BaseSetupImpl) InstantiateCC(chainCodeID string, chainCodePath string, chainCodeVersion string, args []string) error {
	// InstantiateCC requires AdminUser privileges so setting user context with Admin User
	setup.Client.SetUserContext(setup.AdminUser)

	// must reset client user context to normal user once done with Admin privilieges
	defer setup.Client.SetUserContext(setup.NormalUser)

	chaincodePolicy := cauthdsl.SignedByMspMember(setup.Client.UserContext().MspID())

	if err := admin.SendInstantiateCC(setup.Channel, chainCodeID, args, chainCodePath, chainCodeVersion, chaincodePolicy, []apitxn.ProposalProcessor{setup.Channel.PrimaryPeer()}, setup.EventHub); err != nil {
		return err
	}
	return nil
}

func (setup *BaseSetupImpl) GetDeployPath() string {
	pwd, _ := os.Getwd()
	return path.Join(pwd, "../")
}

func (setup *BaseSetupImpl) InvokeFunc(client fab.FabricClient, channel fab.Channel, targets []apitxn.ProposalProcessor,
	eventHub fab.EventHub, chaincodeID string, fcn string, args []string, transientData map[string][]byte) (string, error){

	transactionID, err:= fabricTxn.InvokeChaincode(client, channel, targets, eventHub, chaincodeID, fcn, args, transientData)
	if err != nil {
		return "", err
	}
	return transactionID.ID, nil
}


func (setup *BaseSetupImpl) InvokeFuncQuery(client fab.FabricClient, channel fab.Channel, chaincodeID string, fcn string, args []string) (string, error) {
	return fabricTxn.QueryChaincode(client, channel, chaincodeID, fcn, args)
}

// GenerateRandomID generates random ID
func GenerateRandomID() string {
	rand.Seed(time.Now().UnixNano())
	return randomString(10)
}

// Utility to create random string of strlen length
func randomString(strlen int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

