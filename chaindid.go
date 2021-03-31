package bitxid

import (
	"crypto/sha256"
	"fmt"

	"github.com/meshplus/bitxhub-kit/storage"

	"github.com/sirupsen/logrus"
)

var _ Doc = (*ChainDoc)(nil)

// ChainDoc represents chain identity information
type ChainDoc struct {
	BasicDoc
	Extra []byte `json:"extra"` // for further usage
}

// Marshal marshals chain doc
func (cd *ChainDoc) Marshal() ([]byte, error) {
	return Marshal(cd)
}

// Unmarshal unmarshals chain doc
func (cd *ChainDoc) Unmarshal(docBytes []byte) error {
	return Unmarshal(docBytes, &cd)
}

// GetID gets id of chain doc
func (cd *ChainDoc) GetID() DID {
	return cd.ID
}

// GetType gets type of chain doc
func (cd *ChainDoc) GetType() int {
	return cd.BasicDoc.Type
}

// IsValidFormat checks whether chain doc is valid format
func (cd *ChainDoc) IsValidFormat() bool {
	if cd.Created == 0 || cd.GetType() != int(ChainDIDType) {
		return false
	}
	return true
}

var _ TableItem = (*ChainItem)(nil)

// ChainItem reperesents a chain did item, element of registry table,
// it stores all data about a did.
// Registry table is used together with docdb.
type ChainItem struct {
	BasicItem
	Owner DID // owner of the chain did, is a did, TODO: owner ==> owners
}

// Marshal marshals chain item
func (mi *ChainItem) Marshal() ([]byte, error) {
	return Marshal(mi)
}

// Unmarshal unmarshals chain item
func (mi *ChainItem) Unmarshal(docBytes []byte) error {
	return Unmarshal(docBytes, &mi)
}

// GetID gets id of chain item
func (mi *ChainItem) GetID() DID {
	return mi.ID
}

var _ ChainDIDManager = (*ChainDIDRegistry)(nil)

// ChainDIDRegistry .
type ChainDIDRegistry struct {
	Mode                   RegistryMode  `json:"mode"`
	IsRoot                 bool          `json:"is_root"`
	Admins                 []DID         `json:"admins"`
	Table                  RegistryTable `json:"table"`
	Docdb                  DocDB         `json:"docdb"`
	GenesisChainDID        DID           `json:"genesis_chain_did"`
	GenesisChainDocInfo    DocInfo       `json:"genesis_chain_doc_info"`
	GenesisChainDocContent Doc           `json:"genesis_chain_doc_content"`
	logger                 logrus.FieldLogger
}

// NewChainDIDRegistry news a ChainDIDRegistry
func NewChainDIDRegistry(
	ts storage.Storage,
	l logrus.FieldLogger,
	options ...func(*ChainDIDRegistry)) (*ChainDIDRegistry, error) {
	rt, _ := NewKVTable(ts)
	db, _ := NewKVDocDB(nil)
	// doc := GenesisChainDoc()
	cr := &ChainDIDRegistry{ // default config
		Mode:   ExternalDocDB,
		Table:  rt,
		Docdb:  db,
		logger: l,
		// Admins: []DID{genesisAccountDoc().GetID()},
		// IsRoot: true,
		// GenesisChainDID: doc.GetID(),
		// GenesisChainDoc: DocInfo{
		// 	ID:      doc.GetID(),
		// 	Addr:    ".",
		// 	Hash:    []byte{0},
		// 	Content: doc,
		// },
	}
	// TODO: Check field
	for _, option := range options {
		option(cr)
	}

	return cr, nil
}

// WithChainDocStorage used for InternalDocDB mode
func WithChainDocStorage(ds storage.Storage) func(*ChainDIDRegistry) {
	return func(cr *ChainDIDRegistry) {
		db, _ := NewKVDocDB(ds)
		cr.Docdb = db
		cr.Mode = InternalDocDB
	}
}

// WithAdmin used for admin setup
func WithAdmin(a DID) func(*ChainDIDRegistry) {
	return func(cr *ChainDIDRegistry) {
		cr.Admins = []DID{a}
	}
}

// WithGenesisChainDoc used for genesis chain doc setup
func WithGenesisChainDocInfo(docInfo DocInfo) func(*ChainDIDRegistry) {
	return func(cr *ChainDIDRegistry) {
		cr.GenesisChainDID = docInfo.ID
		cr.GenesisChainDocInfo = docInfo
	}
}

// WithGenesisChainDoc used for genesis chain doc setup
func WithGenesisChainDocContent(doc Doc) func(*ChainDIDRegistry) {
	return func(cr *ChainDIDRegistry) {
		cr.GenesisChainDID = doc.GetID()
		cr.GenesisChainDocContent = doc
	}
}

// SetupGenesis set up genesis to boot the whole methed system
func (r *ChainDIDRegistry) SetupGenesis() error {
	if r.GenesisChainDID == "" {
		return fmt.Errorf("genesis ChainDID is null")
	}
	if len(r.Admins) == 0 {
		return fmt.Errorf("no admins")
	}
	// if r.GenesisChainDID != r.GenesisChainDoc.Content.(*ChainDoc).ID {
	// 	return fmt.Errorf("genesis ChainDID not matched with ChainDoc")
	// }

	// register chain did:
	err := r.Apply(r.Admins[0], r.GenesisChainDID)
	if err != nil {
		return fmt.Errorf("genesis apply err: %w", err)
	}
	err = r.AuditApply(r.GenesisChainDID, true)
	if err != nil {
		return fmt.Errorf("genesis audit err: %w", err)
	}
	if r.Mode == ExternalDocDB {
		_, _, err = r.Register(r.GenesisChainDocInfo.ID, r.GenesisChainDocInfo.Addr, r.GenesisChainDocInfo.Hash)
	} else {
		_, _, err = r.RegisterWithDoc(r.GenesisChainDocContent)
	}
	if err != nil {
		return fmt.Errorf("genesis register err: %w", err)
	}

	return nil
}

// GetSelfID gets genesis did of the registry
func (r *ChainDIDRegistry) GetSelfID() DID {
	return r.GenesisChainDID
}

// GetAdmins gets admin list of the registry
func (r *ChainDIDRegistry) GetAdmins() []DID {
	return r.Admins
}

// AddAdmin adds an admin for the registry
func (r *ChainDIDRegistry) AddAdmin(caller DID) error {
	if r.HasAdmin(caller) {
		return fmt.Errorf("caller %s is already an admin", caller)
	}
	r.Admins = append(r.Admins, caller)
	return nil
}

// RemoveAdmin removes an admin for the registry
func (r *ChainDIDRegistry) RemoveAdmin(caller DID) error {
	for i, admin := range r.Admins {
		if admin == caller {
			r.Admins = append(r.Admins[:i], r.Admins[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("caller %s is not an admin", caller)
}

// HasAdmin checks whether caller is an admin of the registry
func (r *ChainDIDRegistry) HasAdmin(caller DID) bool {
	for _, v := range r.Admins {
		if v == caller {
			return true
		}
	}
	return false
}

// Apply apply for rights of a new methd-name
func (r *ChainDIDRegistry) Apply(caller DID, chainDID DID) error {
	// check if ChainDID Name meets standard
	if !chainDID.IsValidFormat() {
		return fmt.Errorf("chain did is not standard")
	}

	status := r.getChainDIDStatus(chainDID)
	if status != Initial {
		return fmt.Errorf("can not apply %s under status: %s", chainDID, status)
	}
	// creates item in table
	err := r.Table.CreateItem(
		&ChainItem{
			BasicItem{
				ID:     chainDID,
				Status: ApplyAudit},
			caller})
	if err != nil {
		return fmt.Errorf("apply %s on table: %w", chainDID, err)
	}
	return nil
}

// AuditApply audits status of a chain did application
// ATNS: only admin should call this.
func (r *ChainDIDRegistry) AuditApply(chainDID DID, result bool) error {
	exist := r.HasChainDID(chainDID)
	if !exist {
		return fmt.Errorf("auditapply %s not existed", chainDID)
	}
	status := r.getChainDIDStatus(chainDID)
	if !(status == ApplyAudit || status == ApplyFailed) {
		return fmt.Errorf("can not auditapply %s under status: %s", chainDID, status)
	}
	var err error
	if result {
		err = r.auditStatus(chainDID, ApplySuccess)
	} else {
		err = r.auditStatus(chainDID, ApplyFailed)
	}
	return err
}

// Synchronize synchronizes table data between different registrys
func (r *ChainDIDRegistry) Synchronize(item TableItem) error {
	return r.Table.CreateItem(item)
}

// Register ties chain did to a chain doc
// ATN: only did who owns method-name should call this
func (r *ChainDIDRegistry) Register(chainDID DID, addr string, hash []byte) (string, []byte, error) {
	return r.updateByStatus(chainDID, addr, hash, nil, ApplySuccess)
}

// RegisterWithDoc registers with doc
func (r *ChainDIDRegistry) RegisterWithDoc(doc Doc) (string, []byte, error) {
	return r.updateByStatus("", "", []byte{}, doc, ApplySuccess)
}

// Update updates data about a chain did
// ATN: only did who owns method-name should call this.
func (r *ChainDIDRegistry) Update(chainDID DID, addr string, hash []byte) (string, []byte, error) {
	return r.updateByStatus(chainDID, addr, hash, nil, Normal)
}

// UpdateWithDoc updates with doc
func (r *ChainDIDRegistry) UpdateWithDoc(doc Doc) (string, []byte, error) {
	return r.updateByStatus("", "", []byte{}, doc, Normal)
}

func (r *ChainDIDRegistry) updateByStatus(chainDID DID, docAddr string, docHash []byte, doc Doc, expectedStatus StatusType) (string, []byte, error) {
	// update doc concerned data
	docAddr, docHash, chainDID, err := r.updateDocdbOrNot(chainDID, docAddr, docHash, doc, expectedStatus)
	if err != nil {
		return "", nil, err
	}

	// update table concerned data
	item, err := r.Table.GetItem(chainDID, ChainDIDType)
	if err != nil {
		return docAddr, docHash, fmt.Errorf("table get item: %w ", err)
	}
	itemM := item.(*ChainItem)
	itemM.DocAddr = docAddr
	itemM.DocHash = docHash
	itemM.Status = Normal
	err = r.Table.UpdateItem(itemM)
	if err != nil {
		return docAddr, docHash, fmt.Errorf("table update item: %w ", err)
	}

	return docAddr, docHash, nil
}

// updateDocdbOrNot will updata DocDB(when under InternalDocDB mode) or not(when under ExternalDocDB mode)
func (r *ChainDIDRegistry) updateDocdbOrNot(
	chainDID DID,
	docAddr string,
	docHash []byte,
	doc Doc,
	expectedStatus StatusType) (string, []byte, DID, error) {
	if r.Mode == InternalDocDB {
		// check exist
		if doc == nil {
			return "", nil, "", fmt.Errorf("doc content is nil")
		}
		doc := doc.(*ChainDoc)
		chainDID = doc.GetID()
		status := r.getChainDIDStatus(chainDID)
		if status != expectedStatus {
			return "", nil, "",
				fmt.Errorf(
					"chain did %s is under status: %s, expected status: %s",
					chainDID,
					status,
					expectedStatus)
		}

		docBytes, err := doc.Marshal()
		if err != nil {
			return "", nil, "", fmt.Errorf("doc marshal: %w ", err)
		}

		if expectedStatus == ApplySuccess { // register
			docAddr, err = r.Docdb.Create(doc)
		} else { // update
			docAddr, err = r.Docdb.Update(doc)
		}
		if err != nil {
			return "", nil, "", fmt.Errorf("update docdb: %w ", err)
		}
		docHash32 := sha256.Sum256(docBytes)
		docHash = docHash32[:]
	} else {
		status := r.getChainDIDStatus(chainDID)
		if status != expectedStatus {
			return "", nil, "",
				fmt.Errorf(
					"chain did %s is under status: %s, expected status: %s",
					chainDID,
					status,
					expectedStatus)
		}
	}
	return docAddr, docHash, chainDID, nil
}

// Audit audits status of a chain did
// ATN: only admin should call this.
func (r *ChainDIDRegistry) Audit(chainDID DID, status StatusType) error {
	exist := r.HasChainDID(chainDID)
	if !exist {
		return fmt.Errorf("audit %s not existed", chainDID)
	}
	return r.auditStatus(chainDID, status)
}

// Freeze freezes a chain did
// ATN: only admdin should call this.
func (r *ChainDIDRegistry) Freeze(chainDID DID) error {
	exist := r.HasChainDID(chainDID)
	if !exist {
		return fmt.Errorf("freeze %s not existed", chainDID)
	}
	return r.auditStatus(chainDID, Frozen)
}

// UnFreeze unfreezes a chain did
// ATN: only admdin should call this.
func (r *ChainDIDRegistry) UnFreeze(chainDID DID) error {
	exist := r.HasChainDID(chainDID)
	if !exist {
		return fmt.Errorf("unfreeze %s not existed", chainDID)
	}

	return r.auditStatus(chainDID, Normal)
}

// Delete deletes data of a chain did
func (r *ChainDIDRegistry) Delete(chainDID DID) error {
	err := r.auditStatus(chainDID, Initial)
	if err != nil {
		return fmt.Errorf("chain did delete: %w", err)
	}

	r.Table.DeleteItem(chainDID)

	if r.Mode == InternalDocDB {
		r.Docdb.Delete(chainDID)
	}

	return nil
}

// Resolve looks up local-chain to resolve chain did.
// @*ChainDoc returns nil if mode is ExternalDocDB
func (r *ChainDIDRegistry) Resolve(chainDID DID) (*ChainItem, *ChainDoc, bool, error) {
	exist := r.HasChainDID(chainDID)
	if !exist {
		return nil, nil, false, nil
	}
	item, err := r.Table.GetItem(chainDID, ChainDIDType)
	if err != nil {
		return nil, nil, false, fmt.Errorf("chain did resolve table get: %w", err)
	}
	itemM := item.(*ChainItem)

	if r.Mode == InternalDocDB {
		doc, err := r.Docdb.Get(chainDID, ChainDIDType)
		if err != nil {
			return itemM, nil, true, fmt.Errorf("chain did resolve docdb get: %w", err)
		}
		docM := doc.(*ChainDoc)
		return itemM, docM, true, nil
	}
	return itemM, nil, true, nil
}

// HasChainDID checks whether a chain did exists
func (r *ChainDIDRegistry) HasChainDID(chainDID DID) bool {
	exist := r.Table.HasItem(chainDID)
	return exist
}

func (r *ChainDIDRegistry) getChainDIDStatus(chainDID DID) StatusType {
	if !r.Table.HasItem(chainDID) {
		return Initial
	}
	item, err := r.Table.GetItem(chainDID, ChainDIDType)
	if err != nil {
		r.logger.Error("chainDID status get item:", err)
		return BadStatus
	}
	itemM := item.(*ChainItem)
	return itemM.Status
}

func (r *ChainDIDRegistry) auditStatus(chainDID DID, status StatusType) error {
	item, err := r.Table.GetItem(chainDID, ChainDIDType)
	if err != nil {
		return fmt.Errorf("aduitstatus table get: %w", err)
	}
	itemM := item.(*ChainItem)
	itemM.Status = status
	err = r.Table.UpdateItem(itemM)
	if err != nil {
		return fmt.Errorf("aduitstatus table update: %w", err)
	}
	return nil
}
