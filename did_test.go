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

var drtPath string
var ddbPath string
var r *DIDRegistry

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

func TestDIDNew(t *testing.T) {
	dir1, err := ioutil.TempDir("testdata", "did.docdb")
	assert.Nil(t, err)
	dir2, err := ioutil.TempDir("testdata", "did.table")
	assert.Nil(t, err)
	drtPath = dir1
	ddbPath = dir2
	loggerInit()
	l := loggerGet(loggerDID)
	s1, err := leveldb.New(drtPath)
	assert.Nil(t, err)
	s2, err := leveldb.New(ddbPath)
	assert.Nil(t, err)
	r, err = NewDIDRegistry(s1, s2, l)
	assert.Nil(t, err)
}

func TestSetupDIDSucceed(t *testing.T) {
	err := r.SetupGenesis()
	assert.Nil(t, err)
}

func testHasDIDSucceed(t *testing.T) {
	ret1 := r.HasDID(DID(r.config.Admin))
	assert.Equal(t, true, ret1)
}

func TestDIDRegisterSucceed(t *testing.T) {
	docABytes, err := Struct2Bytes(diddocA)
	assert.Nil(t, err)
	// docHashE := sha3.Sum512(docABytes)
	docHashE := sha256.Sum256(docABytes)
	docAddrE := "./" + string(did)
	docAddr, docHash, err := r.Register(&diddocA)
	assert.Nil(t, err)
	strHash := fmt.Sprintf("%x", docHash)
	strHashE := fmt.Sprintf("%x", docHashE)
	assert.Equal(t, strHashE, strHash)
	assert.Equal(t, docAddrE, docAddr)
}

func TestDIDUpdateSucceed(t *testing.T) {
	docBBytes, err := Struct2Bytes(diddocB)
	assert.Nil(t, err)
	// docHashE := sha3.Sum512(docBBytes)
	docHashE := sha256.Sum256(docBBytes)
	docAddrE := "./" + string(did)
	docAddr, docHash, err := r.Update(&diddocB)
	assert.Nil(t, err)
	strHash := fmt.Sprintf("%x", docHash)
	strHashE := fmt.Sprintf("%x", docHashE)
	assert.Equal(t, strHashE, strHash)
	assert.Equal(t, docAddrE, docAddr)
}
func TestDIDResolveSucceed(t *testing.T) {
	item, doc, _, err := r.Resolve(did)
	docBBytes, err := Struct2Bytes(diddocB)
	assert.Nil(t, err)
	// docHashE := sha3.Sum512(docBBytes)
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

func TestDIDFreezeSucceed(t *testing.T) {
	err := r.Freeze(did)
	assert.Nil(t, err)
	item, _, _, err := r.Resolve(did)
	assert.Nil(t, err)
	assert.Equal(t, Frozen, item.Status)
}

func TestDIDUnFreezeSucceed(t *testing.T) {
	err := r.UnFreeze(did)
	assert.Nil(t, err)
	item, _, _, err := r.Resolve(did)
	assert.Nil(t, err)
	assert.Equal(t, Normal, item.Status)
}

func TestDIDDeleteSucceed(t *testing.T) {
	err := r.Delete(did)
	assert.Nil(t, err)
	err = r.Delete(rootDID)
	assert.Nil(t, err)
}

func TestDIDCloseSucceed(t *testing.T) {
	err := r.table.Close()
	assert.Nil(t, err)
	err = r.docdb.Close()
	assert.Nil(t, err)

	err = os.RemoveAll(drtPath)
	assert.Nil(t, err)
	err = os.RemoveAll(ddbPath)
	assert.Nil(t, err)
}
