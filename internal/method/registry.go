package method

import (
	"fmt"

	"github.com/meshplus/bitxhub-kit/storage"
	"github.com/meshplus/bitxid/internal/common/docdb"
	"github.com/meshplus/bitxid/internal/common/registry"
	"github.com/meshplus/bitxid/internal/common/types"
	"github.com/meshplus/bitxid/internal/repo"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
)

// the rule of status code:
// end with 1 (001, 101, 301, etc.) means on audit
// end with 5 (005, 105, 205, 305, etc.) means audit failed
// end with 0 (010, 110, 200, 310, etc.) means good
// 101/105/110 301/305/310 not used currently
const (
	Error           int = -001
	Initial         int = 000
	ApplyAudit      int = 001
	ApplyFailed     int = 005
	ApplySuccess    int = 010
	RegisterAudit   int = 101
	RegisterFailed  int = 105
	RegisterSuccess int = 110
	Normal          int = 200
	Frozen          int = 205
	UpdateAudit     int = 301
	UpdateFailed    int = 305
	UpdateSuccess   int = 310
)

const Size int = 64

// Registry .
type Registry struct {
	config *repo.MethodConfig
	table  *registry.Table
	docdb  *docdb.DocDB
	logger logrus.FieldLogger
	admins []types.DID // admins of the registry
	// network
}

// Item reperesentis a method item.
// Registry table is used together with docdb,
// we suggest to store large data off-chain(in docdb)
// only some frequently used data on-chain(in cache).
type Item struct {
	Key     string    // primary key of the item, like a did
	Owner   types.DID // owner of the method, is a did
	DocAddr string    // addr where the doc file stored
	DocHash []byte    // hash of the doc file
	Status  int       // status of the item
	Cache   []byte    // onchain storage part
}

// New a MethodRegistry
func New(S1 storage.Storage, S2 storage.Storage, L logrus.FieldLogger, MC *repo.MethodConfig) (*Registry, error) {
	rt, err := registry.NewTable(S1)
	if err != nil {
		L.Error("[New] registry.NewTable err", err)
		return nil, err
	}
	db, err := docdb.NewDB(S2)
	if err != nil {
		L.Error("[New] docdb.NewDB err", err)
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

// SetupGenesis set up genesis to boot the whole methed system
func (R *Registry) SetupGenesis(doc []byte) (string, string, error) {
	if !R.config.IsRoot {
		return "", "", fmt.Errorf("[SetupGenesis] This method registry is not a relay root, check the config")
	}
	caller := types.DID(R.config.GenesisAdmin)
	// register genesis method:
	method := R.config.GenesisMetohd
	docAddr, err := R.docdb.Create([]byte(method), doc)
	if err != nil {
		R.logger.Error("[SetupGenesis] R.docdb.Create err:", err)
		return "", "", err
	}
	docHash := sha3.Sum512(doc)
	err = R.table.CreateItem([]byte(method),
		Item{
			Key:     method,
			DocAddr: docAddr,
			DocHash: docHash[:],
			Status:  Normal,
			Owner:   caller,
		})
	if err != nil {
		R.logger.Error("[SetupGenesis] R.table.CreateItem err:", err)
		return docAddr, string(docHash[:]), err
	}
	// add admins did:
	R.admins = append(R.admins, caller)

	return docAddr, string(docHash[:]), nil
}

// Apply apply rights for a new methd-name
func (R *Registry) Apply(caller types.DID, method string, sig []byte) error {
	// check if did exists
	// ..

	// check if Method Name meets standard
	// ..

	// check if Method exists
	exist, err := R.HasMethod(method)
	if err != nil {
		R.logger.Error("[Apply] R.HasMethod err:", err)
		return err
	}
	if exist == true {
		return fmt.Errorf("[Apply] The Method is ALREADY existed")
	}
	//
	status := R.getMethodStatus(method)
	if status != Initial {
		return fmt.Errorf("[Apply] Can not Register for current status: %d", status)
	}
	// creates item in table
	err = R.table.CreateItem([]byte(method),
		Item{
			Key:    method,
			Status: ApplyAudit,
			Owner:  caller,
		})
	if err != nil {
		R.logger.Error("[Apply] R.table.CreateItem err:", err)
		return err
	}
	return nil
}

// AuditApply .
// ATNS: only admin can call this.
func (R *Registry) AuditApply(caller types.DID, method string, result bool, sig []byte) error {
	exist, err := R.HasMethod(method)
	if err != nil {
		R.logger.Error("[AuditApply] R.HasMethod err:", err)
		return err
	}
	if exist == false {
		return fmt.Errorf("[AuditApply] The Method NOT existed")
	}
	// check caller auth(admin of the bitxhub)
	// ...
	status := R.getMethodStatus(method)
	if !(status == ApplyAudit || status == ApplyFailed) {
		return fmt.Errorf("[AuditApply] Can not AuditApply for current status: %d", status)
	}
	if result {
		err = R.auditStatus(method, ApplySuccess)
	} else {
		err = R.auditStatus(method, ApplyFailed)
	}
	if err != nil {
		R.logger.Error("[AuditApply] R.auditStatus err:", err)
		return err
	}
	return nil
}

// Register ties method name to a method doc
// ATN: only did who owns method-name can call this
func (R *Registry) Register(caller types.DID, method string, doc []byte, sig []byte) (string, string, error) {
	exist, err := R.HasMethod(method)
	if err != nil {
		R.logger.Error("[Register] R.HasMethod err:", err)
		return "", "", err
	}
	if exist == false {
		return "", "", fmt.Errorf("[Register] The Method NOT existed")
	}
	// only did who owns method-name can call this
	if !R.owns(caller, method) {
		return "", "", fmt.Errorf("[Register] Caller has no auth")
	}
	status := R.getMethodStatus(method)
	if status != ApplySuccess {
		return "", "", fmt.Errorf("[Register] Can not Register for current status: %d", status)
	}

	docAddr, err := R.docdb.Create([]byte(method), doc)
	if err != nil {
		R.logger.Error("[Register] R.docdb.Create err:", err)
		return "", "", err
	}
	docHash := sha3.Sum512(doc) // docHash := sha256.Sum256(doc)
	// update registry table:
	item := &Item{}
	err = R.table.GetItem([]byte(method), &item)
	if err != nil {
		R.logger.Error("[Register] R.table.GetItem err:", err)
		return "", "", err
	}
	item.Status = Normal
	item.DocAddr = docAddr
	item.DocHash = docHash[:]
	err = R.table.UpdateItem([]byte(method), item)
	if err != nil {
		R.logger.Error("[Register] R.table.UpdateItem err:", err)
		return docAddr, string(docHash[:]), err
	}
	// SyncToPeer
	// ...
	return docAddr, string(docHash[:]), nil
}

// Update .
// ATN: only did who owns method-name can call this
func (R *Registry) Update(caller types.DID, method string, doc []byte, sig []byte) (string, string, error) {
	// check exist
	exist, err := R.HasMethod(method)
	if err != nil {
		R.logger.Error("[Update] R.HasMethod err:", err)
		return "", "", err
	}
	if exist == false {
		return "", "", fmt.Errorf("[Update] The Method NOT existed")
	}
	// only did who owns method-name can call this
	if !R.owns(caller, method) {
		return "", "", fmt.Errorf("[Update] Caller has no auth")
	}
	status := R.getMethodStatus(method)
	if status != Normal {
		return "", "", fmt.Errorf("[Update] Can not Update for current status: %d", status)
	}

	docAddr, err := R.docdb.Update([]byte(method), doc)
	if err != nil {
		R.logger.Error("[Update] R.docdb.Update err:", err)
		return "", "", err
	}
	// docHash := sha256.Sum256(doc)
	docHash := sha3.Sum512(doc)
	item := Item{}
	err = R.table.GetItem([]byte(method), &item)
	if err != nil {
		R.logger.Error("did [Update] R.table.GetItem err:", err)
		return docAddr, string(docHash[:]), err
	}
	item.DocAddr = docAddr
	item.DocHash = docHash[:]
	err = R.table.UpdateItem([]byte(method), item)

	if err != nil {
		R.logger.Error("[Update] R.table.UpdateItem err:", err)
		return docAddr, string(docHash[:]), err
	}

	// SyncToPeer
	// ...
	return docAddr, string(docHash[:]), nil
}

// Audit .
// ATN: only admin can call this.
func (R *Registry) Audit(caller types.DID, method string, status int, sig []byte) error {
	exist, err := R.HasMethod(method)
	if err != nil {
		R.logger.Error("[Audit] R.HasMethod err:", err)
		return err
	}
	if exist == false {
		return fmt.Errorf("[Audit] The Method NOT existed")
	}
	// check caller auth(admin of the bitxhub)
	// ...
	err = R.auditStatus(method, status)

	return nil
}

// Freeze .
// ATN: only someone can call this.
func (R *Registry) Freeze(caller types.DID, method string, sig []byte) error {
	exist, err := R.HasMethod(method)
	if err != nil {
		R.logger.Error("[Freeze] R.HasMethod err:", err)
		return err
	}
	if exist == false {
		return fmt.Errorf("[Freeze] The Method NOT existed")
	}
	// check caller auth(admin of the bitxhub)
	// ...
	err = R.auditStatus(method, Frozen)

	return nil
}

// UnFreeze .
// ATN: only someone can call this.
func (R *Registry) UnFreeze(caller types.DID, method string, sig []byte) error {
	exist, err := R.HasMethod(method)
	if err != nil {
		R.logger.Error("[UnFreeze] R.HasMethod err:", err)
		return err
	}
	if exist == false {
		return fmt.Errorf("[UnFreeze] The Method NOT existed")
	}
	// check caller auth(admin of the bitxhub)
	// ...
	err = R.auditStatus(method, Normal)

	return nil
}

// Delete .
func (R *Registry) Delete(caller types.DID, method string, sig []byte) error {
	err := R.auditStatus(method, Initial)
	if err != nil {
		R.logger.Error("[Delete] R.auditStatus err:", err)
		return err
	}

	if !R.owns(caller, method) {
		return fmt.Errorf("[Delete] Caller has no auth")
	}

	err = R.docdb.Delete([]byte(method))
	if err != nil {
		R.logger.Error("[Delete] R.docdb.Delete err:", err)
		return err
	}
	err = R.table.DeleteItem([]byte(method))
	if err != nil {
		R.logger.Error("[Delete] R.table.DeleteItem err:", err)
		return err
	}

	return nil
}

// Resolve .
func (R *Registry) Resolve(caller types.DID, method string, sig []byte) (Item, []byte, error) {
	item := Item{}
	// looks up local-chain first:
	exist, err := R.HasMethod(method)
	if err != nil {
		R.logger.Error("[Resolve] R.HasMethod err:", err)
		return Item{}, []byte{}, err
	}
	if exist == false {
		return Item{}, []byte{}, fmt.Errorf("[Resolve] The Method NOT existed")
	}

	err = R.table.GetItem([]byte(method), &item)
	if err != nil {
		R.logger.Error("[Resolve] R.table.GetItem err:", err)
		return Item{}, []byte{}, err
	}
	doc, err := R.docdb.Get([]byte(method))
	if err != nil {
		R.logger.Error("[Resolve] R.docdb.Get err:", err)
		return item, []byte{}, err
	}
	return item, doc, nil
}

// MethodHasAccount checks whether account exists on the method blockchain
func (R *Registry) MethodHasAccount(method string, account string) {

}

// HasMethod .
func (R *Registry) HasMethod(method string) (bool, error) {
	exist, err := R.table.HasItem([]byte(method))
	if err != nil {
		R.logger.Error("[HasMethod] R.table.HasItem err:", err)
		return false, err
	}
	return exist, err
}

func (R *Registry) owns(caller types.DID, method string) bool {
	item := Item{}
	err := R.table.GetItem([]byte(method), &item)
	if err != nil {
		R.logger.Error("[owns] R.table.GetItem err: ", err)
		return false
	}
	if item.Owner == caller {
		return true
	}
	return false
}

func (R *Registry) getMethodStatus(method string) int {
	item := Item{}
	err := R.table.GetItem([]byte(method), &item)
	if err != nil {
		R.logger.Warn("[getMethodStatus] R.table.GetItem err:", err)
		return Initial
	}
	return item.Status
}

// auditStatus .
func (R *Registry) auditStatus(method string, status int) error {
	item := &Item{}
	err := R.table.GetItem([]byte(method), &item)
	if err != nil {
		R.logger.Error("[auditStatus] R.table.GetItem err:", err)
		return err
	}
	item.Status = status
	err = R.table.UpdateItem([]byte(method), item)
	if err != nil {
		R.logger.Error("[auditStatus] R.table.UpdateItem err:", err)
		return err
	}
	return nil
}
