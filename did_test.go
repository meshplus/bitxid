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

// var drtPath string
// var ddbPath string
// var r *DIDRegistry

var rootDID DID = DID("did:bitxhub:appchain001:0x00000001")
var did DID = DID("did:bitxhub:appchain001:0x12345678")

var diddoc DIDDoc = getDIDDoc(0)
var diddocA DIDDoc = getDIDDoc(1)
var diddocB DIDDoc = getDIDDoc(2)

func getDIDDoc(ran int) DIDDoc {
	docE := DIDDoc{}
	docE.ID = did
	docE.Type = "user"
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
		docE.ID = rootDID
	} else if ran == 1 {
		docE.PublicKey = []PubKey{pk1}
	} else {
		docE.PublicKey = []PubKey{pk2}
	}
	auth := Auth{
		PublicKey: []string{"KEY#1"},
	}
	docE.Authentication = []Auth{auth}
	return docE
}

func TestDIDMode_Internal(t *testing.T) {
	r, drtPath, ddbPath := newDIDModeInternal(t)

	testSetupDIDSucceed(t, r)
	testHasDIDSucceed(t, r)

	testDIDRegisterSucceedInternal(t, r)
	testDIDUpdateSucceedInternal(t, r)
	testDIDResolveSucceedInternal(t, r)

	testDIDFreezeSucceed(t, r)
	testDIDUnFreezeSucceed(t, r)
	testDIDDeleteSucceed(t, r)
	testDIDCloseSucceedInternal(t, r, drtPath, ddbPath)
}

func TestDIDMode_External(t *testing.T) {
	r, drtPath := newDIDModeExternal(t)
	testSetupDIDSucceed(t, r)
	testHasDIDSucceed(t, r)

	testDIDRegisterSucceedExternal(t, r)
	testDIDUpdateSucceedExternal(t, r)
	testDIDResolveSucceedExternal(t, r)

	testDIDFreezeSucceed(t, r)
	testDIDUnFreezeSucceed(t, r)
	testDIDDeleteSucceed(t, r)
	testDIDCloseSucceedExternal(t, r, drtPath)
}

func newDIDModeInternal(t *testing.T) (*DIDRegistry, string, string) {
	dir1, err := ioutil.TempDir("testdata", "did.table")
	assert.Nil(t, err)
	dir2, err := ioutil.TempDir("testdata", "did.docdb")
	assert.Nil(t, err)

	drtPath := dir1
	ddbPath := dir2
	loggerInit()
	l := loggerGet(loggerDID)
	s1, err := leveldb.New(drtPath)
	assert.Nil(t, err)
	s2, err := leveldb.New(ddbPath)
	assert.Nil(t, err)
	r, err := NewDIDRegistry(s1, l, WithDIDDocStorage(s2))
	assert.Nil(t, err)
	return r, drtPath, ddbPath
}

func newDIDModeExternal(t *testing.T) (*DIDRegistry, string) {
	dir1, err := ioutil.TempDir("testdata", "did.table")
	assert.Nil(t, err)
	drtPath := dir1
	loggerInit()
	l := loggerGet(loggerDID)
	s1, err := leveldb.New(drtPath)
	assert.Nil(t, err)
	r, err := NewDIDRegistry(s1, l)
	assert.Nil(t, err)
	return r, drtPath
}

func testSetupDIDSucceed(t *testing.T, r *DIDRegistry) {
	err := r.SetupGenesis()
	assert.Nil(t, err)
}

func testHasDIDSucceed(t *testing.T, r *DIDRegistry) {
	ret1 := r.HasDID(DID(r.genesisDID))
	assert.Equal(t, true, ret1)
}

func testDIDRegisterSucceedInternal(t *testing.T, r *DIDRegistry) {
	docABytes, err := Struct2Bytes(diddocA)
	assert.Nil(t, err)
	docHashE := sha256.Sum256(docABytes)
	docAddrE := "./" + string(did)
	docAddr, docHash, err := r.Register(DocOption{Content: &diddocA})
	assert.Nil(t, err)
	strHash := fmt.Sprintf("%x", docHash)
	strHashE := fmt.Sprintf("%x", docHashE)
	assert.Equal(t, strHashE, strHash)
	assert.Equal(t, docAddrE, docAddr)
}

func testDIDRegisterSucceedExternal(t *testing.T, r *DIDRegistry) {
	docABytes, err := Struct2Bytes(diddocA)
	assert.Nil(t, err)
	docHashE := sha256.Sum256(docABytes)
	docAddrE := "./addr/" + string(did)
	_, _, err = r.Register(DocOption{ID: did, Addr: docAddrE, Hash: docHashE[:]})
	assert.Nil(t, err)
}

func testDIDUpdateSucceedInternal(t *testing.T, r *DIDRegistry) {
	docBBytes, err := Struct2Bytes(diddocB)
	assert.Nil(t, err)
	docHashE := sha256.Sum256(docBBytes)
	docAddrE := "./" + string(did)
	docAddr, docHash, err := r.Update(DocOption{Content: &diddocB})
	assert.Nil(t, err)
	strHash := fmt.Sprintf("%x", docHash)
	strHashE := fmt.Sprintf("%x", docHashE)
	assert.Equal(t, strHashE, strHash)
	assert.Equal(t, docAddrE, docAddr)
}

func testDIDUpdateSucceedExternal(t *testing.T, r *DIDRegistry) {
	docBBytes, err := Struct2Bytes(diddocB)
	assert.Nil(t, err)
	docHashE := sha256.Sum256(docBBytes)
	docAddrE := "/addr/" + string(did)
	_, _, err = r.Update(DocOption{ID: did, Addr: docAddrE, Hash: docHashE[:]})
	assert.Nil(t, err)
}

func testDIDResolveSucceedInternal(t *testing.T, r *DIDRegistry) {
	item, doc, _, err := r.Resolve(did)
	docBBytes, err := Struct2Bytes(diddocB)
	assert.Nil(t, err)
	docHashE := sha256.Sum256(docBBytes)
	assert.Nil(t, err)
	assert.Equal(t, &diddocB, doc) // compare doc
	itemE := DIDItem{
		BasicItem{
			ID:      did,
			DocHash: docHashE[:],
			DocAddr: "./" + string(did),
			Status:  Normal},
	}
	assert.Equal(t, itemE.ID, item.ID)
	assert.Equal(t, itemE.DocAddr, item.DocAddr)
	assert.Equal(t, itemE.DocHash, item.DocHash)
	assert.Equal(t, itemE.Status, item.Status)
}

func testDIDResolveSucceedExternal(t *testing.T, r *DIDRegistry) {
	item, doc, _, err := r.Resolve(did)
	assert.Nil(t, err)
	assert.Nil(t, doc)
	docBBytes, err := Struct2Bytes(diddocB)
	docHashE := sha256.Sum256(docBBytes)
	assert.Nil(t, err)
	itemE := DIDItem{
		BasicItem{
			ID:      did,
			DocHash: docHashE[:],
			DocAddr: "/addr/" + string(did),
			Status:  Normal},
	}
	assert.Equal(t, itemE.ID, item.ID)
	assert.Equal(t, itemE.DocAddr, item.DocAddr)
	assert.Equal(t, itemE.DocHash, item.DocHash)
	assert.Equal(t, itemE.Status, item.Status)
}

func testDIDFreezeSucceed(t *testing.T, r *DIDRegistry) {
	err := r.Freeze(did)
	assert.Nil(t, err)
	item, _, _, err := r.Resolve(did)
	assert.Nil(t, err)
	assert.Equal(t, Frozen, item.Status)
}

func testDIDUnFreezeSucceed(t *testing.T, r *DIDRegistry) {
	err := r.UnFreeze(did)
	assert.Nil(t, err)
	item, _, _, err := r.Resolve(did)
	assert.Nil(t, err)
	assert.Equal(t, Normal, item.Status)
}

func testDIDDeleteSucceed(t *testing.T, r *DIDRegistry) {
	err := r.Delete(did)
	assert.Nil(t, err)
	err = r.Delete(rootDID)
	assert.Nil(t, err)
}

func testDIDCloseSucceedInternal(t *testing.T, r *DIDRegistry, path ...string) {
	err := r.table.Close()
	assert.Nil(t, err)
	err = r.docdb.Close()
	assert.Nil(t, err)

	for _, p := range path {
		err = os.RemoveAll(p)
		assert.Nil(t, err)
	}
}

func testDIDCloseSucceedExternal(t *testing.T, r *DIDRegistry, path ...string) {
	err := r.table.Close()
	assert.Nil(t, err)

	for _, p := range path {
		err = os.RemoveAll(p)
		assert.Nil(t, err)
	}
}
