package method

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/meshplus/bitxhub-kit/storage/leveldb"
	"github.com/bitxhub/bitxid/internal/common/types"
	"github.com/bitxhub/bitxid/internal/loggers"
	"github.com/bitxhub/bitxid/internal/repo"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/sha3"
)

const (
	dbPath   string = "../../config/doc.db"
	rtPath   string = "../../config/registry.store"
	confPath string = "../../config"
)

var R *Registry

var docB []byte = []byte("{\"MethodName\":\"did:bitxhub:relayroot:.\",\"Auth\": {}}")
var docC []byte = []byte("{\"MethodName\":\"did:bitxhub:appchain001:.\",\"Auth\": {}}")

var docd string = `{	"id":"did:bitxhub:relayroot:0x12345678",
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
var docD []byte = []byte(docd)

var caller types.DID = types.DID("did:bitxhub:relayroot:0x12345678")

var method string = "did:bitxhub:appchain001:."

var sig []byte = []byte("0x9AB567")

func TestNew(t *testing.T) {
	loggers.Init()
	L := loggers.Get(loggers.Method)
	S1, err := leveldb.New(rtPath)
	assert.Nil(t, err)
	S2, err := leveldb.New(dbPath)
	assert.Nil(t, err)
	C, err := repo.UnmarshalConfig(confPath)
	assert.Nil(t, err)
	R, err = New(S1, S2, L, &C.MethodConfig)
	assert.Nil(t, err)
}

func TestSetupGenesisSucceed(t *testing.T) {
	docHashE := sha3.Sum512(docB)
	strHashE := fmt.Sprintf("%x", docHashE)
	docAddrE := "./did:bitxhub:relayroot:."
	docAddr, docHash, err := R.SetupGenesis(docB)
	assert.Nil(t, err)
	strHash := fmt.Sprintf("%x", docHash)
	assert.Equal(t, strHashE, strHash)
	assert.Equal(t, docAddrE, docAddr)
}

func TestHasMethodSucceed(t *testing.T) {
	ret1, err := R.HasMethod(R.config.GenesisMetohd)
	assert.Nil(t, err)
	assert.Equal(t, true, ret1)
}

func TestApplySucceed(t *testing.T) {
	err := R.Apply(caller, method, sig)
	assert.Nil(t, err)
}

func TestAuditApplyFailed(t *testing.T) {
	err := R.AuditApply(caller, method, false, sig)
	assert.Nil(t, err)
}
func TestAuditApplySucceed(t *testing.T) {
	err := R.AuditApply(caller, method, true, sig)
	assert.Nil(t, err)
}

func TestRegisterSucceed(t *testing.T) {
	docHashE := sha3.Sum512(docC)
	strHashE := fmt.Sprintf("%x", docHashE)
	docAddrE := "./did:bitxhub:appchain001:."
	docAddr, docHash, err := R.Register(caller, method, docC, sig)
	assert.Nil(t, err)
	strHash := fmt.Sprintf("%x", docHash)
	item, _, err := R.Resolve(caller, method, sig)
	assert.Nil(t, err)
	assert.Equal(t, strHashE, strHash)
	assert.Equal(t, docAddrE, docAddr)
	assert.Equal(t, caller, item.Owner)
}

func TestUpdateSucceed(t *testing.T) {
	docHashE := sha3.Sum512(docD)
	strHashE := fmt.Sprintf("%x", docHashE)
	docAddrE := "./did:bitxhub:appchain001:."
	docAddr, docHash, err := R.Update(caller, method, docD, sig)
	assert.Nil(t, err)
	strHash := fmt.Sprintf("%x", docHash)
	assert.Equal(t, strHashE, strHash)
	assert.Equal(t, docAddrE, docAddr)
}

func TestAuditRegisterSucceed(t *testing.T) {
	err := R.Audit(caller, method, RegisterSuccess, sig)
	assert.Nil(t, err)
	item, _, err := R.Resolve(caller, method, sig)
	assert.Nil(t, err)
	assert.Equal(t, RegisterSuccess, item.Status)
}

func TestAuditUpdateSucceed(t *testing.T) {
	err := R.Audit(caller, method, UpdateSuccess, sig)
	assert.Nil(t, err)
	item, _, err := R.Resolve(caller, method, sig)
	assert.Nil(t, err)
	assert.Equal(t, UpdateSuccess, item.Status)
}

func TestFreezeSucceed(t *testing.T) {
	err := R.Freeze(caller, method, sig)
	assert.Nil(t, err)
	item, _, err := R.Resolve(caller, method, sig)
	assert.Nil(t, err)
	assert.Equal(t, Frozen, item.Status)
}

func TestUnFreezeSucceed(t *testing.T) {
	err := R.UnFreeze(caller, method, sig)
	assert.Nil(t, err)
	item, _, err := R.Resolve(caller, method, sig)
	assert.Nil(t, err)
	assert.Equal(t, Normal, item.Status)
}

func TestResolveSucceed(t *testing.T) {
	docAddrE := "./did:bitxhub:appchain001:."
	item, doc, err := R.Resolve(caller, method, sig)
	assert.Nil(t, err)
	assert.Equal(t, docD, doc)
	itemE := Item{
		Key:     method,
		Owner:   caller,
		DocAddr: docAddrE,
		Status:  Normal,
	}
	assert.Equal(t, itemE.Key, item.Key)
	assert.Equal(t, itemE.Owner, item.Owner)
	assert.Equal(t, itemE.DocAddr, item.DocAddr)
	assert.Equal(t, itemE.Status, item.Status)
}

func TestUnMarshalSucceed(t *testing.T) {
	_, doc, err := R.Resolve(caller, method, sig)
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

func TestOwnsSucceed(t *testing.T) {
	res := R.owns(caller, method)
	assert.Equal(t, true, res)
}

func TestGetMethodStatus(t *testing.T) {

}

func TestAuditStatus(t *testing.T) {

}

func TestDeleteSucceed(t *testing.T) {
	err := R.Delete(caller, method, sig)
	assert.Nil(t, err)
	err = R.Delete("did:bitxhub:relayroot:0x01", "did:bitxhub:relayroot:.", sig)
	assert.Nil(t, err)
}

func TestCloseSucceed(t *testing.T) {
	err := R.table.Close()
	assert.Nil(t, err)
	err = R.docdb.Close()
	assert.Nil(t, err)
}
