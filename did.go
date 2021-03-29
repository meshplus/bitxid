package bitxid

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/meshplus/bitxhub-kit/storage"
	"github.com/sirupsen/logrus"
)

var _ Doc = (*AccountDoc)(nil)

// AccountDoc .
type AccountDoc struct {
	BasicDoc
	Service string `json:"service"`
}

// Marshal .
func (dd *AccountDoc) Marshal() ([]byte, error) {
	return Marshal(dd)
}

// Unmarshal .
func (dd *AccountDoc) Unmarshal(docBytes []byte) error {
	err := Unmarshal(docBytes, &dd)
	return err
}

// GetID .
func (dd *AccountDoc) GetID() DID {
	return dd.ID
}

func (dd *AccountDoc) IsValidFormat() bool {
	if dd.Created == 0 || !dd.ID.IsAccountDIDFormat() {
		return false
	}
	return true
}

var _ TableItem = (*AccountItem)(nil)

// AccountItem reperesentis a did item, element of registry table.
// Registry table is used together with docdb.
type AccountItem struct {
	BasicItem
}

// Marshal .
func (di *AccountItem) Marshal() ([]byte, error) {
	return Marshal(di)
}

// Unmarshal .
func (di *AccountItem) Unmarshal(docBytes []byte) error {
	return Unmarshal(docBytes, &di)
}

// GetID .
func (di *AccountItem) GetID() DID {
	return di.ID
}

var _ AccountDIDManager = (*AccountDIDRegistry)(nil)

// AccountDIDRegistry for DID Identifier,
// Every appchain should use this DID Registry module.
type AccountDIDRegistry struct {
	Mode              RegistryMode  `json:"mode"`
	SelfChainDID      DID           `json:"method"` // method of the registry
	Admins            []DID         `json:"admins"` // admins of the registry
	Table             RegistryTable `json:"table"`
	Docdb             DocDB         `json:"docdb"`
	GenesisAccountDID DID           `json:"genesis_account_did"`
	GenesisAccountDoc DocOption     `json:"genesis_account_doc"`
	logger            logrus.FieldLogger
	// config *DIDConfig
}

// NewAccountDIDRegistry news a AccountDIDRegistry
func NewAccountDIDRegistry(ts storage.Storage, l logrus.FieldLogger, options ...func(*AccountDIDRegistry)) (*AccountDIDRegistry, error) {
	rt, _ := NewKVTable(ts)
	db, _ := NewKVDocDB(nil)
	// doc := genesisAccountDoc()
	ar := &AccountDIDRegistry{
		Mode:   ExternalDocDB,
		Table:  rt,
		Docdb:  db,
		logger: l,
		// Admins:            []DID{doc.GetID()},
		// GenesisAccountDID: doc.GetID(),
		// GenesisAccountDoc: DocOption{
		// ID:      doc.GetID(),
		// Addr:    ".",
		// Hash:    []byte{0},
		// Content: doc,
		// },
	}

	for _, option := range options {
		option(ar)
	}

	return ar, nil
}

// WithAccountDocStorage .
func WithAccountDocStorage(ds storage.Storage) func(*AccountDIDRegistry) {
	return func(ar *AccountDIDRegistry) {
		db, _ := NewKVDocDB(ds)
		ar.Docdb = db
		ar.Mode = InternalDocDB
	}
}

// WithDIDAdmin .
func WithDIDAdmin(a DID) func(*AccountDIDRegistry) {
	return func(ar *AccountDIDRegistry) {
		ar.Admins = []DID{a}
	}
}

// WithGenesisAccountDoc .
func WithGenesisAccountDoc(docOption DocOption) func(*AccountDIDRegistry) {
	return func(ar *AccountDIDRegistry) {
		ar.GenesisAccountDoc = docOption
		ar.GenesisAccountDID = docOption.ID
		if docOption.ID == "" {
			ar.GenesisAccountDID = docOption.Content.(*AccountDoc).ID
		}
	}
}

// SetupGenesis set up genesis to boot the whole did registry
func (r *AccountDIDRegistry) SetupGenesis() error {
	if r.GenesisAccountDID == "" {
		return fmt.Errorf("genesis AccountDID is null")
	}
	if len(r.Admins) == 0 {
		return fmt.Errorf("No admins")
	}
	// if r.GenesisAccountDID != r.GenesisAccountDoc.Content.GetID() {
	// 	return fmt.Errorf("genesis: admin DID not matched with doc")
	// }
	// register genesis did
	var err error
	if r.Mode == ExternalDocDB {
		_, _, err = r.Register(r.GenesisAccountDoc.ID, r.GenesisAccountDoc.Addr, r.GenesisAccountDoc.Hash)
	} else {
		_, _, err = r.RegisterWithDoc(r.GenesisAccountDoc.Content)
		r.SelfChainDID = DID(DID(r.GenesisAccountDoc.Content.GetID()).GetAddress())
	}

	if err != nil {
		return fmt.Errorf("genesis: %w", err)
	}

	r.SelfChainDID = DID(r.GenesisAccountDID.GetChainDID())

	return nil
}

// GetSelfID .
func (r *AccountDIDRegistry) GetSelfID() DID {
	return DID(r.GenesisAccountDID.GetChainDID())
}

// GetAdmins .
func (r *AccountDIDRegistry) GetAdmins() []DID {
	return r.Admins
}

// AddAdmin .
func (r *AccountDIDRegistry) AddAdmin(caller DID) error {
	if r.HasAdmin(caller) {
		return fmt.Errorf("caller %s is already an admin", caller)
	}
	r.Admins = append(r.Admins, caller)
	return nil
}

// RemoveAdmin .
func (r *AccountDIDRegistry) RemoveAdmin(caller DID) error {
	for i, admin := range r.Admins {
		if admin == caller {
			r.Admins = append(r.Admins[:i], r.Admins[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("caller %s is not an admin", caller)
}

// HasAdmin .
func (r *AccountDIDRegistry) HasAdmin(caller DID) bool {
	for _, v := range r.Admins {
		if v == caller {
			return true
		}
	}
	return false
}

// GetChainDID .
func (r *AccountDIDRegistry) GetChainDID() DID {
	return r.SelfChainDID
}

// Register ties did name to a did doc
// ATN: only did who owns did-name should call this
func (r *AccountDIDRegistry) Register(accountDID DID, addr string, hash []byte) (string, []byte, error) {
	return r.updateByStatus(DocOption{ID: accountDID, Addr: addr, Hash: hash}, Initial, Normal)
}

// RegisterWithDoc registers with doc
func (r *AccountDIDRegistry) RegisterWithDoc(doc Doc) (string, []byte, error) {
	return r.updateByStatus(DocOption{Content: doc}, Initial, Normal)
}

// Update updates data about a account did
// ATN: only caller who owns did should call this
func (r *AccountDIDRegistry) Update(accountDID DID, addr string, hash []byte) (string, []byte, error) {
	return r.updateByStatus(DocOption{ID: accountDID, Addr: addr, Hash: hash}, Normal, Normal)
}

// Update with doc
func (r *AccountDIDRegistry) UpdateWithDoc(doc Doc) (string, []byte, error) {
	return r.updateByStatus(DocOption{Content: doc}, Normal, Normal)
}

func (r *AccountDIDRegistry) updateByStatus(docOption DocOption, expectedStatus StatusType, status StatusType) (string, []byte, error) {
	docAddr, docHash, did, err := r.updateDocdbOrNot(docOption, expectedStatus, status)
	if err != nil {
		return "", nil, err
	}

	if expectedStatus == Initial { // register
		err := r.Table.CreateItem(
			&AccountItem{BasicItem{
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
		item, err := r.Table.GetItem(did, AccountDIDTableType)
		if err != nil {
			return docAddr, docHash, fmt.Errorf("DID table get: %w", err)
		}
		itemD := item.(*AccountItem)
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

func (r *AccountDIDRegistry) updateDocdbOrNot(docOption DocOption, expectedStatus StatusType, status StatusType) (string, []byte, DID, error) {
	var docAddr string
	var docHash []byte
	var did DID
	if r.Mode == InternalDocDB {
		doc := docOption.Content.(*AccountDoc)
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
			return "", nil, "", fmt.Errorf("SelfChainDID %s is under status: %s, expectd status: %s", did, status, expectedStatus)
		}
	}
	return docAddr, docHash, did, nil
}

// Freeze .
// ATN: only admin should call this.
func (r *AccountDIDRegistry) Freeze(did DID) error {
	exist := r.HasDID(did)
	if exist == false {
		return fmt.Errorf("DID %s not existed", did)
	}
	return r.auditStatus(did, Frozen)
}

// UnFreeze .
// ATN: only admin should call this.
func (r *AccountDIDRegistry) UnFreeze(did DID) error {
	exist := r.HasDID(did)
	if exist == false {
		return fmt.Errorf("DID %s not existed", did)
	}
	return r.auditStatus(did, Normal)
}

// Resolve looks up local-chain to resolve did.
// @*AccountDoc returns nil if mode is ExternalDocDB
func (r *AccountDIDRegistry) Resolve(did DID) (*AccountItem, *AccountDoc, bool, error) {
	exist := r.HasDID(did)
	if exist == false {
		return nil, nil, false, fmt.Errorf("DID %s not existed", did)
	}

	item, err := r.Table.GetItem(did, AccountDIDTableType)
	if err != nil {
		return nil, nil, false, fmt.Errorf("resolve DID table get: %w", err)
	}
	itemD := item.(*AccountItem)

	if r.Mode == InternalDocDB {
		doc, err := r.Docdb.Get(did, AccountDocType)
		if err != nil {
			return itemD, nil, true, fmt.Errorf("resolve DID docdb get: %w", err)
		}
		docD := doc.(*AccountDoc)
		return itemD, docD, true, nil
	}

	return itemD, nil, true, nil
}

// Delete .
func (r *AccountDIDRegistry) Delete(did DID) error {
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
func (r *AccountDIDRegistry) HasDID(did DID) bool {
	exist := r.Table.HasItem(did)
	return exist
}

func (r *AccountDIDRegistry) getDIDStatus(did DID) StatusType {
	if !r.Table.HasItem(did) {
		return Initial
	}
	item, err := r.Table.GetItem(did, AccountDIDTableType)
	if err != nil {
		r.logger.Error("DID status get item:", err)
		return BadStatus
	}
	itemD := item.(*AccountItem)
	return itemD.Status
}

//  caller naturally owns the did ended with his address.
func (r *AccountDIDRegistry) owns(caller string, did DID) bool {
	s := strings.Split(string(did), ":")
	if s[len(s)-1] == caller {
		return true
	}
	return false
}

func (r *AccountDIDRegistry) auditStatus(did DID, status StatusType) error {
	item, err := r.Table.GetItem(did, AccountDIDTableType)
	if err != nil {
		return fmt.Errorf("DID status get: %w", err)
	}
	itemD := item.(*AccountItem)
	itemD.Status = status
	err = r.Table.UpdateItem(item)
	if err != nil {
		return fmt.Errorf("DID status update: %w", err)
	}
	return nil
}
