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

// DIDRegistry for DID Identifier,
// Every appchain should use this DID Registry module.
type DIDRegistry struct {
	mode       RegistryMode
	method     DID   // method of the registry
	admins     []DID // admins of the registry
	table      RegistryTable
	docdb      DocDB
	genesisDID DID
	genesisDoc DocOption
	logger     logrus.FieldLogger
	// config *DIDConfig
}

// NewDIDRegistry news a DIDRegistry
func NewDIDRegistry(ts storage.Storage, l logrus.FieldLogger, options ...func(*DIDRegistry)) (*DIDRegistry, error) {
	rt, _ := NewKVTable(ts)
	db, _ := NewKVDocDB(nil)
	doc := genesisDIDDoc()
	dr := &DIDRegistry{
		mode:       ExternalDocDB,
		table:      rt,
		docdb:      db,
		logger:     l,
		admins:     []DID{doc.GetID()},
		genesisDID: doc.GetID(),
		genesisDoc: DocOption{
			ID:      doc.GetID(),
			Addr:    ".",
			Hash:    []byte{0},
			Content: doc,
		},
	}

	for _, option := range options {
		option(dr)
	}

	return dr, nil
}

// WithDIDDocStorage .
func WithDIDDocStorage(ds storage.Storage) func(*DIDRegistry) {
	return func(dr *DIDRegistry) {
		db, _ := NewKVDocDB(ds)
		dr.docdb = db
		dr.mode = InternalDocDB
	}
}

// WithDIDAdmin .
func WithDIDAdmin(a DID) func(*DIDRegistry) {
	return func(dr *DIDRegistry) {
		dr.admins = []DID{a}
	}
}

// WithGenesisDID .
func WithGenesisDID(d DID) func(*DIDRegistry) {
	return func(dr *DIDRegistry) {
		dr.genesisDID = d
	}
}

// WithGenesisDIDDoc .
func WithGenesisDIDDoc(docOption DocOption) func(*DIDRegistry) {
	return func(dr *DIDRegistry) {
		dr.genesisDoc = docOption
	}
}

// SetupGenesis set up genesis to boot the whole did registry
func (r *DIDRegistry) SetupGenesis() error {
	if r.genesisDID != r.genesisDoc.Content.GetID() {
		return fmt.Errorf("genesis: admin DID not matched with doc")
	}
	// register genesis did
	_, _, err := r.Register(r.genesisDoc)
	if err != nil {
		return fmt.Errorf("genesis: %w", err)
	}

	r.method = DID(DID(r.genesisDoc.Content.GetID()).GetAddress())

	return nil
}

// GetSelfID .
func (r *DIDRegistry) GetSelfID() DID {
	return DID(r.genesisDID.GetMethod())
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
func (r *DIDRegistry) Register(docOption DocOption) (string, []byte, error) {
	return r.updateByStatus(docOption, Initial, Normal)
}

// Update .
// ATN: only caller who owns did should call this
func (r *DIDRegistry) Update(docOption DocOption) (string, []byte, error) {
	return r.updateByStatus(docOption, Normal, Normal)
}

func (r *DIDRegistry) updateByStatus(docOption DocOption, expectedStatus StatusType, status StatusType) (string, []byte, error) {
	docAddr, docHash, did, err := r.updateDocdbOrNot(docOption, expectedStatus, status)
	if err != nil {
		return "", nil, err
	}

	if expectedStatus == Initial { // register
		err := r.table.CreateItem(
			&DIDItem{BasicItem{
				ID:      did,
				Status:  Normal,
				DocAddr: docAddr,
				DocHash: docHash,
			},
			})
		if err != nil {
			return docAddr, docHash, fmt.Errorf("register DID on table: %w", err)
		}
	} else { // update
		item, err := r.table.GetItem(did, DIDTableType)
		if err != nil {
			return docAddr, docHash, fmt.Errorf("DID table get: %w", err)
		}
		itemD := item.(*DIDItem)
		itemD.DocAddr = docAddr
		itemD.DocHash = docHash
		itemD.Status = status
		err = r.table.UpdateItem(itemD)
		if err != nil {
			return docAddr, docHash, fmt.Errorf("update DID on table: %w", err)
		}
	}

	return docAddr, docHash, nil
}

func (r *DIDRegistry) updateDocdbOrNot(docOption DocOption, expectedStatus StatusType, status StatusType) (string, []byte, DID, error) {
	var docAddr string
	var docHash []byte
	var did DID
	if r.mode == InternalDocDB {
		doc := docOption.Content.(*DIDDoc)
		did = doc.GetID()

		// check exist
		exist := r.HasDID(did)
		if expectedStatus == Initial && exist == true {
			return "", nil, "", fmt.Errorf("DID %s already existed", did)
		} else if expectedStatus == Normal && exist == false {
			return "", nil, "", fmt.Errorf("DID %s not existed", did)
		}

		status := r.getDIDStatus(did)
		if status != expectedStatus {
			return "", nil, "", fmt.Errorf("DID %s is under status: %s, expectd status: %s", did, status, expectedStatus)
		}

		docBytes, err := doc.Marshal()
		if err != nil {
			r.logger.Error("DID doc marshal:", err)
			return "", nil, "", err
		}

		if expectedStatus == Initial { // register
			docAddr, err = r.docdb.Create(doc)
			if err != nil {
				return "", nil, "", fmt.Errorf("register DID on docdb: %w", err)
			}
		} else { // update
			docAddr, err = r.docdb.Update(doc)
			if err != nil {
				return "", nil, "", fmt.Errorf("update DID on docdb: %w", err)
			}
		}

		docHash32 := sha256.Sum256(docBytes)
		docHash = docHash32[:]
	} else {
		did = docOption.ID
		docAddr = docOption.Addr
		docHash = docOption.Hash
		status := r.getDIDStatus(did)
		if status != expectedStatus {
			return "", nil, "", fmt.Errorf("Method %s is under status: %s, expectd status: %s", did, status, expectedStatus)
		}
	}
	return docAddr, docHash, did, nil
}

// Resolve looks up local-chain to resolve did.
// @*DIDDoc returns nil if mode is ExternalDocDB
func (r *DIDRegistry) Resolve(did DID) (*DIDItem, *DIDDoc, bool, error) {
	exist := r.HasDID(did)
	if exist == false {
		return nil, nil, false, fmt.Errorf("DID %s not existed", did)
	}

	item, err := r.table.GetItem(did, DIDTableType)
	if err != nil {
		return nil, nil, false, fmt.Errorf("resolve DID table get: %w", err)
	}
	itemD := item.(*DIDItem)

	if r.mode == InternalDocDB {
		doc, err := r.docdb.Get(did, DIDDocType)
		if err != nil {
			return itemD, nil, true, fmt.Errorf("resolve DID docdb get: %w", err)
		}
		docD := doc.(*DIDDoc)
		return itemD, docD, true, nil
	}

	return itemD, nil, true, nil
}

// Freeze .
// ATN: only admin should call this.
func (r *DIDRegistry) Freeze(did DID) error {
	exist := r.HasDID(did)
	if exist == false {
		return fmt.Errorf("DID %s not existed", did)
	}
	return r.auditStatus(did, Frozen)
}

// UnFreeze .
// ATN: only admin should call this.
func (r *DIDRegistry) UnFreeze(did DID) error {
	exist := r.HasDID(did)
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
	r.table.DeleteItem(did)
	if r.mode == InternalDocDB {
		r.docdb.Delete(did)
	}
	return nil
}

// HasDID .
func (r *DIDRegistry) HasDID(did DID) bool {
	exist := r.table.HasItem(did)
	return exist
}

func (r *DIDRegistry) getDIDStatus(did DID) StatusType {
	if !r.table.HasItem(did) {
		return Initial
	}
	item, err := r.table.GetItem(did, DIDTableType)
	if err != nil {
		r.logger.Error("DID status get item:", err)
		return BadStatus
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
