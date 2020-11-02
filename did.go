package bitxid

import (
	"fmt"
	"strings"

	"github.com/meshplus/bitxhub-kit/storage"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
)

var _ DIDManager = (*DIDRegistry)(nil)
var _ Doc = (*DIDDoc)(nil)

// DIDRegistry for DID Identifier.
// Every Appchain should implements this DID Registry module.
type DIDRegistry struct {
	config *DIDConfig
	table  RegistryTable
	docdb  DocDB
	logger logrus.FieldLogger
	admins []DID // admins of the registry
	method DID   // method of the registry
}

// DIDItem reperesentis a did item.
// Registry table is used together with docdb,
// we suggest to store large data off-chain(in docdb),
// only some frequently used data on-chain(in cache).
type DIDItem struct {
	Identifier DID    // primary key of the item, like a did
	DocAddr    string // addr where the doc file stored
	DocHash    []byte // hash of the doc file
	Status     int    // status of the item
	Cache      []byte // onchain storage part
}

// DIDDoc .
type DIDDoc struct {
	BasicDoc
	Service string `json:"service"`
}

// Marshal .
func (d *DIDDoc) Marshal() ([]byte, error) {
	return Struct2Bytes(d)
}

// Unmarshal .
func (d *DIDDoc) Unmarshal(docBytes []byte) error {
	err := Bytes2Struct(docBytes, &d)
	return err
}

// NewDIDRegistry news a DIDRegistry
func NewDIDRegistry(s1 storage.Storage, s2 storage.Storage, l logrus.FieldLogger, dc *DIDConfig) (*DIDRegistry, error) {
	rt, err := NewKVTable(s1)
	if err != nil {
		return nil, fmt.Errorf("did new table: %w", err)
	}
	db, err := NewKVDocDB(s2)
	if err != nil {
		return nil, fmt.Errorf("did new docdb: %w", err)
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
		return fmt.Errorf("did genesis: admin did not matched with doc")
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
		return "", nil, fmt.Errorf("register did: %w", err)
	}
	if exist == true {
		return "", nil, fmt.Errorf("did %s already existed", did)
	}

	docAddr, err := r.docdb.Create(did, &doc)
	if err != nil {
		return "", nil, fmt.Errorf("register did on docdb: %w", err)
	}
	docBytes, err := doc.Marshal()
	if err != nil {
		return "", nil, fmt.Errorf("register did doc marshal: %w", err)
	}
	docHash := sha3.Sum512(docBytes)

	// update DIDRegistry table:
	err = r.table.CreateItem(did,
		DIDItem{
			Identifier: did,
			Status:     Normal,
			DocAddr:    docAddr,
			DocHash:    docHash[:],
		})
	if err != nil {
		return docAddr, docHash[:], fmt.Errorf("register did on table: %w", err)
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
		return "", nil, fmt.Errorf("did %s not existed", did)
	}
	status := r.getDIDStatus(did)
	if status != Normal {
		return "", nil, fmt.Errorf("can not update did under current status: %d", status)
	}
	docBytes, err := doc.Marshal()
	if err != nil {
		r.logger.Error("update did doc marshal:", err)
		return "", nil, err
	}
	docAddr, err := r.docdb.Update(did, &doc)
	if err != nil {
		return "", nil, fmt.Errorf("update did on docdb: %w", err)
	}
	docHash := sha3.Sum512(docBytes)
	item := DIDItem{}
	err = r.table.GetItem(did, &item)
	if err != nil {
		return docAddr, docHash[:], fmt.Errorf("update did table get: %w", err)
	}
	item.DocAddr = docAddr
	item.DocHash = docHash[:]
	err = r.table.UpdateItem(did, item)
	if err != nil {
		return docAddr, docHash[:], fmt.Errorf("update did on table: %w", err)
	}

	return docAddr, docHash[:], nil
}

// Resolve looks up local-chain to resolve did.
func (r *DIDRegistry) Resolve(did DID) (DIDItem, DIDDoc, error) {
	item := DIDItem{}
	exist, err := r.HasDID(did)
	if err != nil {
		return DIDItem{}, DIDDoc{}, err
	}
	if exist == false {
		return DIDItem{}, DIDDoc{}, fmt.Errorf("did %s not existed", did)
	}

	err = r.table.GetItem(did, &item)
	if err != nil {
		return DIDItem{}, DIDDoc{}, fmt.Errorf("resolve did table get: %w", err)
	}
	doc, err := r.docdb.Get(did, DIDDocType)
	docD := doc.(*DIDDoc)
	if err != nil {
		return item, DIDDoc{}, fmt.Errorf("resolve did docdb get: %w", err)
	}
	return item, *docD, nil
}

// Freeze .
// ATN: only admin should call this.
func (r *DIDRegistry) Freeze(did DID) error {
	exist, err := r.HasDID(did)
	if err != nil {
		return err
	}
	if exist == false {
		return fmt.Errorf("did %s not existed", did)
	}
	err = r.auditStatus(did, Frozen)
	if err != nil {
		return fmt.Errorf("freeze did: %w", err)
	}
	return nil
}

// UnFreeze .
// ATN: only admin should call this.
func (r *DIDRegistry) UnFreeze(did DID) error {
	exist, err := r.HasDID(did)
	if err != nil {
		return err
	}
	if exist == false {
		return fmt.Errorf("did %s not existed", did)
	}
	err = r.auditStatus(did, Normal)
	if err != nil {
		return fmt.Errorf("unfreeze did: %w", err)
	}
	return nil
}

// Delete .
func (r *DIDRegistry) Delete(did DID) error {
	err := r.auditStatus(did, Initial)
	if err != nil {
		return fmt.Errorf("delete did aduit status: %w", err)
	}
	err = r.docdb.Delete(did)
	if err != nil {
		return fmt.Errorf("delete did on docdb: %w", err)
	}
	err = r.table.DeleteItem(did)
	if err != nil {
		return fmt.Errorf("delete did on table: %w", err)
	}
	return nil
}

// HasDID .
func (r *DIDRegistry) HasDID(did DID) (bool, error) {
	exist, err := r.table.HasItem(did)
	if err != nil {
		return false, fmt.Errorf("did has: %w", err)
	}
	return exist, nil
}

func (r *DIDRegistry) getDIDStatus(did DID) int {
	item := DIDItem{}
	err := r.table.GetItem(did, &item)
	if err != nil {
		return Initial
	}
	return item.Status
}

//  caller naturally owns the did ended with his address.
func (r *DIDRegistry) owns(caller string, did DID) bool {
	s := strings.Split(string(did), ":")
	if s[len(s)-1] == caller {
		return true
	}
	return false
}

func (r *DIDRegistry) auditStatus(did DID, status int) error {
	item := &DIDItem{}
	err := r.table.GetItem(did, &item)
	if err != nil {
		return fmt.Errorf("did status get: %w", err)
	}
	item.Status = status
	err = r.table.UpdateItem(did, item)
	if err != nil {
		return fmt.Errorf("did status update: %w", err)
	}
	return nil
}
