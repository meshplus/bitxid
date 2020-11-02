package bitxid

import (
	"fmt"

	"github.com/meshplus/bitxhub-kit/storage"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
)

//
const (
	BLC int = iota
	DLT
)

// Size of hash byte
const Size int = 64

var _ MethodManager = (*MethodRegistry)(nil)
var _ Doc = (*MethodDoc)(nil)

// MethodRegistry .
type MethodRegistry struct {
	config *MethodConfig
	table  RegistryTable
	docdb  DocDB
	logger logrus.FieldLogger
	admins []DID // admins of the registry
	// network
}

// MethodItem reperesents a method item.
// Registry table is used together with docdb,
// we suggest to store large data off-chain(in docdb)
// only some frequently used data on-chain(in cache).
type MethodItem struct {
	Method  DID    // primary key of the item, like a did
	Owner   DID    // owner of the method, is a did
	DocAddr string // addr where the doc file stored
	DocHash []byte // hash of the doc file
	Status  int    // status of the item
	Cache   []byte // onchain storage part
}

// MethodDoc .
type MethodDoc struct {
	BasicDoc
	Parent string `json:"parent"`
}

// Marshal .
func (d *MethodDoc) Marshal() ([]byte, error) {
	return Struct2Bytes(d)
}

// Unmarshal .
func (d *MethodDoc) Unmarshal(docBytes []byte) error {
	return Bytes2Struct(docBytes, &d)
}

// NewMethodRegistry news a MethodRegistry
func NewMethodRegistry(s1 storage.Storage, s2 storage.Storage, l logrus.FieldLogger, mc *MethodConfig) (*MethodRegistry, error) {
	rt, err := NewKVTable(s1)
	if err != nil {
		return nil, fmt.Errorf("method new table: %w", err)
	}
	db, err := NewKVDocDB(s2)
	if err != nil {
		return nil, fmt.Errorf("method new docdb: %w", err)
	}
	return &MethodRegistry{
		table:  rt,
		docdb:  db,
		logger: l,
		config: mc,
		admins: []DID{""},
	}, nil
}

// SetupGenesis set up genesis to boot the whole methed system
func (r *MethodRegistry) SetupGenesis() error {
	if !r.config.IsRoot {
		return fmt.Errorf("method genesis: registry not root")
	}
	if r.config.GenesisMetohd != r.config.GenesisDoc.ID {
		return fmt.Errorf("method genesis: method not matched with doc")
	}
	// register method
	r.Apply(DID(r.config.Admin), DID(r.config.GenesisMetohd))
	r.AuditApply(DID(r.config.GenesisMetohd), true)
	_, _, err := r.Register(r.config.GenesisDoc)
	if err != nil {
		return fmt.Errorf("genesis: %w", err)
	}
	// add admins did
	r.AddAdmin(DID(r.config.Admin))

	return nil
}

// GetAdmins .
func (r *MethodRegistry) GetAdmins() []DID {
	return r.admins
}

// AddAdmin .
func (r *MethodRegistry) AddAdmin(caller DID) error {
	if r.HasAdmin(caller) {
		return fmt.Errorf("caller %s is already an admin", caller)
	}
	r.admins = append(r.admins, caller)
	return nil
}

// HasAdmin .
func (r *MethodRegistry) HasAdmin(caller DID) bool {
	for _, v := range r.admins {
		if v == caller {
			return true
		}
	}
	return false
}

// Apply apply for rights of a new methd-name
func (r *MethodRegistry) Apply(caller DID, method DID) error {
	// check if Method Name meets standard
	if !DID(method).IsValidFormat() {
		return fmt.Errorf("method name is not standard")
	}
	// check if Method exists
	exist, err := r.HasMethod(method)
	if err != nil {
		return err
	}
	if exist == true {
		return fmt.Errorf("apply method %s already existed", method)
	}

	status := r.getMethodStatus(method)
	if status != Initial {
		return fmt.Errorf("can not apply method under status: %d", status)
	}
	// creates item in table
	err = r.table.CreateItem(method,
		MethodItem{
			Method: DID(method),
			Status: ApplyAudit,
			Owner:  caller,
		})
	if err != nil {
		return fmt.Errorf("apply method on table: %w", err)
	}
	return nil
}

// AuditApply .
// ATNS: only admin should call this.
func (r *MethodRegistry) AuditApply(method DID, result bool) error {
	exist, err := r.HasMethod(method)
	if err != nil {
		return err
	}
	if exist == false {
		return fmt.Errorf("auditapply method %s not existed", method)
	}
	status := r.getMethodStatus(method)
	if !(status == ApplyAudit || status == ApplyFailed) {
		return fmt.Errorf("can not auditapply under status: %d", status)
	}
	if result {
		err = r.auditStatus(method, ApplySuccess)
	} else {
		err = r.auditStatus(method, ApplyFailed)
	}
	if err != nil {
		return fmt.Errorf("method auditapply status: %w", err)
	}
	return nil
}

// Register ties method name to a method doc
// ATN: only did who owns method-name should call this
func (r *MethodRegistry) Register(doc MethodDoc) (string, []byte, error) {
	method := DID(doc.ID)
	exist, err := r.HasMethod(method)
	if err != nil {
		return "", nil, err
	}
	if exist == false {
		return "", nil, fmt.Errorf("register method %s not existed", method)
	}
	status := r.getMethodStatus(method)
	if status != ApplySuccess {
		return "", nil, fmt.Errorf("can not register under status: %d", status)
	}

	docBytes, err := doc.Marshal()
	if err != nil {
		return "", nil, fmt.Errorf("method register doc marshal: %w", err)
	}

	docAddr, err := r.docdb.Create(method, &doc)
	if err != nil {
		return "", nil, fmt.Errorf("method register on docdb: %w", err)
	}
	docHash := sha3.Sum512(docBytes) // docHash := sha256.Sum256(doc)
	// update MethodRegistry table
	item := &MethodItem{}
	err = r.table.GetItem(method, &item)
	if err != nil {
		return "", nil, fmt.Errorf("method register table get: %w", err)
	}
	item.Status = Normal
	item.DocAddr = docAddr
	item.DocHash = docHash[:]
	err = r.table.UpdateItem(method, item)
	if err != nil {
		return docAddr, item.DocHash, fmt.Errorf("method register table update: %w", err)
	}
	return docAddr, item.DocHash, nil
}

// Update .
// ATN: only did who owns method-name should call this.
func (r *MethodRegistry) Update(doc MethodDoc) (string, []byte, error) {
	// check exist
	method := DID(doc.ID)
	exist, err := r.HasMethod(method)
	if err != nil {
		return "", nil, err
	}
	if exist == false {
		return "", nil, fmt.Errorf("update method %s not existed", method)
	}
	status := r.getMethodStatus(method)
	if status != Normal {
		return "", nil, fmt.Errorf("can not update under status: %d", status)
	}

	docBytes, err := doc.Marshal()
	if err != nil {
		return "", nil, fmt.Errorf("method update doc marshal: %w", err)
	}

	docAddr, err := r.docdb.Update(method, &doc)
	if err != nil {
		return "", nil, fmt.Errorf("method update on docdb: %w", err)
	}
	docHash := sha3.Sum512(docBytes)

	item := MethodItem{}
	err = r.table.GetItem(method, &item)
	if err != nil {
		return docAddr, docHash[:], fmt.Errorf("method update table get: %w", err)
	}
	item.DocAddr = docAddr
	item.DocHash = docHash[:]
	err = r.table.UpdateItem(method, item)
	if err != nil {
		return docAddr, docHash[:], fmt.Errorf("method update table update: %w", err)
	}

	return docAddr, docHash[:], nil
}

// Audit .
// ATN: only admin should call this.
func (r *MethodRegistry) Audit(method DID, status int) error {
	exist, err := r.HasMethod(method)
	if err != nil {
		return err
	}
	if exist == false {
		return fmt.Errorf("audit method %s not existed", method)
	}
	err = r.auditStatus(method, status)
	if err != nil {
		return fmt.Errorf("method audit status: %w", err)
	}
	return nil
}

// Freeze .
// ATN: only admdin should call this.
func (r *MethodRegistry) Freeze(method DID) error {
	exist, err := r.HasMethod(method)
	if err != nil {
		return err
	}
	if exist == false {
		return fmt.Errorf("freeze method %s not existed", method)
	}
	err = r.auditStatus(method, Frozen)
	if err != nil {
		return fmt.Errorf("method freeze status aduit: %w", err)
	}
	return nil
}

// UnFreeze .
// ATN: only admdin should call this.
func (r *MethodRegistry) UnFreeze(method DID) error {
	exist, err := r.HasMethod(method)
	if err != nil {
		return err
	}
	if exist == false {
		return fmt.Errorf("unfreeze method %s not existed", method)
	}

	err = r.auditStatus(method, Normal)
	if err != nil {
		return fmt.Errorf("method unfreeze status aduit: %w", err)
	}
	return nil
}

// Delete .
func (r *MethodRegistry) Delete(method DID) error {
	err := r.auditStatus(method, Initial)
	if err != nil {
		return fmt.Errorf("method delete status aduit: %w", err)
	}

	err = r.docdb.Delete(method)
	if err != nil {
		return fmt.Errorf("method delete docdb: %w", err)
	}
	err = r.table.DeleteItem(method)
	if err != nil {
		return fmt.Errorf("method delete table: %w", err)
	}

	return nil
}

// Resolve .
func (r *MethodRegistry) Resolve(method DID) (MethodItem, MethodDoc, error) {
	item := MethodItem{}
	exist, err := r.HasMethod(method)
	if err != nil {
		return MethodItem{}, MethodDoc{}, err
	}
	if exist == false {
		return MethodItem{}, MethodDoc{}, fmt.Errorf("resolve method %s not existed", method)
	}

	err = r.table.GetItem(method, &item)
	if err != nil {
		return MethodItem{}, MethodDoc{}, fmt.Errorf("method resolve table get: %w", err)
	}
	doc, err := r.docdb.Get(method, MethodDocType)
	docM := doc.(*MethodDoc)
	if err != nil {
		return item, MethodDoc{}, fmt.Errorf("method resolve docdb get: %w", err)
	}
	return item, *docM, nil
}

// MethodHasAccount checks whether account exists on the method blockchain
func (r *MethodRegistry) MethodHasAccount(method string, account string) {

}

// HasMethod .
func (r *MethodRegistry) HasMethod(method DID) (bool, error) {
	exist, err := r.table.HasItem(method)
	if err != nil {
		return false, fmt.Errorf("has method: %w", err)
	}
	return exist, nil
}

func (r *MethodRegistry) getMethodStatus(method DID) int {
	item := MethodItem{}
	err := r.table.GetItem(method, &item)
	if err != nil {
		// r.logger.Warn("[getMethodStatus] r.table.GetItem err:", err)
		return Initial
	}
	return item.Status
}

// auditStatus .
func (r *MethodRegistry) auditStatus(method DID, status int) error {
	item := &MethodItem{}
	err := r.table.GetItem(method, &item)
	if err != nil {
		return fmt.Errorf("method aduit status table get: %w", err)
	}
	item.Status = status
	err = r.table.UpdateItem(method, item)
	if err != nil {
		return fmt.Errorf("method aduit status table update: %w", err)
	}
	return nil
}
