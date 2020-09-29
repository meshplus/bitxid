package method

import (
	"fmt"
	"testing"

	"github.com/meshplus/bitxhub-kit/storage/leveldb"
	"github.com/meshplus/bitxid/internal/loggers"
	"github.com/meshplus/bitxid/internal/repo"
	"github.com/stretchr/testify/assert"
)

const (
	dbPath   string = "../../config/doc.db"
	rtPath   string = "../../config/registry.store"
	confPath string = "../../config"
)

var R *Registry

var docB []byte = []byte("{\"MethodName\":\"did:bitxhub:relayroot:.\",\"Auth\": {}}")
var docC []byte = []byte("{\"MethodName\":\"did:bitxhub:appchain001:.\",\"Auth\": {}}")
var docD []byte = []byte("{\"MethodName\":\"did:bitxhub:appchain001:.\",\"Auth\": {0x12345678}}")

var caller did = did("did:bitxhub:relayroot:0x12345678")

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

func TestSetupGenesis(t *testing.T) {
	docHashE := "d449d8bdd3c92be218033594f5ae694bd7d105bf22b1a42875106a40f290669a56af06d7f6f5f7efcd14fae1798d9bc46ff28332503ab9567bbc00e5977874dc"
	docAddrE := "./did:bitxhub:relayroot:."
	docAddr, docHash, err := R.SetupGenesis(docB)
	assert.Nil(t, err)
	strHash := fmt.Sprintf("%x", docHash)
	assert.Equal(t, docHashE, strHash)
	fmt.Println("docAddr:", docAddr)
	assert.Equal(t, docAddrE, docAddr)
}

func TestHasMethod(t *testing.T) {
	ret1, err := R.HasMethod(R.config.GenesisMetohd)
	assert.Nil(t, err)
	assert.Equal(t, true, ret1)
}

func TestApply(t *testing.T) {
	err := R.Apply(caller, method, sig)
	assert.Nil(t, err)
}

func TestAuditApplyFailed(t *testing.T) {
	err := R.AuditApply(caller, method, false, sig)
	assert.Nil(t, err)
}
func TestAuditApplySuccess(t *testing.T) {
	err := R.AuditApply(caller, method, true, sig)
	assert.Nil(t, err)
}

func TestRegister(t *testing.T) {
	docHashE := "86e48d2030c5443f871c158b82aca22b0ac8e36c5c0b78b1e3834dd272387d6c5581278f16ca9645fc221469c848fa715b3f5673cd635a2f7fe6cfc75cd8a54a"
	docAddrE := "./did:bitxhub:appchain001:."
	docAddr, docHash, err := R.Register(caller, method, docC, sig)
	assert.Nil(t, err)
	strHash := fmt.Sprintf("%x", docHash)
	item, _, err := R.Resolve(caller, method, sig)
	assert.Nil(t, err)
	assert.Equal(t, docHashE, strHash)
	assert.Equal(t, docAddrE, docAddr)
	assert.Equal(t, caller, item.Owner)
}

func TestUpdate(t *testing.T) {
	docHashE := "07fff05bd1be2cbbb1aec145adf58da965d2f7106d0f777eb734586a925e4a6ebc371a0d990154711d13c0be8984b0a66e6c15a57d95bda21f4be0e32e0e4571"
	docAddrE := "./did:bitxhub:appchain001:."
	docAddr, docHash, err := R.Update(caller, method, docD, sig)
	assert.Nil(t, err)
	strHash := fmt.Sprintf("%x", docHash)
	assert.Equal(t, docHashE, strHash)
	assert.Equal(t, docAddrE, docAddr)
}

func TestAuditRegisterSuccess(t *testing.T) {
	err := R.Audit(caller, method, RegisterSuccess, sig)
	assert.Nil(t, err)
	item, _, err := R.Resolve(caller, method, sig)
	assert.Nil(t, err)
	assert.Equal(t, RegisterSuccess, item.Status)
}

func TestAuditUpdateSuccess(t *testing.T) {
	err := R.Audit(caller, method, UpdateSuccess, sig)
	assert.Nil(t, err)
	item, _, err := R.Resolve(caller, method, sig)
	assert.Nil(t, err)
	assert.Equal(t, UpdateSuccess, item.Status)
}

func TestFreeze(t *testing.T) {
	err := R.Freeze(caller, method, sig)
	assert.Nil(t, err)
	item, _, err := R.Resolve(caller, method, sig)
	assert.Nil(t, err)
	assert.Equal(t, Frozen, item.Status)
}

func TestUnFreeze(t *testing.T) {
	err := R.UnFreeze(caller, method, sig)
	assert.Nil(t, err)
	item, _, err := R.Resolve(caller, method, sig)
	assert.Nil(t, err)
	assert.Equal(t, Normal, item.Status)
}

func TestResolve(t *testing.T) {
	docAddrE := "./did:bitxhub:appchain001:."
	// docHashE := "9f9c10312b85589aed30e4fd88676b580640e12ff682f5143ec0cdeba97a8e44239f0b5e152280f6be386b9871828a9084a0054df6ad4c0ce7d0435b2cadb77c"
	item, doc, err := R.Resolve(caller, method, sig)
	assert.Nil(t, err)
	assert.Equal(t, docD, doc)
	itemE := Item{
		Key:     method,
		Owner:   caller,
		DocAddr: docAddrE,
		Status:  Normal,
	}
	assert.Equal(t, item.Key, itemE.Key)
	assert.Equal(t, item.Owner, itemE.Owner)
	assert.Equal(t, item.DocAddr, itemE.DocAddr)
	assert.Equal(t, item.Status, itemE.Status)
}

func TestDelete(t *testing.T) {
	err := R.Delete(caller, method, sig)
	assert.Nil(t, err)
}
func TestClose(t *testing.T) {
	err := R.table.Close()
	assert.Nil(t, err)
}
