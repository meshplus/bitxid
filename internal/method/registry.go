package method

import (
	"errors"
	"fmt"

	"github.com/meshplus/bitxhub-kit/storage"
	"github.com/meshplus/bitxid/internal/common/docdb"
	"github.com/meshplus/bitxid/internal/common/registry"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
)

// the rule of status code:
// end with 1 (001, 101, 301, etc.) means on audit
// end with 5 (005, 105, 205, 305, etc.) means audit failed
// end with 0 (010, 110, 200, 310, etc.) means good
// difference between
const (
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
	table  *registry.Table
	docdb  *docdb.DocDB
	logger logrus.FieldLogger
	admins []did // admins of the registry
	// network
}

// Item is item in Registry
// registry table is used together with docdb
// we suggest to store large data off-chain(in docdb)
// and only some frequently used data on-chain(in cache)
type Item struct {
	key     string // primary key of the item, like a did
	owner   did    // owner of the method, is a did
	docAddr string // addr where the doc file stored
	docHash []byte // hash of the doc file
	status  int    // status of the item
	cache   []byte // onchain storage part
}

type did string

// New a MethodRegistry
func New(S1 storage.Storage, S2 storage.Storage, L logrus.FieldLogger) (*Registry, error) {
	rt, err := registry.NewTable(S1)
	if err != nil {
		fmt.Println("[registry.NewTable] err", err)
		return nil, err
	}
	db, err := docdb.NewDB(S2)
	if err != nil {
		fmt.Println("[docdb.NewDB] err", err)
		return nil, err
	}
	return &Registry{
		table:  rt,
		docdb:  db,
		logger: L,
		admins: []did{""},
	}, nil
}

// SetupGenesis set up genesis to boot the whole methed system
func SetupGenesis() {}

// Apply apply rights for a new methd-name
func (R *Registry) Apply(caller did, method string, sig []byte) error {
	// check if did exists
	// ..

	// check if Method Name meets standard
	// ..

	// check if Method exists
	exist, err := R.HasMethod(method)
	if err != nil {
		R.logger.Error("[R.HasMethod] err:", err)
		return err
	}
	if exist == true {
		return errors.New("The Method is ALREADY existed")
	}
	// creates item in table
	err = R.table.CreateItem([]byte(method),
		Item{
			key:    method,
			status: ApplyAudit,
			owner:  caller,
		})
	if err != nil {
		R.logger.Error("[R.table.CreateItem] err:", err)
		return err
	}
	return nil
}

// AuditApply .
// ATNS: only admin can call this.
func (R *Registry) AuditApply(caller did, method string, result bool, sig []byte) error {
	exist, err := R.HasMethod(method)
	if err != nil {
		R.logger.Error("[R.HasMethod] err:", err)
		return err
	}
	if exist == false {
		return errors.New("The Method NOT existed")
	}
	// check caller auth(admin of the bitxhub)
	// ...
	if !R.isStatus(method, ApplyAudit) {
		return errors.New("Can not AuditApply for current status")
	}
	if result {
		err = R.auditStatus(method, ApplySuccess)
	} else {
		err = R.auditStatus(method, ApplyFailed)
	}
	if err != nil {
		R.logger.Error("[R.auditStatus] err:", err)
		return err
	}
	return nil
}

// Register ties method name to a method doc
// ATN: only did who owns method-name can call this
func (R *Registry) Register(caller did, method string, doc []byte, sig []byte) (string, string, error) {
	exist, err := R.HasMethod(method)
	if err != nil {
		R.logger.Error("[R.HasMethod] err:", err)
		return "", "", err
	}
	if exist == false {
		return "", "", errors.New("The Method NOT existed")
	}
	// only did who owns method-name can call this
	if !R.owns(caller, method) {
		return "", "", errors.New("Caller has no auth")
	}
	if !R.isStatus(method, ApplySuccess) {
		return "", "", errors.New("Can not Register for current status")
	}

	docAddr, err := R.docdb.Create([]byte(method), doc)
	if err != nil {
		R.logger.Error("[R.docdb.Create] err:", err)
		return "", "", err
	}
	docHash := sha3.Sum512(doc) // docHash := sha256.Sum256(doc)
	// update registry table:
	err = R.table.UpdateItem([]byte(method),
		Item{
			key:     method,
			docAddr: docAddr,
			docHash: docHash[:],
			status:  RegisterSuccess,
			owner:   caller,
		})
	if err != nil {
		R.logger.Error("[R.table.UpdateItem] err:", err)
		return docAddr, string(docHash[:]), err
	}
	// SyncToPeer
	// ...
	return docAddr, string(docHash[:]), nil
}

// Update .
// ATN: only did who owns method-name can call this
func (R *Registry) Update(caller did, method string, doc []byte, sig []byte) (string, string, error) {
	// check auth
	exist, err := R.HasMethod(method)
	if err != nil {
		R.logger.Error("[R.HasMethod] err:", err)
		return "", "", err
	}
	if exist == false {
		return "", "", errors.New("The Method NOT existed")
	}
	// only did who owns method-name can call this
	if !R.owns(caller, method) {
		return "", "", errors.New("Caller has no auth")
	}
	if !R.isStatus(method, Normal) {
		return "", "", errors.New("Can not Update for current status")
	}

	docAddr, err := R.docdb.Update([]byte(method), doc)
	if err != nil {
		R.logger.Error("[R.docdb.Update] err:", err)
		return "", "", err
	}
	// docHash := sha256.Sum256(doc)
	docHash := sha3.Sum512(doc)
	err = R.table.UpdateItem([]byte(method),
		Item{
			key:     method,
			docAddr: docAddr,
			docHash: docHash[:],
			status:  RegisterSuccess,
			owner:   caller,
		})
	if err != nil {
		R.logger.Error("[R.table.UpdateItem] err:", err)
		return docAddr, string(docHash[:]), err
	}

	// SyncToPeer
	// ...
	return docAddr, string(docHash[:]), nil
}

// Audit .
// ATN: only admin can call this.
func (R *Registry) Audit(caller did, method string, status int, sig []byte) error {
	exist, err := R.HasMethod(method)
	if err != nil {
		R.logger.Error("[R.HasMethod] err:", err)
		return err
	}
	if exist == false {
		return errors.New("The Method NOT existed")
	}
	// check caller auth(admin of the bitxhub)
	// ...
	err = R.auditStatus(method, status)

	return nil
}

// Freeze .
// ATN: only someone can call this.
func (R *Registry) Freeze(caller did, method string, sig []byte) error {
	exist, err := R.HasMethod(method)
	if err != nil {
		R.logger.Error("[R.HasMethod] err:", err)
		return err
	}
	if exist == false {
		return errors.New("The Method NOT existed")
	}
	// check caller auth(admin of the bitxhub)
	// ...
	err = R.auditStatus(method, Frozen)

	return nil
}

// UnFreeze .
// ATN: only someone can call this.
func (R *Registry) UnFreeze(caller did, method string, sig []byte) error {
	exist, err := R.HasMethod(method)
	if err != nil {
		R.logger.Error("[R.HasMethod] err:", err)
		return err
	}
	if exist == false {
		return errors.New("The Method NOT existed")
	}
	// check caller auth(admin of the bitxhub)
	// ...
	err = R.auditStatus(method, Normal)

	return nil
}

// Resolve .
func (R *Registry) Resolve(caller did, method string, sig []byte) (Item, []byte, error) {
	item := Item{}
	exist, err := R.HasMethod(method)
	if err != nil {
		R.logger.Error("[R.HasMethod] err:", err)
		return Item{}, []byte{}, err
	}
	if exist == false {
		return Item{}, []byte{}, errors.New("The Method NOT existed")
	}
	// look up local-chain first
	err = R.table.GetItem([]byte(method), &item)
	if err != nil {
		R.logger.Error("[R.table.GetItem] err:", err)
		return Item{}, []byte{}, err
	}
	doc, err := R.docdb.Get([]byte(method))
	if err != nil {
		R.logger.Error("[R.docdb.Get] err:", err)
		return item, []byte{}, err
	}
	return item, doc, nil
}

func (R *Registry) owns(caller did, method string) bool {
	item := Item{}
	err := R.table.GetItem([]byte(method), &item)
	if err != nil {
		R.logger.Error("[R.docdb.Get] err:", err)
		return false
	}
	if item.owner == caller {
		return true
	}
	return false
}

func (R *Registry) isStatus(method string, status int) bool {
	item := Item{}
	R.table.GetItem([]byte(method), &item)
	if item.status == status {
		return true
	}
	return false
}

// Audit .
func (R *Registry) auditStatus(method string, status int) error {
	err := R.table.UpdateItem([]byte(method),
		Item{
			key:    method,
			status: status,
		})
	if err != nil {
		R.logger.Error("[R.table.UpdateItem] err:", err)
		return err
	}
	return nil
}

// HasMethod .
func (R *Registry) HasMethod(method string) (bool, error) {
	exist, err := R.table.HasItem([]byte(method))
	if err != nil {
		R.logger.Error("[R.table.HasItem] err:", err)
		return false, err
	}
	return exist, err
}
