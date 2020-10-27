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
		l.Error("[New] NewTable err", err)
		return nil, err
	}
	db, err := NewKVDocDB(s2)
	if err != nil {
		l.Error("[New] NewDB err", err)
		return nil, err
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
		return fmt.Errorf("Admin DID not matched with DID Document")
	}
	// register genesis did:
	_, _, err := r.Register(r.config.AdminDoc)
	if err != nil {
		r.logger.Error("[SetupGenesis] register admin fail:", err)
		return err
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
		return fmt.Errorf("caller %s is already the admin", caller)
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
		r.logger.Error("[Register] r.HasDID err:", err)
		return "", nil, err
	}
	if exist == true {
		return "", nil, fmt.Errorf("[Register] The DID Already existed")
	}

	docAddr, err := r.docdb.Create(did, &doc)
	if err != nil {
		r.logger.Error("[Register] r.docdb.Create err:", err)
		return "", nil, err
	}
	docBytes, err := doc.Marshal()
	if err != nil {
		r.logger.Error("[Register] doc.Marshal err:", err)
		return "", nil, err
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
		r.logger.Error("[Apply] r.table.CreateItem err:", err)
		return docAddr, docHash[:], err
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
		r.logger.Error("[Update] r.HasDID err:", err)
		return "", nil, err
	}
	if exist == false {
		return "", nil, fmt.Errorf("[Update] The DID NOT existed")
	}
	status := r.getDIDStatus(did)
	if status != Normal {
		return "", nil, fmt.Errorf("[Update] Can not Update for current status: %d", status)
	}
	docBytes, err := doc.Marshal()
	if err != nil {
		r.logger.Error("[Register] doc.Marshal err:", err)
		return "", nil, err
	}
	docAddr, err := r.docdb.Update(did, &doc)
	if err != nil {
		r.logger.Error("[Update] r.docdb.Update err:", err)
		return "", nil, err
	}
	docHash := sha3.Sum512(docBytes)
	item := DIDItem{}
	err = r.table.GetItem(did, &item)
	if err != nil {
		r.logger.Error("[Update] r.table.GetItem err:", err)
		return docAddr, docHash[:], err
	}
	item.DocAddr = docAddr
	item.DocHash = docHash[:]
	err = r.table.UpdateItem(did, item)
	if err != nil {
		r.logger.Error("[Update] r.table.UpdateItem err:", err)
		return docAddr, docHash[:], err
	}

	return docAddr, docHash[:], nil
}

// Resolve looks up local-chain to resolve did.
func (r *DIDRegistry) Resolve(did DID) (DIDItem, DIDDoc, error) {
	item := DIDItem{}
	exist, err := r.HasDID(did)
	if err != nil {
		r.logger.Error("[Resolve] r.HasDID err:", err)
		return DIDItem{}, DIDDoc{}, err
	}
	if exist == false {
		return DIDItem{}, DIDDoc{}, fmt.Errorf("[Resolve] The Method NOT existed")
	}

	err = r.table.GetItem(did, &item)
	if err != nil {
		r.logger.Error("[Resolve] r.table.GetItem err:", err)
		return DIDItem{}, DIDDoc{}, err
	}
	doc, err := r.docdb.Get(did, DIDDocType)
	docD := doc.(*DIDDoc)
	if err != nil {
		r.logger.Error("[Resolve] r.docdb.Get err:", err)
		return item, DIDDoc{}, err
	}
	return item, *docD, nil
}

// Freeze .
// ATN: only admin should call this.
func (r *DIDRegistry) Freeze(did DID) error {
	exist, err := r.HasDID(did)
	if err != nil {
		r.logger.Error("[Freeze] r.HasMethod err:", err)
		return err
	}
	if exist == false {
		return fmt.Errorf("[Freeze] The Method NOT existed")
	}
	err = r.auditStatus(did, Frozen)

	return nil
}

// UnFreeze .
// ATN: only admin should call this.
func (r *DIDRegistry) UnFreeze(did DID) error {
	exist, err := r.HasDID(did)
	if err != nil {
		r.logger.Error("[UnFreeze] r.HasDID err:", err)
		return err
	}
	if exist == false {
		return fmt.Errorf("[UnFreeze] The DID NOT existed")
	}
	err = r.auditStatus(did, Normal)

	return nil
}

// Delete .
func (r *DIDRegistry) Delete(did DID) error {
	err := r.auditStatus(did, Initial)
	if err != nil {
		r.logger.Error("[Delete] r.auditStatus err:", err)
		return err
	}

	err = r.docdb.Delete(did)
	if err != nil {
		r.logger.Error("[Delete] r.docdb.Delete err:", err)
		return err
	}
	err = r.table.DeleteItem(did)
	if err != nil {
		r.logger.Error("[Delete] r.table.DeleteItem err:", err)
		return err
	}

	return nil
}

// HasDID .
func (r *DIDRegistry) HasDID(did DID) (bool, error) {
	exist, err := r.table.HasItem(did)
	if err != nil {
		r.logger.Error("[HasDID] r.table.HasItem err:", err)
		return false, err
	}
	return exist, err
}

func (r *DIDRegistry) getDIDStatus(did DID) int {
	item := DIDItem{}
	err := r.table.GetItem(did, &item)
	if err != nil {
		r.logger.Error("[getDIDStatus] r.table.GetItem err:", err)
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
		r.logger.Error("[auditStatus] r.table.GetItem err:", err)
		return err
	}
	item.Status = status
	err = r.table.UpdateItem(did, item)
	if err != nil {
		r.logger.Error("[auditStatus] r.table.UpdateItem err:", err)
		return err
	}
	return nil
}
