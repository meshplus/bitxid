package bitxid

import (
	"crypto/sha256"
	"fmt"

	"github.com/meshplus/bitxhub-kit/storage"

	"github.com/sirupsen/logrus"
)

var _ Doc = (*ChainDoc)(nil)

// ChainDoc .
type ChainDoc struct {
	BasicDoc
	Extra []byte `json:"extra"` // for further usage
}

// Marshal .
func (cd *ChainDoc) Marshal() ([]byte, error) {
	return Struct2Bytes(cd)
}

// Unmarshal .
func (cd *ChainDoc) Unmarshal(docBytes []byte) error {
	return Bytes2Struct(docBytes, &cd)
}

// GetID .
func (cd *ChainDoc) GetID() DID {
	return cd.ID
}

func (md *ChainDoc) IsValidFormat() bool {
	if md.Created == 0 || !md.ID.IsChainDIDFormat() {
		return false
	}
	return true
}

var _ TableItem = (*ChainItem)(nil)

// ChainItem reperesents a method item, element of registry table,
// it stores all data about a did.
// Registry table is used together with docdb.
type ChainItem struct {
	BasicItem
	Owner DID // owner of the method, is a did, TODO: owner ==> owners
}

// Marshal .
func (mi *ChainItem) Marshal() ([]byte, error) {
	return Struct2Bytes(mi)
}

// Unmarshal .
func (mi *ChainItem) Unmarshal(docBytes []byte) error {
	return Bytes2Struct(docBytes, &mi)
}

// GetID .
func (mi *ChainItem) GetID() DID {
	return mi.ID
}

var _ ChainDIDManager = (*ChainDIDRegistry)(nil)

// ChainDIDRegistry .
type ChainDIDRegistry struct {
	Mode            RegistryMode  `json:"mode"`
	IsRoot          bool          `json:"is_root"`
	Admins          []DID         `json:"admins"`
	Table           RegistryTable `json:"table"`
	Docdb           DocDB         `json:"docdb"`
	GenesisChainDID DID           `json:"genesis_chain_did"`
	GenesisChainDoc DocOption     `json:"genesis_chain_doc"`
	logger          logrus.FieldLogger
}

// NewChainDIDRegistry news a ChainDIDRegistry
func NewChainDIDRegistry(ts storage.Storage, l logrus.FieldLogger, options ...func(*ChainDIDRegistry)) (*ChainDIDRegistry, error) {
	rt, _ := NewKVTable(ts)
	db, _ := NewKVDocDB(nil)
	// doc := GenesisChainDoc()
	cr := &ChainDIDRegistry{ // default config
		Mode:   ExternalDocDB,
		Table:  rt,
		Docdb:  db,
		logger: l,
		// Admins: []DID{genesisDIDDoc().GetID()},
		// IsRoot: true,
		// GenesisChainDID: doc.GetID(),
		// GenesisChainDoc: DocOption{
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

// WithChainDocStorage .
func WithChainDocStorage(ds storage.Storage) func(*ChainDIDRegistry) {
	return func(cr *ChainDIDRegistry) {
		db, _ := NewKVDocDB(ds)
		cr.Docdb = db
		cr.Mode = InternalDocDB
	}
}

// WithMethodAdmin .
func WithMethodAdmin(a DID) func(*ChainDIDRegistry) {
	return func(cr *ChainDIDRegistry) {
		cr.Admins = []DID{a}
	}
}

// WithGenesisChainDoc .
func WithGenesisChainDoc(docOption DocOption) func(*ChainDIDRegistry) {
	return func(cr *ChainDIDRegistry) {
		cr.GenesisChainDoc = docOption
		cr.GenesisChainDID = DID(docOption.ID.GetChainDID())
	}
}

// SetupGenesis set up genesis to boot the whole methed system
func (r *ChainDIDRegistry) SetupGenesis() error {
	if r.GenesisChainDID == "" {
		return fmt.Errorf("genesis ChainDID is null")
	}
	if len(r.Admins) == 0 {
		return fmt.Errorf("No admins")
	}
	if r.GenesisChainDID != r.GenesisChainDoc.Content.(*ChainDoc).ID {
		return fmt.Errorf("genesis ChainDID not matched with ChainDoc")
	}

	// register method:
	err := r.Apply(r.Admins[0], r.GenesisChainDID)
	if err != nil {
		return fmt.Errorf("genesis apply err: %w", err)
	}
	err = r.AuditApply(r.GenesisChainDID, true)
	if err != nil {
		return fmt.Errorf("genesis audit err: %w", err)
	}
	_, _, err = r.Register(r.GenesisChainDoc)
	if err != nil {
		return fmt.Errorf("genesis register err: %w", err)
	}

	return nil
}

// GetSelfID .
func (r *ChainDIDRegistry) GetSelfID() DID {
	return r.GenesisChainDID
}

// GetAdmins .
func (r *ChainDIDRegistry) GetAdmins() []DID {
	return r.Admins
}

// AddAdmin .
func (r *ChainDIDRegistry) AddAdmin(caller DID) error {
	if r.HasAdmin(caller) {
		return fmt.Errorf("caller %s is already an admin", caller)
	}
	r.Admins = append(r.Admins, caller)
	return nil
}

// RemoveAdmin .
func (r *ChainDIDRegistry) RemoveAdmin(caller DID) error {
	for i, admin := range r.Admins {
		if admin == caller {
			r.Admins = append(r.Admins[:i], r.Admins[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("caller %s is not an admin", caller)
}

// HasAdmin .
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
	if !DID(chainDID).IsValidFormat() {
		return fmt.Errorf("ChainDID is not standard")
	}

	status := r.GetChainDIDStatus(chainDID)
	if status != Initial {
		return fmt.Errorf("can not apply %s under status: %s", chainDID, status)
	}
	// creates item in table
	err := r.Table.CreateItem(
		&ChainItem{
			BasicItem{
				ID:     DID(chainDID),
				Status: ApplyAudit},
			caller})
	if err != nil {
		return fmt.Errorf("apply %s on table: %w", chainDID, err)
	}
	return nil
}

// AuditApply .
// ATNS: only admin should call this.
func (r *ChainDIDRegistry) AuditApply(chainDID DID, result bool) error {
	exist := r.HasChainDID(chainDID)
	if exist == false {
		return fmt.Errorf("auditapply %s not existed", chainDID)
	}
	status := r.GetChainDIDStatus(chainDID)
	if !(status == ApplyAudit || status == ApplyFailed) {
		return fmt.Errorf("can not auditapply %s under status: %s", chainDID, status)
	}
	var err error = nil
	if result {
		err = r.auditStatus(chainDID, ApplySuccess)
	} else {
		err = r.auditStatus(chainDID, ApplyFailed)
	}
	return err
}

// Synchronize synchronizes table data between different registrys
func (r *ChainDIDRegistry) Synchronize(item *ChainItem) error {
	return r.Table.CreateItem(item)
}

// Register ties method name to a method doc
// ATN: only did who owns method-name should call this
func (r *ChainDIDRegistry) Register(docOption DocOption) (string, []byte, error) { // doc *ChainDoc
	return r.updateByStatus(docOption, ApplySuccess, Normal)
}

// Update .
// ATN: only did who owns method-name should call this.
func (r *ChainDIDRegistry) Update(docOption DocOption) (string, []byte, error) {
	return r.updateByStatus(docOption, Normal, Normal)
}

// update with expected status
func (r *ChainDIDRegistry) updateByStatus(docOption DocOption, expectedStatus StatusType, status StatusType) (string, []byte, error) {

	docAddr, docHash, method, err := r.updateDocdbOrNot(docOption, expectedStatus, status)
	if err != nil {
		return "", nil, err
	}

	item, err := r.Table.GetItem(method, MethodTableType)
	if err != nil {
		return docAddr, docHash, fmt.Errorf("table get item: %w ", err)
	}
	itemM := item.(*ChainItem)
	itemM.DocAddr = docAddr
	itemM.DocHash = docHash
	itemM.Status = status
	err = r.Table.UpdateItem(itemM)
	if err != nil {
		return docAddr, docHash, fmt.Errorf("table update item: %w ", err)
	}

	return docAddr, docHash, nil
}

func (r *ChainDIDRegistry) updateDocdbOrNot(docOption DocOption, expectedStatus StatusType, status StatusType) (string, []byte, DID, error) {
	var docAddr string
	var docHash []byte
	var chainDID DID
	if r.Mode == InternalDocDB {
		// check exist
		doc := docOption.Content.(*ChainDoc)
		chainDID = doc.GetID()
		status := r.GetChainDIDStatus(chainDID)
		if status != expectedStatus {
			return "", nil, "", fmt.Errorf("Method %s is under status: %s, expected status: %s", chainDID, status, expectedStatus)
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
		chainDID = docOption.ID
		docAddr = docOption.Addr
		docHash = docOption.Hash
		status := r.GetChainDIDStatus(chainDID)
		if status != expectedStatus {
			return "", nil, "", fmt.Errorf("Method %s is under status: %s, expected status: %s", chainDID, status, expectedStatus)
		}
	}
	return docAddr, docHash, chainDID, nil
}

// Audit .
// ATN: only admin should call this.
func (r *ChainDIDRegistry) Audit(chainDID DID, status StatusType) error {
	exist := r.HasChainDID(chainDID)
	if exist == false {
		return fmt.Errorf("audit %s not existed", chainDID)
	}
	return r.auditStatus(chainDID, status)
}

// Freeze .
// ATN: only admdin should call this.
func (r *ChainDIDRegistry) Freeze(chainDID DID) error {
	exist := r.HasChainDID(chainDID)
	if exist == false {
		return fmt.Errorf("freeze %s not existed", chainDID)
	}
	return r.auditStatus(chainDID, Frozen)
}

// UnFreeze .
// ATN: only admdin should call this.
func (r *ChainDIDRegistry) UnFreeze(chainDID DID) error {
	exist := r.HasChainDID(chainDID)
	if exist == false {
		return fmt.Errorf("unfreeze %s not existed", chainDID)
	}

	return r.auditStatus(chainDID, Normal)
}

// Delete .
func (r *ChainDIDRegistry) Delete(chainDID DID) error {
	err := r.auditStatus(chainDID, Initial)
	if err != nil {
		return fmt.Errorf("Method delete: %w", err)
	}

	r.Table.DeleteItem(chainDID)

	if r.Mode == InternalDocDB {
		r.Docdb.Delete(chainDID)
	}

	return nil
}

// Resolve looks up local-chain to resolve method.
// @*DIDDoc returns nil if mode is ExternalDocDB
func (r *ChainDIDRegistry) Resolve(chainDID DID) (*ChainItem, *ChainDoc, bool, error) {
	exist := r.HasChainDID(chainDID)
	if exist == false {
		return nil, nil, false, nil
	}
	item, err := r.Table.GetItem(chainDID, MethodTableType)
	if err != nil {
		return nil, nil, false, fmt.Errorf("Method resolve table get: %w", err)
	}
	itemM := item.(*ChainItem)

	if r.Mode == InternalDocDB {
		doc, err := r.Docdb.Get(chainDID, ChainDocType)
		if err != nil {
			return itemM, nil, true, fmt.Errorf("Method resolve docdb get: %w", err)
		}
		docM := doc.(*ChainDoc)
		return itemM, docM, true, nil
	}
	return itemM, nil, true, nil
}

// HasChainDID .
func (r *ChainDIDRegistry) HasChainDID(chainDID DID) bool {
	exist := r.Table.HasItem(chainDID)
	return exist
}

func (r *ChainDIDRegistry) GetChainDIDStatus(chainDID DID) StatusType {
	if !r.Table.HasItem(chainDID) {
		return Initial
	}
	item, err := r.Table.GetItem(chainDID, MethodTableType)
	if err != nil {
		r.logger.Error("chainDID status get item:", err)
		return BadStatus
	}
	itemM := item.(*ChainItem)
	return itemM.Status
}

// auditStatus .
func (r *ChainDIDRegistry) auditStatus(chainDID DID, status StatusType) error {
	item, err := r.Table.GetItem(chainDID, MethodTableType)
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
