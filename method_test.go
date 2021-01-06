package bitxid

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/meshplus/bitxhub-kit/storage/leveldb"
	"github.com/stretchr/testify/assert"
)

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

func newMethodModeInternal(t *testing.T) (*MethodRegistry, string, string) {
	dir1, err := ioutil.TempDir("testdata", "method.table")
	assert.Nil(t, err)
	dir2, err := ioutil.TempDir("testdata", "method.docdb")
	assert.Nil(t, err)

	mrtPath := dir1
	mdbPath := dir2
	loggerInit()
	l := loggerGet(loggerMethod)
	s1, err := leveldb.New(mrtPath)
	assert.Nil(t, err)
	s2, err := leveldb.New(mdbPath)
	assert.Nil(t, err)
	mr, err := NewMethodRegistry(s1, l, WithMethodAdmin(DID("")), WithMethodDocStorage(s2))
	assert.Nil(t, err)
	return mr, mrtPath, mdbPath
}

func newMethodModeExternal(t *testing.T) (*MethodRegistry, string) {
	dir1, err := ioutil.TempDir("testdata", "method.table")
	assert.Nil(t, err)
	mrtPath := dir1

	loggerInit()
	l := loggerGet(loggerMethod)
	s1, err := leveldb.New(mrtPath)
	assert.Nil(t, err)
	mr, err := NewMethodRegistry(s1, l, WithMethodAdmin(DID("")))
	assert.Nil(t, err)
	return mr, mrtPath
}

func TestMethodMode_Internal(t *testing.T) {
	mr, drtPath, ddbPath := newMethodModeInternal(t)

	testMethodSetupGenesSucceed(t, mr)
	testHasMethodSucceed(t, mr)
	testMethodApplySucceed(t, mr)
	testMethodAuditApplyFailed(t, mr)
	testMethodAuditApplySucceed(t, mr)

	testMethodRegisterSucceedInternal(t, mr)
	testMethodUpdateSucceedInternal(t, mr)

	testMethodAuditUpdateSucceed(t, mr)
	testMethodAuditStatusNormal(t, mr)
	testMethodFreezeSucceed(t, mr)
	testMethodUnFreezeSucceed(t, mr)
	testMethodResolveSucceedInternal(t, mr)
	testMethodDeleteSucceed(t, mr)
	testCloseSucceedInternal(t, mr, drtPath, ddbPath)
}

func TestMethodMode_External(t *testing.T) {
	mr, drtPath := newMethodModeExternal(t)

	testMethodSetupGenesSucceed(t, mr)
	testHasMethodSucceed(t, mr)
	testMethodApplySucceed(t, mr)
	testMethodAuditApplyFailed(t, mr)
	testMethodAuditApplySucceed(t, mr)

	testMethodRegisterSucceedExternal(t, mr)
	testMethodUpdateSucceedExternal(t, mr)

	testMethodAuditUpdateSucceed(t, mr)
	testMethodAuditStatusNormal(t, mr)
	testMethodFreezeSucceed(t, mr)
	testMethodUnFreezeSucceed(t, mr)

	testMethodResolveSucceedExternal(t, mr)
	testMethodDeleteSucceed(t, mr)

	testCloseSucceedExternal(t, mr, drtPath)
}

func testMethodSetupGenesSucceed(t *testing.T, mr *MethodRegistry) {
	err := mr.SetupGenesis()
	assert.Nil(t, err)
}

func testHasMethodSucceed(t *testing.T, mr *MethodRegistry) {
	ret1 := mr.HasMethod(DID(mr.genesisMetohd))
	assert.Equal(t, true, ret1)
}

func testMethodApplySucceed(t *testing.T, mr *MethodRegistry) {
	err := mr.Apply(mcaller, method) // sig
	assert.Nil(t, err)
}

func testMethodAuditApplyFailed(t *testing.T, mr *MethodRegistry) {
	err := mr.AuditApply(method, false)
	assert.Nil(t, err)
}

func testMethodAuditApplySucceed(t *testing.T, mr *MethodRegistry) {
	err := mr.AuditApply(method, true)
	assert.Nil(t, err)
}

func testMethodRegisterSucceedInternal(t *testing.T, mr *MethodRegistry) {
	docBytes, err := Struct2Bytes(mdocA)
	assert.Nil(t, err)
	docHashE := sha256.Sum256(docBytes)

	strHashE := fmt.Sprintf("%x", docHashE)
	docAddrE := "./" + string(method)

	docAddr, docHash, err := mr.Register(DocOption{Content: &mdocA})
	assert.Nil(t, err)
	strHash := fmt.Sprintf("%x", docHash)
	item, _, _, err := mr.Resolve(method)
	assert.Nil(t, err)
	assert.Equal(t, strHashE, strHash)
	assert.Equal(t, docAddrE, docAddr)
	assert.Equal(t, mcaller, item.Owner)
}

func testMethodRegisterSucceedExternal(t *testing.T, mr *MethodRegistry) {
	docBytes, err := Struct2Bytes(mdocA)
	assert.Nil(t, err)
	docHashE := sha256.Sum256(docBytes)
	docAddrE := "./addr/" + string(method)
	docAddr, docHash, err := mr.Register(DocOption{
		ID:   method,
		Addr: docAddrE,
		Hash: docHashE[:]})
	assert.Nil(t, err)

	strHashE := fmt.Sprintf("%x", docHashE)
	strHash := fmt.Sprintf("%x", docHash)

	item, _, _, err := mr.Resolve(method)
	assert.Nil(t, err)
	assert.Equal(t, strHashE, strHash)
	assert.Equal(t, docAddrE, docAddr)
	assert.Equal(t, mcaller, item.Owner)
}

func testMethodUpdateSucceedInternal(t *testing.T, mr *MethodRegistry) {
	docBytes, err := Struct2Bytes(mdocB)
	assert.Nil(t, err)
	docHashE := sha256.Sum256(docBytes)
	strHashE := fmt.Sprintf("%x", docHashE)
	docAddrE := "./" + string(method)

	docAddr, docHash, err := mr.Update(DocOption{Content: &mdocB})
	assert.Nil(t, err)
	strHash := fmt.Sprintf("%x", docHash)
	assert.Equal(t, strHashE, strHash)
	assert.Equal(t, docAddrE, docAddr)
}

func testMethodUpdateSucceedExternal(t *testing.T, mr *MethodRegistry) {
	docBytes, err := Struct2Bytes(mdocB)
	assert.Nil(t, err)
	docHashE := sha256.Sum256(docBytes)
	docAddrE := "/addr/" + string(method)

	_, _, err = mr.Update(DocOption{
		ID:   method,
		Addr: docAddrE,
		Hash: docHashE[:]})
	assert.Nil(t, err)
}

func testMethodAuditUpdateSucceed(t *testing.T, mr *MethodRegistry) {
	err := mr.Audit(method, RegisterFailed)
	assert.Nil(t, err)
	item, _, _, err := mr.Resolve(method)
	assert.Nil(t, err)
	assert.Equal(t, RegisterFailed, item.Status)
}

func testMethodAuditStatusNormal(t *testing.T, mr *MethodRegistry) {
	err := mr.Audit(method, Normal)
	assert.Nil(t, err)
	item, _, _, err := mr.Resolve(method)
	assert.Nil(t, err)
	assert.Equal(t, Normal, item.Status)
}

func testMethodFreezeSucceed(t *testing.T, mr *MethodRegistry) {
	err := mr.Freeze(method)
	assert.Nil(t, err)
	item, _, _, err := mr.Resolve(method)
	assert.Nil(t, err)
	assert.Equal(t, Frozen, item.Status)
}

func testMethodUnFreezeSucceed(t *testing.T, mr *MethodRegistry) {
	err := mr.UnFreeze(method)
	assert.Nil(t, err)
	item, _, _, err := mr.Resolve(method)
	assert.Nil(t, err)
	assert.Equal(t, Normal, item.Status)
}

func testMethodResolveSucceedInternal(t *testing.T, mr *MethodRegistry) {
	docAddrE := "./" + string(method)
	item, doc, _, err := mr.Resolve(method)
	assert.Nil(t, err)
	assert.Equal(t, &mdocB, doc) // compare doc
	itemE := MethodItem{
		BasicItem{ID: DID(method),
			DocAddr: docAddrE,
			Status:  Normal},
		mcaller,
	}
	assert.Equal(t, itemE.ID, item.ID)
	assert.Equal(t, itemE.Owner, item.Owner)
	assert.Equal(t, itemE.DocAddr, item.DocAddr)
	assert.Equal(t, itemE.Status, item.Status)
}

func testMethodResolveSucceedExternal(t *testing.T, mr *MethodRegistry) {
	docAddrE := "/addr/" + string(method)
	item, doc, _, err := mr.Resolve(method)
	assert.Nil(t, err)
	assert.Nil(t, doc)
	itemE := MethodItem{
		BasicItem: BasicItem{ID: DID(method),
			DocAddr: docAddrE,
			Status:  Normal},
	}
	assert.Equal(t, itemE.ID, item.ID)
	assert.Equal(t, itemE.DocAddr, item.DocAddr)
	assert.Equal(t, itemE.Status, item.Status)
}

func testMethodDeleteSucceed(t *testing.T, mr *MethodRegistry) {
	err := mr.Delete(method)
	assert.Nil(t, err)
	err = mr.Delete(rootMethod)
	assert.Nil(t, err)
}

func testCloseSucceedInternal(t *testing.T, mr *MethodRegistry, path ...string) {
	err := mr.table.Close()
	assert.Nil(t, err)
	err = mr.docdb.Close()
	assert.Nil(t, err)

	for _, p := range path {
		err = os.RemoveAll(p)
		assert.Nil(t, err)
	}
}

func testCloseSucceedExternal(t *testing.T, mr *MethodRegistry, path ...string) {
	err := mr.table.Close()
	assert.Nil(t, err)

	for _, p := range path {
		err = os.RemoveAll(p)
		assert.Nil(t, err)
	}
}
