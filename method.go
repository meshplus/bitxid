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

// NewMethodRegistry news a MethodRegistry
func NewMethodRegistry(s1 storage.Storage, s2 storage.Storage, l logrus.FieldLogger, mc *MethodConfig) (*MethodRegistry, error) {
	rt, err := NewKVTable(s1)
	if err != nil {
		l.Error("[New] NewTable err", err)
		return nil, err
	}
	db, err := NewKVDocDB(s2)
	if err != nil {
		l.Error("[New] docdNewDB err", err)
		return nil, err
	}
	return &MethodRegistry{
		table:  rt,
		docdb:  db,
		logger: l,
		config: mc,
		admins: []DID{""},
	}, nil
}

// NewMethodRegistry .
func newMethodRegistry(rt RegistryTable, db DocDB, l logrus.FieldLogger, mc *MethodConfig) (*MethodRegistry, error) {
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
		return fmt.Errorf("[SetupGenesis] This method registry is not a relay root, check the config")
	}
	if r.config.GenesisMetohd != r.config.GenesisDoc.ID {
		return fmt.Errorf("Method not matched with Method Document")
	}
	// register method
	r.Apply(DID(r.config.Admin), DID(r.config.GenesisMetohd))
	r.AuditApply(DID(r.config.GenesisMetohd), true)
	_, _, err := r.Register(r.config.GenesisDoc)
	if err != nil {
		r.logger.Error("[SetupGenesis] r.Register err:", err)
		return err
	}
	// add admins did:
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
		return fmt.Errorf("caller %s is already the admin", caller)
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
		return fmt.Errorf("[Apply] The Method name is not standard")
	}
	// check if Method exists
	exist, err := r.HasMethod(method)
	if err != nil {
		r.logger.Error("[Apply] r.HasMethod err:", err)
		return err
	}
	if exist == true {
		return fmt.Errorf("[Apply] The Method is ALREADY existed")
	}

	status := r.getMethodStatus(method)
	if status != Initial {
		return fmt.Errorf("[Apply] Can not Register for current status: %d", status)
	}
	// creates item in table
	err = r.table.CreateItem([]byte(method),
		MethodItem{
			Method: DID(method),
			Status: ApplyAudit,
			Owner:  caller,
		})
	if err != nil {
		r.logger.Error("[Apply] r.table.CreateItem err:", err)
		return err
	}
	return nil
}

// AuditApply .
// ATNS: only admin should call this.
func (r *MethodRegistry) AuditApply(method DID, result bool) error {
	exist, err := r.HasMethod(method)
	if err != nil {
		r.logger.Error("[AuditApply] r.HasMethod err:", err)
		return err
	}
	if exist == false {
		return fmt.Errorf("[AuditApply] The Method NOT existed")
	}
	status := r.getMethodStatus(method)
	if !(status == ApplyAudit || status == ApplyFailed) {
		return fmt.Errorf("[AuditApply] Can not AuditApply for current status: %d", status)
	}
	if result {
		err = r.auditStatus(method, ApplySuccess)
	} else {
		err = r.auditStatus(method, ApplyFailed)
	}
	if err != nil {
		r.logger.Error("[AuditApply] r.auditStatus err:", err)
		return err
	}
	return nil
}

// Register ties method name to a method doc
// ATN: only did who owns method-name should call this
func (r *MethodRegistry) Register(doc MethodDoc) (string, []byte, error) {
	method := DID(doc.ID)
	exist, err := r.HasMethod(method)
	if err != nil {
		r.logger.Error("[Register] r.HasMethod err:", err)
		return "", nil, err
	}
	if exist == false {
		return "", nil, fmt.Errorf("[Register] The Method NOT existed")
	}
	status := r.getMethodStatus(method)
	if status != ApplySuccess {
		return "", nil, fmt.Errorf("[Register] Can not Register for current status: %d", status)
	}

	docBytes, err := Struct2Bytes(doc)
	if err != nil {
		r.logger.Error("[Register] Struct2Bytes err:", err)
		return "", nil, err
	}

	docAddr, err := r.docdb.Create([]byte(method), docBytes)
	if err != nil {
		r.logger.Error("[Register] r.docdb.Create err:", err)
		return "", nil, err
	}
	docHash := sha3.Sum512(docBytes) // docHash := sha256.Sum256(doc)
	// update MethodRegistry table:
	item := &MethodItem{}
	err = r.table.GetItem([]byte(method), &item)
	if err != nil {
		r.logger.Error("[Register] r.table.GetItem err:", err)
		return "", nil, err
	}
	item.Status = Normal
	item.DocAddr = docAddr
	item.DocHash = docHash[:]
	err = r.table.UpdateItem([]byte(method), item)
	if err != nil {
		r.logger.Error("[Register] r.table.UpdateItem err:", err)
		return docAddr, item.DocHash, err
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
		r.logger.Error("[Update] r.HasMethod err:", err)
		return "", nil, err
	}
	if exist == false {
		return "", nil, fmt.Errorf("[Update] The Method NOT existed")
	}
	status := r.getMethodStatus(method)
	if status != Normal {
		return "", nil, fmt.Errorf("[Update] Can not Update for current status: %d", status)
	}

	docBytes, err := Struct2Bytes(doc)
	if err != nil {
		r.logger.Error("[Register] Struct2Bytes err:", err)
		return "", nil, err
	}

	docAddr, err := r.docdb.Update([]byte(method), docBytes)
	if err != nil {
		r.logger.Error("[Update] r.docdb.Update err:", err)
		return "", nil, err
	}
	docHash := sha3.Sum512(docBytes)

	item := MethodItem{}
	err = r.table.GetItem([]byte(method), &item)
	if err != nil {
		r.logger.Error("[Update] r.table.GetItem err:", err)
		return docAddr, docHash[:], err
	}
	item.DocAddr = docAddr
	item.DocHash = docHash[:]
	err = r.table.UpdateItem([]byte(method), item)

	if err != nil {
		r.logger.Error("[Update] r.table.UpdateItem err:", err)
		return docAddr, docHash[:], err
	}

	return docAddr, docHash[:], nil
}

// Audit .
// ATN: only admin should call this.
func (r *MethodRegistry) Audit(method DID, status int) error {
	exist, err := r.HasMethod(method)
	if err != nil {
		r.logger.Error("[Audit] r.HasMethod err:", err)
		return err
	}
	if exist == false {
		return fmt.Errorf("[Audit] The Method NOT existed")
	}
	err = r.auditStatus(method, status)

	return nil
}

// Freeze .
// ATN: only admdin should call this.
func (r *MethodRegistry) Freeze(method DID) error {
	exist, err := r.HasMethod(method)
	if err != nil {
		r.logger.Error("[Freeze] r.HasMethod err:", err)
		return err
	}
	if exist == false {
		return fmt.Errorf("[Freeze] The Method NOT existed")
	}
	err = r.auditStatus(method, Frozen)

	return nil
}

// UnFreeze .
// ATN: only admdin should call this.
func (r *MethodRegistry) UnFreeze(method DID) error {
	exist, err := r.HasMethod(method)
	if err != nil {
		r.logger.Error("[UnFreeze] r.HasMethod err:", err)
		return err
	}
	if exist == false {
		return fmt.Errorf("[UnFreeze] The Method NOT existed")
	}

	err = r.auditStatus(method, Normal)

	return nil
}

// Delete .
func (r *MethodRegistry) Delete(method DID) error {
	err := r.auditStatus(method, Initial)
	if err != nil {
		r.logger.Error("[Delete] r.auditStatus err:", err)
		return err
	}

	err = r.docdb.Delete([]byte(method))
	if err != nil {
		r.logger.Error("[Delete] r.docdb.Delete err:", err)
		return err
	}
	err = r.table.DeleteItem([]byte(method))
	if err != nil {
		r.logger.Error("[Delete] r.table.DeleteItem err:", err)
		return err
	}

	return nil
}

// Resolve .
func (r *MethodRegistry) Resolve(method DID) (MethodItem, MethodDoc, error) {
	item := MethodItem{}
	exist, err := r.HasMethod(method)
	if err != nil {
		r.logger.Error("[Resolve] r.HasMethod err:", err)
		return MethodItem{}, MethodDoc{}, err
	}
	if exist == false {
		return MethodItem{}, MethodDoc{}, fmt.Errorf("[Resolve] The Method NOT existed")
	}

	err = r.table.GetItem([]byte(method), &item)
	if err != nil {
		r.logger.Error("[Resolve] r.table.GetItem err:", err)
		return MethodItem{}, MethodDoc{}, err
	}
	doc, err := r.docdb.Get([]byte(method))
	if err != nil {
		r.logger.Error("[Resolve] r.docdb.Get err:", err)
		return item, MethodDoc{}, err
	}
	docStruct := MethodDoc{}
	Bytes2Struct(doc, &docStruct)
	return item, docStruct, nil
}

// MethodHasAccount checks whether account exists on the method blockchain
func (r *MethodRegistry) MethodHasAccount(method string, account string) {

}

// HasMethod .
func (r *MethodRegistry) HasMethod(method DID) (bool, error) {
	exist, err := r.table.HasItem([]byte(method))
	if err != nil {
		r.logger.Error("[HasMethod] r.table.HasItem err:", err)
		return false, err
	}
	return exist, err
}

func (r *MethodRegistry) getMethodStatus(method DID) int {
	item := MethodItem{}
	err := r.table.GetItem([]byte(method), &item)
	if err != nil {
		r.logger.Warn("[getMethodStatus] r.table.GetItem err:", err)
		return Initial
	}
	return item.Status
}

// auditStatus .
func (r *MethodRegistry) auditStatus(method DID, status int) error {
	item := &MethodItem{}
	err := r.table.GetItem([]byte(method), &item)
	if err != nil {
		r.logger.Error("[auditStatus] r.table.GetItem err:", err)
		return err
	}
	item.Status = status
	err = r.table.UpdateItem([]byte(method), item)
	if err != nil {
		r.logger.Error("[auditStatus] r.table.UpdateItem err:", err)
		return err
	}
	return nil
}
