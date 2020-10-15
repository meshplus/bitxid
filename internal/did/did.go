package did

import (
	"fmt"
	"strings"

	"github.com/meshplus/bitxhub-kit/storage"
	"github.com/meshplus/bitxid/internal/common/docdb"
	"github.com/meshplus/bitxid/internal/common/registry"
	"github.com/meshplus/bitxid/internal/common/types"
	"github.com/meshplus/bitxid/internal/common/utils"
	"github.com/meshplus/bitxid/internal/repo"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
)

// the rule of status code:
// end with 1 (001, 101, 301, etc.) means on audit
// end with 5 (005, 105, 205, 305, etc.) means audit failed
// end with 0 (010, 110, 200, 310, etc.) means good
const (
	Error   int = -001
	Initial int = 000
	Normal  int = 200
	Frozen  int = 205
)

// Registry for DID Identifier.
// Every Appchain should implements this DID Registry module.
type Registry struct {
	config *repo.DIDConfig
	table  *registry.Table
	docdb  *docdb.DocDB
	logger logrus.FieldLogger
	admins []types.DID // admins of the registry
	// network
}

// Item reperesentis a did item.
// Registry table is used together with docdb,
// we suggest to store large data off-chain(in docdb),
// only some frequently used data on-chain(in cache).
type Item struct {
	Identifier types.DID // primary key of the item, like a did
	DocAddr    string    // addr where the doc file stored
	DocHash    []byte    // hash of the doc file
	Status     int       // status of the item
	Cache      []byte    // onchain storage part
}

// Doc .
type Doc struct {
	types.BasicDoc
	Service string `json:"service"`
}

// New a MethodRegistry
func New(S1 storage.Storage, S2 storage.Storage, L logrus.FieldLogger, MC *repo.DIDConfig) (*Registry, error) {
	rt, err := registry.NewTable(S1)
	if err != nil {
		L.Error("did [New] registry.NewTable err", err)
		return nil, err
	}
	db, err := docdb.NewDB(S2)
	if err != nil {
		L.Error("did [New] docdb.NewDB err", err)
		return nil, err
	}
	return &Registry{
		config: MC,
		table:  rt,
		docdb:  db,
		logger: L,
		admins: []types.DID{""},
	}, nil
}

// UnmarshalDoc convert byte doc to struct doc
func UnmarshalDoc(docBytes []byte) (Doc, error) {
	docStruct := Doc{}
	err := utils.Bytes2Struct(docBytes, &docStruct)
	if err != nil {
		return Doc{}, err
	}
	return docStruct, nil
}

// MarshalDoc convert struct doc to byte doc
func MarshalDoc(docStruct Doc) ([]byte, error) {
	docBytes, err := utils.Struct2Bytes(docStruct)
	if err != nil {
		return []byte{}, err
	}
	return docBytes, nil
}

// SetupGenesis set up genesis to boot the whole did registry
func (R *Registry) SetupGenesis() error {
	caller := types.DID(R.config.GenesisAdmin)
	// register genesis did:
	_, _, err := R.Register(R.config.GenesisAccount, caller, []byte(R.config.GenesisDoc), []byte(""))
	if err != nil {
		R.logger.Error("did [SetupGenesis] register admin fail:", err)
		return err
	}
	// add admins did:
	R.admins = append(R.admins, caller)

	return nil
}

// Register ties method name to a method doc
// ATN: only did who owns method-name can call this
func (R *Registry) Register(caller string, did types.DID, doc []byte, sig []byte) (string, string, error) {
	exist, err := R.HasDID(did)
	if err != nil {
		R.logger.Error("did [Register] R.HasDID err:", err)
		return "", "", err
	}
	if exist == true {
		return "", "", fmt.Errorf("did [Register] The DID Already existed")
	}
	// check if caller owns the did
	if !R.owns(caller, did) {
		return "", "", fmt.Errorf("did [Register] Caller(%s) does not own this DID(%s)", caller, string(did))
	}

	docAddr, err := R.docdb.Create([]byte(did), doc)
	if err != nil {
		R.logger.Error("did [Register] R.docdb.Create err:", err)
		return "", "", err
	}
	docHash := sha3.Sum512(doc)
	// update registry table:

	err = R.table.CreateItem([]byte(did),
		Item{
			Identifier: did,
			Status:     Normal,
			DocAddr:    docAddr,
			DocHash:    docHash[:],
		})
	if err != nil {
		R.logger.Error("[Apply] R.table.CreateItem err:", err)
		return docAddr, string(docHash[:]), err
	}

	// SyncToPeer
	// ...
	return docAddr, string(docHash[:]), nil
}

// Update .
// ATN: only caller who owns did can call this
func (R *Registry) Update(caller string, did types.DID, doc []byte, sig []byte) (string, string, error) {
	// check exist
	exist, err := R.HasDID(did)
	if err != nil {
		R.logger.Error("did [Update] R.HasDID err:", err)
		return "", "", err
	}
	if exist == false {
		return "", "", fmt.Errorf("did [Update] The DID NOT existed")
	}
	// only caller who owns did can call this
	if !R.owns(caller, did) {
		return "", "", fmt.Errorf("did [Update] Caller does not own this DID")
	}
	status := R.getDIDStatus(did)
	if status != Normal {
		return "", "", fmt.Errorf("did [Update] Can not Update for current status: %d", status)
	}

	docAddr, err := R.docdb.Update([]byte(did), doc)
	if err != nil {
		R.logger.Error("did [Update] R.docdb.Update err:", err)
		return "", "", err
	}
	docHash := sha3.Sum512(doc)
	item := Item{}
	err = R.table.GetItem([]byte(did), &item)
	if err != nil {
		R.logger.Error("did [Update] R.table.GetItem err:", err)
		return docAddr, string(docHash[:]), err
	}
	item.DocAddr = docAddr
	item.DocHash = docHash[:]
	err = R.table.UpdateItem([]byte(did), item)
	if err != nil {
		R.logger.Error("did [Update] R.table.UpdateItem err:", err)
		return docAddr, string(docHash[:]), err
	}

	// SyncToPeer
	// ...
	return docAddr, string(docHash[:]), nil
}

// Resolve looks up local-chain to resolve did.
func (R *Registry) Resolve(did types.DID) (Item, []byte, error) {
	item := Item{}
	exist, err := R.HasDID(did)
	if err != nil {
		R.logger.Error("did [Resolve] R.HasDID err:", err)
		return Item{}, []byte{}, err
	}
	if exist == false {
		return Item{}, []byte{}, fmt.Errorf("did [Resolve] The Method NOT existed")
	}

	err = R.table.GetItem([]byte(did), &item)
	if err != nil {
		R.logger.Error("did [Resolve] R.table.GetItem err:", err)
		return Item{}, []byte{}, err
	}
	doc, err := R.docdb.Get([]byte(did))
	if err != nil {
		R.logger.Error("did [Resolve] R.docdb.Get err:", err)
		return item, []byte{}, err
	}
	return item, doc, nil
}

// Freeze .
// ATN: only someone can call this.
func (R *Registry) Freeze(caller types.DID, did types.DID, sig []byte) error {
	exist, err := R.HasDID(did)
	if err != nil {
		R.logger.Error("[Freeze] R.HasMethod err:", err)
		return err
	}
	if exist == false {
		return fmt.Errorf("[Freeze] The Method NOT existed")
	}
	// check caller auth(admin of the bitxhub)
	// ...
	err = R.auditStatus(did, Frozen)

	return nil
}

// UnFreeze .
// ATN: only someone can call this.
func (R *Registry) UnFreeze(caller types.DID, did types.DID, sig []byte) error {
	exist, err := R.HasDID(did)
	if err != nil {
		R.logger.Error("[UnFreeze] R.HasDID err:", err)
		return err
	}
	if exist == false {
		return fmt.Errorf("[UnFreeze] The DID NOT existed")
	}
	// check caller auth(admin of the bitxhub)
	// ...
	err = R.auditStatus(did, Normal)

	return nil
}

// Delete .
func (R *Registry) Delete(did types.DID, sig []byte) error {
	// if !R.owns(caller, did) {
	// 	return fmt.Errorf("[Delete] Caller has no auth")
	// }

	// verify sig ann auth

	err := R.auditStatus(did, Initial)
	if err != nil {
		R.logger.Error("[Delete] R.auditStatus err:", err)
		return err
	}

	err = R.docdb.Delete([]byte(did))
	if err != nil {
		R.logger.Error("[Delete] R.docdb.Delete err:", err)
		return err
	}
	err = R.table.DeleteItem([]byte(did))
	if err != nil {
		R.logger.Error("[Delete] R.table.DeleteItem err:", err)
		return err
	}

	return nil
}

// HasDID .
func (R *Registry) HasDID(did types.DID) (bool, error) {
	exist, err := R.table.HasItem([]byte(did))
	if err != nil {
		R.logger.Error("did [HasDID] R.table.HasItem err:", err)
		return false, err
	}
	return exist, err
}

func (R *Registry) getDIDStatus(did types.DID) int {
	item := Item{}
	err := R.table.GetItem([]byte(did), &item)
	if err != nil {
		R.logger.Error("did [getDIDStatus] R.table.GetItem err:", err)
		return Initial
	}
	return item.Status
}

//  caller naturally owns the did ended with his address.
func (R *Registry) owns(caller string, did types.DID) bool {
	s := strings.Split(string(did), ":")
	if s[len(s)-1] == caller {
		return true
	}
	// need sig verify ...
	return false
}

func (R *Registry) auditStatus(did types.DID, status int) error {
	item := &Item{}
	err := R.table.GetItem([]byte(did), &item)
	if err != nil {
		R.logger.Error("did [auditStatus] R.table.GetItem err:", err)
		return err
	}
	item.Status = status
	err = R.table.UpdateItem([]byte(did), item)
	if err != nil {
		R.logger.Error("did [auditStatus] R.table.UpdateItem err:", err)
		return err
	}
	return nil
}
