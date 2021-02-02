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
	Mode       RegistryMode  `json:"mode"`
	Method     DID           `json:"method"` // method of the registry
	Admins     []DID         `json:"admins"` // admins of the registry
	Table      RegistryTable `json:"table"`
	Docdb      DocDB         `json:"docdb"`
	GenesisDID DID           `json:"genesis_did"`
	GenesisDoc DocOption     `json:"genesis_doc"`
	logger     logrus.FieldLogger
	// config *DIDConfig
}

// NewDIDRegistry news a DIDRegistry
func NewDIDRegistry(ts storage.Storage, l logrus.FieldLogger, options ...func(*DIDRegistry)) (*DIDRegistry, error) {
	rt, _ := NewKVTable(ts)
	db, _ := NewKVDocDB(nil)
	doc := genesisDIDDoc()
	dr := &DIDRegistry{
		Mode:       ExternalDocDB,
		Table:      rt,
		Docdb:      db,
		logger:     l,
		Admins:     []DID{doc.GetID()},
		GenesisDID: doc.GetID(),
		GenesisDoc: DocOption{
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
		dr.Docdb = db
		dr.Mode = InternalDocDB
	}
}

// WithDIDAdmin .
func WithDIDAdmin(a DID) func(*DIDRegistry) {
	return func(dr *DIDRegistry) {
		dr.Admins = []DID{a}
	}
}

// WithGenesisDID .
func WithGenesisDID(d DID) func(*DIDRegistry) {
	return func(dr *DIDRegistry) {
		dr.GenesisDID = d
	}
}

// WithGenesisDIDDoc .
func WithGenesisDIDDoc(docOption DocOption) func(*DIDRegistry) {
	return func(dr *DIDRegistry) {
		dr.GenesisDoc = docOption
	}
}

// SetupGenesis set up genesis to boot the whole did registry
func (r *DIDRegistry) SetupGenesis() error {
	if r.GenesisDID != r.GenesisDoc.Content.GetID() {
		return fmt.Errorf("genesis: admin DID not matched with doc")
	}
	// register genesis did
	_, _, err := r.Register(r.GenesisDoc)
	if err != nil {
		return fmt.Errorf("genesis: %w", err)
	}

	r.Method = DID(DID(r.GenesisDoc.Content.GetID()).GetAddress())

	return nil
}

// GetSelfID .
func (r *DIDRegistry) GetSelfID() DID {
	return DID(r.GenesisDID.GetMethod())
}

// GetAdmins .
func (r *DIDRegistry) GetAdmins() []DID {
	return r.Admins
}

// AddAdmin .
func (r *DIDRegistry) AddAdmin(caller DID) error {
	if r.HasAdmin(caller) {
		return fmt.Errorf("caller %s is already an admin", caller)
	}
	r.Admins = append(r.Admins, caller)
	return nil
}

// RemoveAdmin .
func (r *DIDRegistry) RemoveAdmin(caller DID) error {
	for i, admin := range r.Admins {
		if admin == caller {
			r.Admins = append(r.Admins[:i], r.Admins[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("caller %s is not an admin", caller)
}

// HasAdmin .
func (r *DIDRegistry) HasAdmin(caller DID) bool {
	for _, v := range r.Admins {
		if v == caller {
			return true
		}
	}
	return false
}

// GetMethod .
func (r *DIDRegistry) GetMethod() DID {
	return r.Method
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
		err := r.Table.CreateItem(
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
		item, err := r.Table.GetItem(did, DIDTableType)
		if err != nil {
			return docAddr, docHash, fmt.Errorf("DID table get: %w", err)
		}
		itemD := item.(*DIDItem)
		itemD.DocAddr = docAddr
		itemD.DocHash = docHash
		itemD.Status = status
		err = r.Table.UpdateItem(itemD)
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
	if r.Mode == InternalDocDB {
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
			docAddr, err = r.Docdb.Create(doc)
			if err != nil {
				return "", nil, "", fmt.Errorf("register DID on docdb: %w", err)
			}
		} else { // update
			docAddr, err = r.Docdb.Update(doc)
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

	item, err := r.Table.GetItem(did, DIDTableType)
	if err != nil {
		return nil, nil, false, fmt.Errorf("resolve DID table get: %w", err)
	}
	itemD := item.(*DIDItem)

	if r.Mode == InternalDocDB {
		doc, err := r.Docdb.Get(did, DIDDocType)
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
	r.Table.DeleteItem(did)
	if r.Mode == InternalDocDB {
		r.Docdb.Delete(did)
	}
	return nil
}

// HasDID .
func (r *DIDRegistry) HasDID(did DID) bool {
	exist := r.Table.HasItem(did)
	return exist
}

func (r *DIDRegistry) getDIDStatus(did DID) StatusType {
	if !r.Table.HasItem(did) {
		return Initial
	}
	item, err := r.Table.GetItem(did, DIDTableType)
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
	item, err := r.Table.GetItem(did, DIDTableType)
	if err != nil {
		return fmt.Errorf("DID status get: %w", err)
	}
	itemD := item.(*DIDItem)
	itemD.Status = status
	err = r.Table.UpdateItem(item)
	if err != nil {
		return fmt.Errorf("DID status update: %w", err)
	}
	return nil
}
