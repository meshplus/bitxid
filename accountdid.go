package bitxid

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/meshplus/bitxhub-kit/storage"
	"github.com/sirupsen/logrus"
)

var _ Doc = (*AccountDoc)(nil)

// AccountDoc represents account identity information
type AccountDoc struct {
	BasicDoc
	Service string `json:"service"`
}

// Marshal marshals account doc
func (dd *AccountDoc) Marshal() ([]byte, error) {
	return Marshal(dd)
}

// Unmarshal unmarshals account doc
func (dd *AccountDoc) Unmarshal(docBytes []byte) error {
	err := Unmarshal(docBytes, &dd)
	return err
}

// GetID gets id of account doc
func (dd *AccountDoc) GetID() DID {
	return dd.ID
}

// GetType gets type of account doc
func (dd *AccountDoc) GetType() int {
	return dd.BasicDoc.Type
}

// IsValidFormat checks whether account doc is valid format
func (dd *AccountDoc) IsValidFormat() bool {
	if dd.Created == 0 || dd.GetType() != int(AccountDIDType) {
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

// Marshal marshals account item
func (di *AccountItem) Marshal() ([]byte, error) {
	return Marshal(di)
}

// Unmarshal unmarshals account item
func (di *AccountItem) Unmarshal(docBytes []byte) error {
	return Unmarshal(docBytes, &di)
}

// GetID gets id of account item
func (di *AccountItem) GetID() DID {
	return di.ID
}

var _ AccountDIDManager = (*AccountDIDRegistry)(nil)

// AccountDIDRegistry for DID Identifier,
// Every appchain should use this DID Registry module.
type AccountDIDRegistry struct {
	Mode                     RegistryMode  `json:"mode"`
	SelfChainDID             DID           `json:"method"` // method of the registry
	Admins                   []DID         `json:"admins"` // admins of the registry
	Table                    RegistryTable `json:"table"`
	Docdb                    DocDB         `json:"docdb"`
	GenesisAccountDID        DID           `json:"genesis_account_did"`
	GenesisAccountDocInfo    DocInfo       `json:"genesis_account_doc_info"`
	GenesisAccountDocContent Doc           `json:"genesis_account_doc_content"`
	logger                   logrus.FieldLogger
	// config *DIDConfig
}

// NewAccountDIDRegistry news a AccountDIDRegistry
func NewAccountDIDRegistry(
	ts storage.Storage,
	l logrus.FieldLogger,
	options ...func(*AccountDIDRegistry)) (*AccountDIDRegistry, error) {
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
		// GenesisAccountDoc: DocInfo{
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

// WithAccountDocStorage used for InternalDocDB mode
func WithAccountDocStorage(ds storage.Storage) func(*AccountDIDRegistry) {
	return func(ar *AccountDIDRegistry) {
		db, _ := NewKVDocDB(ds)
		ar.Docdb = db
		ar.Mode = InternalDocDB
	}
}

// WithDIDAdmin used for admin setup
func WithDIDAdmin(a DID) func(*AccountDIDRegistry) {
	return func(ar *AccountDIDRegistry) {
		ar.Admins = []DID{a}
	}
}

// WithGenesisAccountDoc used for genesis account doc setup
func WithGenesisAccountDocInfo(DocInfo DocInfo) func(*AccountDIDRegistry) {
	return func(ar *AccountDIDRegistry) {
		ar.GenesisAccountDID = DocInfo.ID
		ar.GenesisAccountDocInfo = DocInfo
	}
}

// WithGenesisAccountDoc used for genesis account doc setup
func WithGenesisAccountDocContent(doc Doc) func(*AccountDIDRegistry) {
	return func(ar *AccountDIDRegistry) {
		ar.GenesisAccountDocContent = doc
		ar.GenesisAccountDID = doc.GetID()
	}
}

// SetupGenesis set up genesis to boot the whole did registry
func (r *AccountDIDRegistry) SetupGenesis() error {
	if r.GenesisAccountDID == "" {
		return fmt.Errorf("genesis AccountDID is null")
	}
	if len(r.Admins) == 0 {
		return fmt.Errorf("no admins")
	}
	// if r.GenesisAccountDID != r.GenesisAccountDoc.Content.GetID() {
	// 	return fmt.Errorf("genesis: admin DID not matched with doc")
	// }
	// register genesis did
	var err error
	if r.Mode == ExternalDocDB {
		_, _, err = r.Register(r.GenesisAccountDID, r.GenesisAccountDocInfo.Addr, r.GenesisAccountDocInfo.Hash)
	} else {
		_, _, err = r.RegisterWithDoc(r.GenesisAccountDocContent)
	}

	if err != nil {
		return fmt.Errorf("genesis: %w", err)
	}

	r.SelfChainDID = DID(r.GenesisAccountDID.GetChainDID())

	return nil
}

// GetSelfID gets genesis did of the registry
func (r *AccountDIDRegistry) GetSelfID() DID {
	return r.GenesisAccountDID
}

// GetAdmins gets admin list of the registry
func (r *AccountDIDRegistry) GetAdmins() []DID {
	return r.Admins
}

// AddAdmin adds an admin for the registry
func (r *AccountDIDRegistry) AddAdmin(caller DID) error {
	if r.HasAdmin(caller) {
		return fmt.Errorf("caller %s is already an admin", caller)
	}
	r.Admins = append(r.Admins, caller)
	return nil
}

// RemoveAdmin removes an admin for the registry
func (r *AccountDIDRegistry) RemoveAdmin(caller DID) error {
	for i, admin := range r.Admins {
		if admin == caller {
			r.Admins = append(r.Admins[:i], r.Admins[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("caller %s is not an admin", caller)
}

// HasAdmin checks whether caller is an admin of the registry
func (r *AccountDIDRegistry) HasAdmin(caller DID) bool {
	for _, v := range r.Admins {
		if v == caller {
			return true
		}
	}
	return false
}

// GetChainDID get chain did of the registry
func (r *AccountDIDRegistry) GetChainDID() DID {
	return r.SelfChainDID
}

// Register ties did name to a did doc
// ATN: only did who owns did-name should call this
func (r *AccountDIDRegistry) Register(accountDID DID, addr string, hash []byte) (string, []byte, error) {
	return r.updateByStatus(accountDID, addr, hash, nil, Initial)
}

// RegisterWithDoc registers with doc
func (r *AccountDIDRegistry) RegisterWithDoc(doc Doc) (string, []byte, error) {
	if !doc.IsValidFormat() {
		return "", nil, fmt.Errorf("invalid doc format")
	}
	return r.updateByStatus("", "", []byte{}, doc, Initial)
}

// Update updates data of an account did
// ATN: only caller who owns did should call this
func (r *AccountDIDRegistry) Update(accountDID DID, addr string, hash []byte) (string, []byte, error) {
	return r.updateByStatus(accountDID, addr, hash, nil, Normal)
}

// UpdateWithDoc updates with doc
func (r *AccountDIDRegistry) UpdateWithDoc(doc Doc) (string, []byte, error) {
	return r.updateByStatus("", "", []byte{}, doc, Normal)
}

func (r *AccountDIDRegistry) updateByStatus(did DID, docAddr string, docHash []byte, doc Doc, expectedStatus StatusType) (string, []byte, error) {
	docAddr, docHash, did, err := r.updateDocdbOrNot(did, docAddr, docHash, doc, expectedStatus)
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
		item, err := r.Table.GetItem(did, AccountDIDType)
		if err != nil {
			return docAddr, docHash, fmt.Errorf("DID table get: %w", err)
		}
		itemD := item.(*AccountItem)
		itemD.DocAddr = docAddr
		itemD.DocHash = docHash
		err = r.Table.UpdateItem(itemD)
		if err != nil {
			return docAddr, docHash, fmt.Errorf("update DID on table: %w", err)
		}
	}

	return docAddr, docHash, nil
}

func (r *AccountDIDRegistry) updateDocdbOrNot(
	did DID,
	docAddr string,
	docHash []byte,
	doc Doc,
	expectedStatus StatusType) (string, []byte, DID, error) {
	if r.Mode == InternalDocDB {
		if doc == nil {
			return "", nil, "", fmt.Errorf("doc content is nil")
		}
		doc := doc.(*AccountDoc)
		did = doc.GetID()

		// check exist
		exist := r.HasAccountDID(did)
		if expectedStatus == Initial && exist {
			return "", nil, "", fmt.Errorf("did %s already existed", did)
		} else if expectedStatus == Normal && !exist {
			return "", nil, "", fmt.Errorf("did %s not existed", did)
		}

		status := r.getDIDStatus(did)
		if status != expectedStatus {
			return "", nil, "", fmt.Errorf("did %s is under status: %s, expectd status: %s", did, status, expectedStatus)
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
		status := r.getDIDStatus(did)
		if status != expectedStatus {
			return "", nil, "",
				fmt.Errorf(
					"SelfChainDID %s is under status: %s, expectd status: %s",
					did,
					status,
					expectedStatus)
		}
	}
	return docAddr, docHash, did, nil
}

// Freeze freezes an account did
// ATN: only admin should call this.
func (r *AccountDIDRegistry) Freeze(did DID) error {
	exist := r.HasAccountDID(did)
	if !exist {
		return fmt.Errorf("did %s not existed", did)
	}
	return r.auditStatus(did, Frozen)
}

// UnFreeze unfreezes an account did
// ATN: only admin should call this.
func (r *AccountDIDRegistry) UnFreeze(did DID) error {
	exist := r.HasAccountDID(did)
	if !exist {
		return fmt.Errorf("did %s not existed", did)
	}
	return r.auditStatus(did, Normal)
}

// Resolve looks up local-chain to resolve did.
// @*AccountDoc returns nil if mode is ExternalDocDB
func (r *AccountDIDRegistry) Resolve(did DID) (*AccountItem, *AccountDoc, bool, error) {
	exist := r.HasAccountDID(did)
	if !exist {
		return nil, nil, false, fmt.Errorf("did %s not existed", did)
	}

	item, err := r.Table.GetItem(did, AccountDIDType)
	if err != nil {
		return nil, nil, false, fmt.Errorf("resolve DID table get: %w", err)
	}
	itemD := item.(*AccountItem)

	if r.Mode == InternalDocDB {
		doc, err := r.Docdb.Get(did, AccountDIDType)
		if err != nil {
			return itemD, nil, true, fmt.Errorf("resolve DID docdb get: %w", err)
		}
		docD := doc.(*AccountDoc)
		return itemD, docD, true, nil
	}

	return itemD, nil, true, nil
}

// Delete deletes data of an account did
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

// HasAccountDID checks whether an account did exists
func (r *AccountDIDRegistry) HasAccountDID(did DID) bool {
	exist := r.Table.HasItem(did)
	return exist
}

func (r *AccountDIDRegistry) getDIDStatus(did DID) StatusType {
	if !r.Table.HasItem(did) {
		return Initial
	}
	item, err := r.Table.GetItem(did, AccountDIDType)
	if err != nil {
		r.logger.Error("did status get item:", err)
		return BadStatus
	}
	itemD := item.(*AccountItem)
	return itemD.Status
}

//  caller naturally owns the did ended with his address.
func (r *AccountDIDRegistry) owns(caller string, did DID) bool {
	s := strings.Split(string(did), ":")
	return s[len(s)-1] == caller
}

func (r *AccountDIDRegistry) auditStatus(did DID, status StatusType) error {
	item, err := r.Table.GetItem(did, AccountDIDType)
	if err != nil {
		return fmt.Errorf("did status get: %w", err)
	}
	itemD := item.(*AccountItem)
	itemD.Status = status
	err = r.Table.UpdateItem(item)
	if err != nil {
		return fmt.Errorf("did status update: %w", err)
	}
	return nil
}
