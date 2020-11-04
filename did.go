package bitxid

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/meshplus/bitxhub-kit/storage"
	"github.com/sirupsen/logrus"
)

var _ Doc = (*DIDDoc)(nil)

// DIDDoc .
type DIDDoc struct {
	BasicDoc
	Service string `json:"service"`
}

// Marshal .
func (dd *DIDDoc) Marshal() ([]byte, error) {
	return Struct2Bytes(dd)
}

// Unmarshal .
func (dd *DIDDoc) Unmarshal(docBytes []byte) error {
	err := Bytes2Struct(docBytes, &dd)
	return err
}

// GetID .
func (dd *DIDDoc) GetID() DID {
	return dd.ID
}

var _ TableItem = (*DIDItem)(nil)

// DIDItem reperesentis a did item, element of registry table.
// Registry table is used together with docdb.
type DIDItem struct {
	BasicItem
}

// Marshal .
func (di *DIDItem) Marshal() ([]byte, error) {
	return Struct2Bytes(di)
}

// Unmarshal .
func (di *DIDItem) Unmarshal(docBytes []byte) error {
	return Bytes2Struct(docBytes, &di)
}

// GetID .
func (di *DIDItem) GetID() DID {
	return di.ID
}

var _ DIDManager = (*DIDRegistry)(nil)

// DIDRegistry for DID Identifier.
// Every appchain should use this DID Registry module.
type DIDRegistry struct {
	config *DIDConfig
	table  RegistryTable
	docdb  DocDB
	logger logrus.FieldLogger
	admins []DID // admins of the registry
	method DID   // method of the registry
}

// NewDIDRegistry news a DIDRegistry
func NewDIDRegistry(ts storage.Storage, ds storage.Storage, l logrus.FieldLogger, dc *DIDConfig) (*DIDRegistry, error) {
	rt, err := NewKVTable(ts)
	if err != nil {
		return nil, fmt.Errorf("DID new table: %w", err)
	}
	db, err := NewKVDocDB(ds)
	if err != nil {
		return nil, fmt.Errorf("DID new docdb: %w", err)
	}
	return &DIDRegistry{
		config: dc,
		table:  rt,
		docdb:  db,
		logger: l,
		admins: []DID{""},
	}, nil
}

// SetupGenesis set up genesis to boot the whole did registry
func (r *DIDRegistry) SetupGenesis() error {
	if r.config.Admin != r.config.AdminDoc.ID {
		return fmt.Errorf("DID genesis: admin DID not matched with doc")
	}
	// register genesis did
	_, _, err := r.Register(r.config.AdminDoc)
	if err != nil {
		return fmt.Errorf("genesis: %w", err)
	}
	r.admins = append(r.admins, DID(r.config.Admin))
	r.method = DID(DID(r.config.AdminDoc.ID).GetAddress())

	return nil
}

// GetAdmins .
func (r *DIDRegistry) GetAdmins() []DID {
	return r.admins
}

// AddAdmin .
func (r *DIDRegistry) AddAdmin(caller DID) error {
	if r.HasAdmin(caller) {
		return fmt.Errorf("caller %s is already an admin", caller)
	}
	r.admins = append(r.admins, caller)
	return nil
}

// HasAdmin .
func (r *DIDRegistry) HasAdmin(caller DID) bool {
	for _, v := range r.admins {
		if v == caller {
			return true
		}
	}
	return false
}

// GetMethod .
func (r *DIDRegistry) GetMethod() DID {
	return r.method
}

// Register ties did name to a did doc
// ATN: only did who owns did-name should call this
func (r *DIDRegistry) Register(doc DIDDoc) (string, []byte, error) {
	did := DID(doc.ID)
	exist, err := r.HasDID(did)
	if err != nil {
		return "", nil, fmt.Errorf("register DID: %w", err)
	}
	if exist == true {
		return "", nil, fmt.Errorf("DID %s already existed", did)
	}

	docAddr, err := r.docdb.Create(&doc)
	if err != nil {
		return "", nil, fmt.Errorf("register DID on docdb: %w", err)
	}
	docBytes, err := doc.Marshal()
	if err != nil {
		return "", nil, fmt.Errorf("register DID doc marshal: %w", err)
	}
	docHash := sha256.Sum256(docBytes)
	// update DIDRegistry table:
	err = r.table.CreateItem(
		&DIDItem{BasicItem{
			ID:      did,
			Status:  Normal,
			DocAddr: docAddr,
			DocHash: docHash[:],
		},
		})
	if err != nil {
		return docAddr, docHash[:], fmt.Errorf("register DID on table: %w", err)
	}

	return docAddr, docHash[:], nil
}

// Update .
// ATN: only caller who owns did should call this
func (r *DIDRegistry) Update(doc DIDDoc) (string, []byte, error) {
	did := DID(doc.ID)
	// check exist
	exist, err := r.HasDID(did)
	if err != nil {
		return "", nil, err
	}
	if exist == false {
		return "", nil, fmt.Errorf("DID %s not existed", did)
	}
	status := r.getDIDStatus(did)
	if status != Normal {
		return "", nil, fmt.Errorf("can not update DID under current status: %d", status)
	}
	docBytes, err := doc.Marshal()
	if err != nil {
		r.logger.Error("update DID doc marshal:", err)
		return "", nil, err
	}
	docAddr, err := r.docdb.Update(&doc)
	if err != nil {
		return "", nil, fmt.Errorf("update DID on docdb: %w", err)
	}
	docHash := sha256.Sum256(docBytes)
	item, err := r.table.GetItem(did, DIDTableType)
	if err != nil {
		return docAddr, docHash[:], fmt.Errorf("update DID table get: %w", err)
	}
	itemD := item.(*DIDItem)
	itemD.DocAddr = docAddr
	itemD.DocHash = docHash[:]
	err = r.table.UpdateItem(itemD)
	if err != nil {
		return docAddr, docHash[:], fmt.Errorf("update DID on table: %w", err)
	}

	return docAddr, docHash[:], nil
}

// Resolve looks up local-chain to resolve did.
func (r *DIDRegistry) Resolve(did DID) (DIDItem, DIDDoc, error) {
	exist, err := r.HasDID(did)
	if err != nil {
		return DIDItem{}, DIDDoc{}, err
	}
	if exist == false {
		return DIDItem{}, DIDDoc{}, fmt.Errorf("DID %s not existed", did)
	}

	item, err := r.table.GetItem(did, DIDTableType)
	if err != nil {
		return DIDItem{}, DIDDoc{}, fmt.Errorf("resolve DID table get: %w", err)
	}
	itemD := item.(*DIDItem)
	doc, err := r.docdb.Get(did, DIDDocType)
	if err != nil {
		return *itemD, DIDDoc{}, fmt.Errorf("resolve DID docdb get: %w", err)
	}
	docD := doc.(*DIDDoc)
	return *itemD, *docD, nil
}

// Freeze .
// ATN: only admin should call this.
func (r *DIDRegistry) Freeze(did DID) error {
	exist, err := r.HasDID(did)
	if err != nil {
		return err
	}
	if exist == false {
		return fmt.Errorf("DID %s not existed", did)
	}
	return r.auditStatus(did, Frozen)
}

// UnFreeze .
// ATN: only admin should call this.
func (r *DIDRegistry) UnFreeze(did DID) error {
	exist, err := r.HasDID(did)
	if err != nil {
		return err
	}
	if exist == false {
		return fmt.Errorf("DID %s not existed", did)
	}
	return r.auditStatus(did, Normal)
}

// Delete .
func (r *DIDRegistry) Delete(did DID) error {
	err := r.auditStatus(did, Initial)
	if err != nil {
		return fmt.Errorf("delete DID aduit status: %w", err)
	}
	err = r.docdb.Delete(did)
	if err != nil {
		return fmt.Errorf("delete DID on docdb: %w", err)
	}
	err = r.table.DeleteItem(did)
	if err != nil {
		return fmt.Errorf("delete DID on table: %w", err)
	}
	return nil
}

// HasDID .
func (r *DIDRegistry) HasDID(did DID) (bool, error) {
	exist, err := r.table.HasItem(did)
	if err != nil {
		return false, fmt.Errorf("DID has: %w", err)
	}
	return exist, nil
}

func (r *DIDRegistry) getDIDStatus(did DID) StatusType {
	item, err := r.table.GetItem(did, DIDTableType)
	if err != nil {
		r.logger.Warn("DID status get item:", err)
		return Initial
	}
	itemD := item.(*DIDItem)
	return itemD.Status
}

//  caller naturally owns the did ended with his address.
func (r *DIDRegistry) owns(caller string, did DID) bool {
	s := strings.Split(string(did), ":")
	if s[len(s)-1] == caller {
		return true
	}
	return false
}

func (r *DIDRegistry) auditStatus(did DID, status StatusType) error {
	item, err := r.table.GetItem(did, DIDTableType)
	if err != nil {
		return fmt.Errorf("DID status get: %w", err)
	}
	itemD := item.(*DIDItem)
	itemD.Status = status
	err = r.table.UpdateItem(item)
	if err != nil {
		return fmt.Errorf("DID status update: %w", err)
	}
	return nil
}
