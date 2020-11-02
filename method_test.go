package bitxid

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/meshplus/bitxhub-kit/storage/leveldb"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/sha3"
)

var mdbPath string
var mrtPath string
var mr *MethodRegistry

var rootMethod = DID("did:bitxhub:relayroot:.")
var method DID = DID("did:bitxhub:appchain001:.")
var mcaller DID = DID("did:bitxhub:relayroot:0x12345678")

var mdoc MethodDoc = getMethodDoc(0)
var mdocA MethodDoc = getMethodDoc(1)
var mdocB MethodDoc = getMethodDoc(2)

func getMethodDoc(ran int) MethodDoc {
	docE := MethodDoc{}
	docE.ID = method
	docE.Type = "method"
	pk1 := PubKey{
		ID:           "KEY#1",
		Type:         "Ed25519",
		PublicKeyPem: "H3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPV",
	}
	pk2 := PubKey{
		ID:           "KEY#1",
		Type:         "Secp256k1",
		PublicKeyPem: "02b97c30de767f084ce3080168ee293053ba33b235d7116a3263d29f1450936b71",
	}

	if ran == 0 {
		docE.ID = rootMethod
	} else if ran == 1 {
		docE.PublicKey = []PubKey{pk1}
	} else {
		docE.PublicKey = []PubKey{pk2}
	}

	docE.Controller = DID("did:bitxhub:relayroot:0x12345678")
	auth := Auth{
		PublicKey: []string{"KEY#1"},
	}
	docE.Authentication = []Auth{auth}
	return docE
}

func TestMethodNew(t *testing.T) {
	dir1, err := ioutil.TempDir("testdata", "method.docdb")
	assert.Nil(t, err)
	dir2, err := ioutil.TempDir("testdata", "method.table")
	mdbPath = dir1
	mrtPath = dir2
	assert.Nil(t, err)
	loggerInit()
	l := loggerGet(loggerMethod)
	s1, err := leveldb.New(mrtPath)
	assert.Nil(t, err)
	s2, err := leveldb.New(mdbPath)
	assert.Nil(t, err)
	c, err := DefaultBitXIDConfig()
	assert.Nil(t, err)
	mr, err = NewMethodRegistry(s1, s2, l, &c.MethodConfig)
	assert.Nil(t, err)
}

func TestMethodSetupGenesisSucceed(t *testing.T) {
	// docBytes, err := Struct2Bytes(mdoc)
	// assert.Nil(t, err)
	// docHashE := sha3.Sum512(docBytes)
	// strHashE := fmt.Sprintf("%x", docHashE)
	// docAddrE := "./" + string(rootMethod)

	err := mr.SetupGenesis()
	assert.Nil(t, err)
	// strHash := fmt.Sprintf("%x", docHash)
	// assert.Equal(t, strHashE, strHash)
	// assert.Equal(t, docAddrE, docAddr)
}

func TestHasMethodSucceed(t *testing.T) {
	ret1, err := mr.HasMethod(DID(mr.config.GenesisMetohd))
	assert.Nil(t, err)
	assert.Equal(t, true, ret1)
}

func TestMethodApplySucceed(t *testing.T) {
	err := mr.Apply(mcaller, method) // sig
	assert.Nil(t, err)
}

func TestMethodAuditApplyFailed(t *testing.T) {
	err := mr.AuditApply(method, false)
	assert.Nil(t, err)
}
func TestMethodAuditApplySucceed(t *testing.T) {
	err := mr.AuditApply(method, true)
	assert.Nil(t, err)
}

func TestMethodRegisterSucceed(t *testing.T) {
	docBytes, err := Struct2Bytes(mdocA)
	assert.Nil(t, err)
	docHashE := sha3.Sum512(docBytes)
	strHashE := fmt.Sprintf("%x", docHashE)
	docAddrE := "./" + string(method)

	docAddr, docHash, err := mr.Register(mdocA)
	assert.Nil(t, err)
	strHash := fmt.Sprintf("%x", docHash)
	item, _, err := mr.Resolve(method)
	assert.Nil(t, err)
	assert.Equal(t, strHashE, strHash)
	assert.Equal(t, docAddrE, docAddr)
	assert.Equal(t, mcaller, item.Owner)
}

func TestMethodUpdateSucceed(t *testing.T) {
	docBytes, err := Struct2Bytes(mdocB)
	assert.Nil(t, err)
	docHashE := sha3.Sum512(docBytes)
	strHashE := fmt.Sprintf("%x", docHashE)
	docAddrE := "./" + string(method)

	docAddr, docHash, err := mr.Update(mdocB)
	assert.Nil(t, err)
	strHash := fmt.Sprintf("%x", docHash)
	assert.Equal(t, strHashE, strHash)
	assert.Equal(t, docAddrE, docAddr)
}

func TestMethodAuditRegisterSucceed(t *testing.T) {
	err := mr.Audit(method, RegisterSuccess)
	assert.Nil(t, err)
	item, _, err := mr.Resolve(method)
	assert.Nil(t, err)
	assert.Equal(t, RegisterSuccess, item.Status)
}

func TestMethodAuditUpdateSucceed(t *testing.T) {
	err := mr.Audit(method, UpdateSuccess)
	assert.Nil(t, err)
	item, _, err := mr.Resolve(method)
	assert.Nil(t, err)
	assert.Equal(t, UpdateSuccess, item.Status)
}

func TestMethodFreezeSucceed(t *testing.T) {
	err := mr.Freeze(method)
	assert.Nil(t, err)
	item, _, err := mr.Resolve(method)
	assert.Nil(t, err)
	assert.Equal(t, Frozen, item.Status)
}

func TestMethodUnFreezeSucceed(t *testing.T) {
	err := mr.UnFreeze(method)
	assert.Nil(t, err)
	item, _, err := mr.Resolve(method)
	assert.Nil(t, err)
	assert.Equal(t, Normal, item.Status)
}

func TestMethodResolveSucceed(t *testing.T) {
	docAddrE := "./" + string(method)
	item, doc, err := mr.Resolve(method)
	assert.Nil(t, err)
	assert.Equal(t, mdocB, doc) // compare doc
	itemE := MethodItem{
		Method:  DID(method),
		Owner:   mcaller,
		DocAddr: docAddrE,
		Status:  Normal,
	}
	assert.Equal(t, itemE.Method, item.Method)
	assert.Equal(t, itemE.Owner, item.Owner)
	assert.Equal(t, itemE.DocAddr, item.DocAddr)
	assert.Equal(t, itemE.Status, item.Status)
}

func TestGetMethodStatus(t *testing.T) {

}

func TestMethodAuditStatus(t *testing.T) {

}

func TestMethodDeleteSucceed(t *testing.T) {
	err := mr.Delete(method)
	assert.Nil(t, err)
	err = mr.Delete(rootMethod)
	assert.Nil(t, err)
}

func TestCloseSucceed(t *testing.T) {
	err := mr.table.Close()
	assert.Nil(t, err)
	err = mr.docdb.Close()
	assert.Nil(t, err)

	fmt.Println("mdbPath:", mdbPath)
	err = os.RemoveAll(mdbPath)
	assert.Nil(t, err)
	err = os.RemoveAll(mrtPath)
	assert.Nil(t, err)
}
