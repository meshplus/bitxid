package did

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/bitxhub/bitxid/pkg/common/types"
	"github.com/bitxhub/bitxid/pkg/loggers"
	"github.com/bitxhub/bitxid/pkg/repo"
	"github.com/meshplus/bitxhub-kit/storage/leveldb"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/sha3"
)

const (
	dbPath   string = "../../config/doc.db"
	rtPath   string = "../../config/registry.store"
	confPath string = "../../config"
)

var R *Registry
var caller string = "0x12345678"
var did types.DID = types.DID("did:bitxhub:relayroot:0x12345678")
var doc string = `{	"id":"did:bitxhub:relayroot:0x12345678",
	"type": "user", 
	"publicKey":[
	{"type": "Secp256k1",
	"publicKeyPem": "02b97c30de767f084ce3080168ee293053ba33b235d7116a3263d29f1450936b71"
	}]
}`
var diddoc []byte = []byte(doc)
var sig []byte = []byte("0x9AB567")

var docB string = `{	"id":"did:bitxhub:relayroot:0x12345678",
	"type": "user", 
	"publicKey":[{
		"id":"KEY#1",
		"type": "Ed25519",
		"publicKeyPem": "H3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPV"
	}],
	"authentication":[
		{"publicKey":["KEY#1"]}
	]
}`
var diddocB []byte = []byte(docB)

func Test(t *testing.T) {
	// did := "did:bitxhub:appchain001:."
	// didT := types.DID(did)
	// fmt.Println([]byte(did), "\n", []byte(didT))
	// fmt.Println(doc)

	hash1 := sha3.Sum256([]byte("班委"))

	fmt.Printf("%X\n", hash1)
}

func TestNew(t *testing.T) {
	loggers.Init()
	L := loggers.Get(loggers.DID)
	S1, err := leveldb.New(rtPath)
	assert.Nil(t, err)
	S2, err := leveldb.New(dbPath)
	assert.Nil(t, err)
	C, err := repo.UnmarshalConfig(confPath)
	assert.Nil(t, err)
	R, err = New(S1, S2, L, &C.DIDConfig)
	assert.Nil(t, err)
}

func TestSetupGenesisSucceed(t *testing.T) {
	err := R.SetupGenesis()
	assert.Nil(t, err)
}

func TestHasMethodSucceed(t *testing.T) {
	ret1, err := R.HasDID(types.DID(R.config.GenesisAdmin))
	assert.Nil(t, err)
	assert.Equal(t, true, ret1)
}

func TestRegisterSucceed(t *testing.T) {
	docHashE := sha3.Sum512(diddoc)
	docAddrE := "./" + string(did)
	docAddr, docHash, err := R.Register(caller, did, diddoc, sig)
	assert.Nil(t, err)
	strHash := fmt.Sprintf("%x", docHash)
	strHashE := fmt.Sprintf("%x", docHashE)
	assert.Equal(t, strHashE, strHash)
	assert.Equal(t, docAddrE, docAddr)
}

func TestUpdateSucceed(t *testing.T) {
	docHashE := sha3.Sum512(diddocB)
	docAddrE := "./" + string(did)
	docAddr, docHash, err := R.Update(caller, did, diddocB, sig)
	assert.Nil(t, err)
	strHash := fmt.Sprintf("%x", docHash)
	strHashE := fmt.Sprintf("%x", docHashE)
	assert.Equal(t, strHashE, strHash)
	assert.Equal(t, docAddrE, docAddr)
}

func TestResolveSucceed(t *testing.T) {
	item, doc, err := R.Resolve(did)
	docHashE := sha3.Sum512(diddocB)
	assert.Nil(t, err)
	assert.Equal(t, diddocB, doc)
	itemE := Item{
		Identifier: did,
		DocHash:    docHashE[:],
		DocAddr:    "./" + string(did),
		Status:     Normal,
	}
	assert.Equal(t, itemE.Identifier, item.Identifier)
	assert.Equal(t, itemE.DocAddr, item.DocAddr)
	assert.Equal(t, itemE.DocHash, item.DocHash)
	assert.Equal(t, itemE.Status, item.Status)
}

func TestUnMarshalSucceed(t *testing.T) {
	_, doc, err := R.Resolve(did)
	assert.Nil(t, err)
	docE := Doc{}
	docE.ID = "did:bitxhub:relayroot:0x12345678"
	docE.Type = "user"
	pk := types.PubKey{
		ID:           "KEY#1",
		Type:         "Ed25519",
		PublicKeyPem: "H3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPV",
	}
	docE.PublicKey = []types.PubKey{pk}
	auth := types.Auth{
		PublicKey: []string{"KEY#1"},
	}
	docE.Authentication = []types.Auth{auth}
	// Unmarshal doc json byte to doc struct:
	docR := Doc{}
	err = json.Unmarshal(doc, &docR)
	assert.Nil(t, err)

	assert.Equal(t, docR, docE)
}

func TestFreezeSucceed(t *testing.T) {
	err := R.Freeze(did, did, sig)
	assert.Nil(t, err)
	item, _, err := R.Resolve(did)
	assert.Nil(t, err)
	assert.Equal(t, Frozen, item.Status)
}

func TestUnFreezeSucceed(t *testing.T) {
	err := R.UnFreeze(did, did, sig)
	assert.Nil(t, err)
	item, _, err := R.Resolve(did)
	assert.Nil(t, err)
	assert.Equal(t, Normal, item.Status)
}

func TestDeleteSucceed(t *testing.T) {
	err := R.Delete(did, sig)
	assert.Nil(t, err)
	err = R.Delete("did:bitxhub:relayroot:0x12348848", sig)
	assert.Nil(t, err)
}

func TestCloseSucceed(t *testing.T) {
	err := R.table.Close()
	assert.Nil(t, err)
	err = R.docdb.Close()
	assert.Nil(t, err)
}
