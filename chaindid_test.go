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

var superAdmin = DID("did:bitxhub:relayroot:superadmin")
var admin = DID("did:bitxhub:relayroot:admin")

var rootChainDID = DID("did:bitxhub:relayroot:.")
var chainDID DID = DID("did:bitxhub:appchain001:.")
var mcaller DID = DID("did:bitxhub:relayroot:0x12345678")

var mdoc ChainDoc = getChainDoc(0)
var mdocA ChainDoc = getChainDoc(1)
var mdocB ChainDoc = getChainDoc(2)

func getChainDoc(ran int) ChainDoc {
	docE := ChainDoc{}
	docE.ID = chainDID
	docE.Type = int(ChainDIDType)
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
	switch ran {
	case 0:
		docE.ID = rootChainDID
	case 1:
		docE.PublicKey = []PubKey{pk1}
	case 2:
		docE.PublicKey = []PubKey{pk2}
	}

	docE.Controller = DID("did:bitxhub:relayroot:0x12345678")
	auth := Auth{
		PublicKey: []string{"KEY#1"},
	}
	docE.Authentication = []Auth{auth}
	return docE
}

func newChainDIDModeInternal(t *testing.T) (*ChainDIDRegistry, string, string) {
	dir1, err := ioutil.TempDir("", "chainDID.table")
	assert.Nil(t, err)
	dir2, err := ioutil.TempDir("", "chainDID.docdb")
	assert.Nil(t, err)

	mrtPath := dir1
	mdbPath := dir2
	loggerInit()
	l := loggerGet(loggerChainDID)
	s1, err := leveldb.New(mrtPath)
	assert.Nil(t, err)
	s2, err := leveldb.New(mdbPath)
	assert.Nil(t, err)
	mr, err := NewChainDIDRegistry(s1, l,
		WithGenesisChainDocContent(&mdoc),
		WithAdmin(superAdmin),
		WithChainDocStorage(s2))
	assert.Nil(t, err)
	return mr, mrtPath, mdbPath
}

func newChainDIDModeExternal(t *testing.T) (*ChainDIDRegistry, string) {
	dir1, err := ioutil.TempDir("", "chainDID.table")
	assert.Nil(t, err)
	mrtPath := dir1

	loggerInit()
	l := loggerGet(loggerChainDID)
	s1, err := leveldb.New(mrtPath)
	assert.Nil(t, err)
	mr, err := NewChainDIDRegistry(s1, l,
		WithGenesisChainDocInfo(DocInfo{rootChainDID, "/addr/to/doc", []byte{1}}),
		WithAdmin(superAdmin))
	assert.Nil(t, err)
	return mr, mrtPath
}

func TestChainDIDModeInternal(t *testing.T) {
	mr, drtPath, ddbPath := newChainDIDModeInternal(t)

	testChainDIDSetupGenesSucceed(t, mr)
	testHasChainDIDSucceed(t, mr)
	testChainDIDAddAdminsSucceed(t, mr)
	testChainDIDRemoveAdminsSucceed(t, mr)

	testChainDIDApplySucceed(t, mr)
	testChainDIDAuditApplyFailed(t, mr)
	testChainDIDAuditApplySucceed(t, mr)

	testChainDIDRegisterSucceedInternal(t, mr)
	testChainDIDUpdateSucceedInternal(t, mr)

	testChainDIDAuditUpdateSucceed(t, mr)
	testChainDIDAuditStatusNormal(t, mr)
	testChainDIDFreezeSucceed(t, mr)
	testChainDIDUnFreezeSucceed(t, mr)
	testChainDIDResolveSucceedInternal(t, mr)
	testChainDIDDeleteSucceed(t, mr)
	testCloseSucceedInternal(t, mr, drtPath, ddbPath)
}

func TestChainDIDModeExternal(t *testing.T) {
	mr, drtPath := newChainDIDModeExternal(t)

	testChainDIDSetupGenesSucceed(t, mr)
	testHasChainDIDSucceed(t, mr)
	testChainDIDAddAdminsSucceed(t, mr)
	testChainDIDRemoveAdminsSucceed(t, mr)

	testChainDIDApplySucceed(t, mr)
	testChainDIDAuditApplyFailed(t, mr)
	testChainDIDAuditApplySucceed(t, mr)

	testChainDIDRegisterSucceedExternal(t, mr)
	testChainDIDUpdateSucceedExternal(t, mr)

	testChainDIDAuditUpdateSucceed(t, mr)
	testChainDIDAuditStatusNormal(t, mr)
	testChainDIDFreezeSucceed(t, mr)
	testChainDIDUnFreezeSucceed(t, mr)

	testChainDIDResolveSucceedExternal(t, mr)
	testChainDIDDeleteSucceed(t, mr)

	testCloseSucceedExternal(t, mr, drtPath)
}

func testChainDIDSetupGenesSucceed(t *testing.T, mr *ChainDIDRegistry) {
	err := mr.SetupGenesis()
	assert.Nil(t, err)
}

func testHasChainDIDSucceed(t *testing.T, mr *ChainDIDRegistry) {
	ret1 := mr.HasChainDID(mr.GenesisChainDID)
	assert.Equal(t, true, ret1)
}

func testChainDIDAddAdminsSucceed(t *testing.T, mr *ChainDIDRegistry) {
	err := mr.AddAdmin(admin)
	assert.Nil(t, err)
	ret := mr.HasAdmin(admin)
	assert.Equal(t, true, ret)
}

func testChainDIDRemoveAdminsSucceed(t *testing.T, mr *ChainDIDRegistry) {
	err := mr.RemoveAdmin(admin)
	assert.Nil(t, err)
	ret := mr.HasAdmin(admin)
	assert.Equal(t, false, ret)
}

func testChainDIDApplySucceed(t *testing.T, mr *ChainDIDRegistry) {
	err := mr.Apply(mcaller, chainDID) // sig
	assert.Nil(t, err)
}

func testChainDIDAuditApplyFailed(t *testing.T, mr *ChainDIDRegistry) {
	err := mr.AuditApply(chainDID, false)
	assert.Nil(t, err)
}

func testChainDIDAuditApplySucceed(t *testing.T, mr *ChainDIDRegistry) {
	err := mr.AuditApply(chainDID, true)
	assert.Nil(t, err)
}

func testChainDIDRegisterSucceedInternal(t *testing.T, mr *ChainDIDRegistry) {
	docBytes, err := Marshal(mdocA)
	assert.Nil(t, err)
	docHashE := sha256.Sum256(docBytes)

	strHashE := fmt.Sprintf("%x", docHashE)
	docAddrE := "./" + string(chainDID)

	docAddr, docHash, err := mr.RegisterWithDoc(&mdocA)
	assert.Nil(t, err)
	strHash := fmt.Sprintf("%x", docHash)
	item, _, _, err := mr.Resolve(chainDID)
	assert.Nil(t, err)
	assert.Equal(t, strHashE, strHash)
	assert.Equal(t, docAddrE, docAddr)
	assert.Equal(t, mcaller, item.Owner)
}

func testChainDIDRegisterSucceedExternal(t *testing.T, mr *ChainDIDRegistry) {
	docBytes, err := Marshal(mdocA)
	assert.Nil(t, err)
	docHashE := sha256.Sum256(docBytes)
	docAddrE := "./addr/" + string(chainDID)
	docAddr, docHash, err := mr.Register(chainDID, docAddrE, docHashE[:])
	assert.Nil(t, err)

	strHashE := fmt.Sprintf("%x", docHashE)
	strHash := fmt.Sprintf("%x", docHash)

	item, _, _, err := mr.Resolve(chainDID)
	assert.Nil(t, err)
	assert.Equal(t, strHashE, strHash)
	assert.Equal(t, docAddrE, docAddr)
	assert.Equal(t, mcaller, item.Owner)
}

func testChainDIDUpdateSucceedInternal(t *testing.T, mr *ChainDIDRegistry) {
	docBytes, err := Marshal(mdocB)
	assert.Nil(t, err)
	docHashE := sha256.Sum256(docBytes)
	strHashE := fmt.Sprintf("%x", docHashE)
	docAddrE := "./" + string(chainDID)

	docAddr, docHash, err := mr.UpdateWithDoc(&mdocB)
	assert.Nil(t, err)
	strHash := fmt.Sprintf("%x", docHash)
	assert.Equal(t, strHashE, strHash)
	assert.Equal(t, docAddrE, docAddr)
}

func testChainDIDUpdateSucceedExternal(t *testing.T, mr *ChainDIDRegistry) {
	docBytes, err := Marshal(mdocB)
	assert.Nil(t, err)
	docHashE := sha256.Sum256(docBytes)
	docAddrE := "/addr/" + string(chainDID)

	_, _, err = mr.Update(chainDID, docAddrE, docHashE[:])
	assert.Nil(t, err)
}

func testChainDIDAuditUpdateSucceed(t *testing.T, mr *ChainDIDRegistry) {
	err := mr.Audit(chainDID, RegisterFailed)
	assert.Nil(t, err)
	item, _, _, err := mr.Resolve(chainDID)
	assert.Nil(t, err)
	assert.Equal(t, RegisterFailed, item.Status)
}

func testChainDIDAuditStatusNormal(t *testing.T, mr *ChainDIDRegistry) {
	err := mr.Audit(chainDID, Normal)
	assert.Nil(t, err)
	item, _, _, err := mr.Resolve(chainDID)
	assert.Nil(t, err)
	assert.Equal(t, Normal, item.Status)
}

func testChainDIDFreezeSucceed(t *testing.T, mr *ChainDIDRegistry) {
	err := mr.Freeze(chainDID)
	assert.Nil(t, err)
	item, _, _, err := mr.Resolve(chainDID)
	assert.Nil(t, err)
	assert.Equal(t, Frozen, item.Status)
}

func testChainDIDUnFreezeSucceed(t *testing.T, mr *ChainDIDRegistry) {
	err := mr.UnFreeze(chainDID)
	assert.Nil(t, err)
	item, _, _, err := mr.Resolve(chainDID)
	assert.Nil(t, err)
	assert.Equal(t, Normal, item.Status)
}

func testChainDIDResolveSucceedInternal(t *testing.T, mr *ChainDIDRegistry) {
	docAddrE := "./" + string(chainDID)
	item, doc, _, err := mr.Resolve(chainDID)
	assert.Nil(t, err)
	assert.Equal(t, &mdocB, doc) // compare doc
	itemE := ChainItem{
		BasicItem{ID: chainDID,
			DocAddr: docAddrE,
			Status:  Normal},
		mcaller,
	}
	assert.Equal(t, itemE.ID, item.ID)
	assert.Equal(t, itemE.Owner, item.Owner)
	assert.Equal(t, itemE.DocAddr, item.DocAddr)
	assert.Equal(t, itemE.Status, item.Status)
}

func testChainDIDResolveSucceedExternal(t *testing.T, mr *ChainDIDRegistry) {
	docAddrE := "/addr/" + string(chainDID)
	item, doc, _, err := mr.Resolve(chainDID)
	assert.Nil(t, err)
	assert.Nil(t, doc)
	itemE := ChainItem{
		BasicItem: BasicItem{ID: chainDID,
			DocAddr: docAddrE,
			Status:  Normal},
	}
	assert.Equal(t, itemE.ID, item.ID)
	assert.Equal(t, itemE.DocAddr, item.DocAddr)
	assert.Equal(t, itemE.Status, item.Status)
}

func testChainDIDDeleteSucceed(t *testing.T, mr *ChainDIDRegistry) {
	err := mr.Delete(chainDID)
	assert.Nil(t, err)
	err = mr.Delete(rootChainDID)
	assert.Nil(t, err)

	item, _, _, err := mr.Resolve(chainDID)
	assert.Nil(t, err)
	assert.Nil(t, item)

	item2, _, _, err := mr.Resolve(rootChainDID)
	assert.Nil(t, err)
	assert.Nil(t, item2)
}

func testCloseSucceedInternal(t *testing.T, mr *ChainDIDRegistry, path ...string) {
	err := mr.Table.Close()
	assert.Nil(t, err)
	err = mr.Docdb.Close()
	assert.Nil(t, err)

	for _, p := range path {
		err = os.RemoveAll(p)
		assert.Nil(t, err)
	}
}

func testCloseSucceedExternal(t *testing.T, mr *ChainDIDRegistry, path ...string) {
	err := mr.Table.Close()
	assert.Nil(t, err)

	for _, p := range path {
		err = os.RemoveAll(p)
		assert.Nil(t, err)
	}
}
