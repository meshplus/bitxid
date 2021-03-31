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
// var r *AccountDIDRegistry

var rootAccountDID DID = DID("did:bitxhub:appchain001:0x00000001")
var testAccountDID DID = DID("did:bitxhub:appchain001:0x12345678")

var accountDoc AccountDoc = getAccountDoc(0)
var accountDocA AccountDoc = getAccountDoc(1)
var accountDocB AccountDoc = getAccountDoc(2)

func getAccountDoc(ran int) AccountDoc {
	docE := AccountDoc{}
	docE.ID = testAccountDID
	docE.Type = int(AccountDIDType)
	docE.Created = 1617006461
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
		docE.ID = rootAccountDID
	case 1:
		docE.PublicKey = []PubKey{pk1}
	case 2:
		docE.PublicKey = []PubKey{pk2}
	}

	auth := Auth{
		PublicKey: []string{"KEY#1"},
	}
	docE.Authentication = []Auth{auth}
	return docE
}

func TestDIDModeInternal(t *testing.T) {
	r, drtPath, ddbPath := newDIDModeInternal(t)

	testSetupDIDSucceed(t, r)
	testHasDIDSucceed(t, r)
	testDIDAddAdminsSucceed(t, r)
	testDIDRemoveAdminsSucceed(t, r)

	testDIDRegisterSucceedInternal(t, r)
	testDIDUpdateSucceedInternal(t, r)
	testDIDResolveSucceedInternal(t, r)

	testDIDFreezeSucceed(t, r)
	testDIDUnFreezeSucceed(t, r)
	testDIDDeleteSucceed(t, r)
	testDIDCloseSucceedInternal(t, r, drtPath, ddbPath)
}

func TestDIDModeExternal(t *testing.T) {
	r, drtPath := newDIDModeExternal(t)
	testSetupDIDSucceed(t, r)
	testHasDIDSucceed(t, r)
	testDIDAddAdminsSucceed(t, r)
	testDIDRemoveAdminsSucceed(t, r)

	testDIDRegisterSucceedExternal(t, r)
	testDIDUpdateSucceedExternal(t, r)
	testDIDResolveSucceedExternal(t, r)

	testDIDFreezeSucceed(t, r)
	testDIDUnFreezeSucceed(t, r)
	testDIDDeleteSucceed(t, r)
	testDIDCloseSucceedExternal(t, r, drtPath)
}

func newDIDModeInternal(t *testing.T) (*AccountDIDRegistry, string, string) {
	dir1, err := ioutil.TempDir("", "did.table") //
	assert.Nil(t, err)
	dir2, err := ioutil.TempDir("", "did.docdb")
	assert.Nil(t, err)

	drtPath := dir1
	ddbPath := dir2
	loggerInit()
	l := loggerGet(loggerAccountDID)
	s1, err := leveldb.New(drtPath)
	assert.Nil(t, err)
	s2, err := leveldb.New(ddbPath)
	assert.Nil(t, err)
	r, err := NewAccountDIDRegistry(s1, l,
		WithAccountDocStorage(s2),
		WithDIDAdmin(rootAccountDID),
		WithGenesisAccountDocContent(&accountDoc),
	)
	assert.Nil(t, err)
	return r, drtPath, ddbPath
}

func newDIDModeExternal(t *testing.T) (*AccountDIDRegistry, string) {
	dir1, err := ioutil.TempDir("", "did.table")
	assert.Nil(t, err)
	drtPath := dir1
	loggerInit()
	l := loggerGet(loggerAccountDID)
	s1, err := leveldb.New(drtPath)
	assert.Nil(t, err)
	r, err := NewAccountDIDRegistry(s1, l,
		WithDIDAdmin(rootAccountDID),
		WithGenesisAccountDocInfo(DocInfo{rootAccountDID, "/addr/to/doc", []byte{1}}),
	)
	assert.Nil(t, err)
	return r, drtPath
}

func testSetupDIDSucceed(t *testing.T, r *AccountDIDRegistry) {
	err := r.SetupGenesis()
	assert.Nil(t, err)
}

func testHasDIDSucceed(t *testing.T, r *AccountDIDRegistry) {
	ret1 := r.HasAccountDID(r.GenesisAccountDID)
	assert.Equal(t, true, ret1)
}

func testDIDAddAdminsSucceed(t *testing.T, r *AccountDIDRegistry) {
	err := r.AddAdmin(admin)
	assert.Nil(t, err)
	ret := r.HasAdmin(admin)
	assert.Equal(t, true, ret)
}

func testDIDRemoveAdminsSucceed(t *testing.T, r *AccountDIDRegistry) {
	err := r.RemoveAdmin(admin)
	assert.Nil(t, err)
	ret := r.HasAdmin(admin)
	assert.Equal(t, false, ret)
}

func testDIDRegisterSucceedInternal(t *testing.T, r *AccountDIDRegistry) {
	docABytes, err := Marshal(accountDocA)
	assert.Nil(t, err)
	docHashE := sha256.Sum256(docABytes)
	docAddrE := "./" + string(testAccountDID)
	docAddr, docHash, err := r.RegisterWithDoc(&accountDocA)
	assert.Nil(t, err)
	strHash := fmt.Sprintf("%x", docHash)
	strHashE := fmt.Sprintf("%x", docHashE)
	assert.Equal(t, strHashE, strHash)
	assert.Equal(t, docAddrE, docAddr)
}

func testDIDRegisterSucceedExternal(t *testing.T, r *AccountDIDRegistry) {
	docABytes, err := Marshal(accountDocA)
	assert.Nil(t, err)
	docHashE := sha256.Sum256(docABytes)
	docAddrE := "./addr/" + string(testAccountDID)
	_, _, err = r.Register(testAccountDID, docAddrE, docHashE[:])
	assert.Nil(t, err)
}

func testDIDUpdateSucceedInternal(t *testing.T, r *AccountDIDRegistry) {
	docBBytes, err := Marshal(accountDocB)
	assert.Nil(t, err)
	docHashE := sha256.Sum256(docBBytes)
	docAddrE := "./" + string(testAccountDID)
	docAddr, docHash, err := r.UpdateWithDoc(&accountDocB)
	assert.Nil(t, err)
	strHash := fmt.Sprintf("%x", docHash)
	strHashE := fmt.Sprintf("%x", docHashE)
	assert.Equal(t, strHashE, strHash)
	assert.Equal(t, docAddrE, docAddr)
}

func testDIDUpdateSucceedExternal(t *testing.T, r *AccountDIDRegistry) {
	docBBytes, err := Marshal(accountDocB)
	assert.Nil(t, err)
	docHashE := sha256.Sum256(docBBytes)
	docAddrE := "/addr/" + string(testAccountDID)
	_, _, err = r.Update(testAccountDID, docAddrE, docHashE[:])
	assert.Nil(t, err)
}

func testDIDResolveSucceedInternal(t *testing.T, r *AccountDIDRegistry) {
	item, doc, _, err := r.Resolve(testAccountDID)
	assert.Nil(t, err)
	docBBytes, err := Marshal(accountDocB)
	assert.Nil(t, err)
	docHashE := sha256.Sum256(docBBytes)
	assert.Nil(t, err)
	assert.Equal(t, &accountDocB, doc) // compare doc
	itemE := AccountItem{
		BasicItem{
			ID:      testAccountDID,
			DocHash: docHashE[:],
			DocAddr: "./" + string(testAccountDID),
			Status:  Normal},
	}
	assert.Equal(t, itemE.ID, item.ID)
	assert.Equal(t, itemE.DocAddr, item.DocAddr)
	assert.Equal(t, itemE.DocHash, item.DocHash)
	assert.Equal(t, itemE.Status, item.Status)
}

func testDIDResolveSucceedExternal(t *testing.T, r *AccountDIDRegistry) {
	item, doc, _, err := r.Resolve(testAccountDID)
	assert.Nil(t, err)
	assert.Nil(t, doc)
	docBBytes, err := Marshal(accountDocB)
	docHashE := sha256.Sum256(docBBytes)
	assert.Nil(t, err)
	itemE := AccountItem{
		BasicItem{
			ID:      testAccountDID,
			DocHash: docHashE[:],
			DocAddr: "/addr/" + string(testAccountDID),
			Status:  Normal},
	}
	assert.Equal(t, itemE.ID, item.ID)
	assert.Equal(t, itemE.DocAddr, item.DocAddr)
	assert.Equal(t, itemE.DocHash, item.DocHash)
	assert.Equal(t, itemE.Status, item.Status)
}

func testDIDFreezeSucceed(t *testing.T, r *AccountDIDRegistry) {
	err := r.Freeze(testAccountDID)
	assert.Nil(t, err)
	item, _, _, err := r.Resolve(testAccountDID)
	assert.Nil(t, err)
	assert.Equal(t, Frozen, item.Status)
}

func testDIDUnFreezeSucceed(t *testing.T, r *AccountDIDRegistry) {
	err := r.UnFreeze(testAccountDID)
	assert.Nil(t, err)
	item, _, _, err := r.Resolve(testAccountDID)
	assert.Nil(t, err)
	assert.Equal(t, Normal, item.Status)
}

func testDIDDeleteSucceed(t *testing.T, r *AccountDIDRegistry) {
	err := r.Delete(testAccountDID)
	assert.Nil(t, err)
	err = r.Delete(rootAccountDID)
	assert.Nil(t, err)
}

func testDIDCloseSucceedInternal(t *testing.T, r *AccountDIDRegistry, path ...string) {
	err := r.Table.Close()
	assert.Nil(t, err)
	err = r.Docdb.Close()
	assert.Nil(t, err)

	for _, p := range path {
		err = os.RemoveAll(p)
		assert.Nil(t, err)
	}
}

func testDIDCloseSucceedExternal(t *testing.T, r *AccountDIDRegistry, path ...string) {
	err := r.Table.Close()
	assert.Nil(t, err)

	for _, p := range path {
		err = os.RemoveAll(p)
		assert.Nil(t, err)
	}
}
