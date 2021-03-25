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
func (md *ChainDoc) Marshal() ([]byte, error) {
	return Struct2Bytes(md)
}

// Unmarshal .
func (md *ChainDoc) Unmarshal(docBytes []byte) error {
	return Bytes2Struct(docBytes, &md)
}

// GetID .
func (md *ChainDoc) GetID() DID {
	return md.ID
}

func (md *ChainDoc) IsValidFormat() bool {
	if md.Created == 0 || !md.ID.IsChainDIDFormat() {
		return false
	}
	return true
}

var _ TableItem = (*MethodItem)(nil)

// MethodItem reperesents a method item, element of registry table,
// it stores all data about a did.
// Registry table is used together with docdb.
type MethodItem struct {
	BasicItem
	Owner DID // owner of the method, is a did, TODO: owner ==> owners
}

// Marshal .
func (mi *MethodItem) Marshal() ([]byte, error) {
	return Struct2Bytes(mi)
}

// Unmarshal .
func (mi *MethodItem) Unmarshal(docBytes []byte) error {
	return Bytes2Struct(docBytes, &mi)
}

// GetID .
func (mi *MethodItem) GetID() DID {
	return mi.ID
}

var _ MethodManager = (*MethodRegistry)(nil)

// MethodRegistry .
type MethodRegistry struct {
	Mode          RegistryMode  `json:"mode"`
	IsRoot        bool          `json:"is_root"`
	Admins        []DID         `json:"admins"`
	Table         RegistryTable `json:"table"`
	Docdb         DocDB         `json:"docdb"`
	GenesisMethod DID           `json:"genesis_method"`
	GenesisDoc    DocOption     `json:"genesis_doc"`
	logger        logrus.FieldLogger
}

// NewMethodRegistry news a MethodRegistry
func NewMethodRegistry(ts storage.Storage, l logrus.FieldLogger, options ...func(*MethodRegistry)) (*MethodRegistry, error) {
	rt, _ := NewKVTable(ts)
	db, _ := NewKVDocDB(nil)
	// doc := GenesisChainDoc()
	mr := &MethodRegistry{ // default config
		Mode:   ExternalDocDB,
		Table:  rt,
		Docdb:  db,
		logger: l,
		Admins: []DID{genesisDIDDoc().GetID()},
		// IsRoot: true,
		// GenesisMethod: doc.GetID(),
		// GenesisDoc: DocOption{
		// 	ID:      doc.GetID(),
		// 	Addr:    ".",
		// 	Hash:    []byte{0},
		// 	Content: doc,
		// },
	}
	// TODO: Check field
	for _, option := range options {
		option(mr)
	}

	return mr, nil
}

// WithChainDocStorage .
func WithChainDocStorage(ds storage.Storage) func(*MethodRegistry) {
	return func(mr *MethodRegistry) {
		db, _ := NewKVDocDB(ds)
		mr.Docdb = db
		mr.Mode = InternalDocDB
	}
}

// WithMethodAdmin .
func WithMethodAdmin(a DID) func(*MethodRegistry) {
	return func(mr *MethodRegistry) {
		mr.Admins = []DID{a}
	}
}

// // WithGenesisMethod .
// func WithGenesisMethod(m DID) func(*MethodRegistry) {
// 	return func(mr *MethodRegistry) {
// 		mr.GenesisMethod = m
// 	}
// }

// WithGenesisChainDoc .
func WithGenesisChainDoc(docOption DocOption) func(*MethodRegistry) {
	return func(mr *MethodRegistry) {
		mr.GenesisDoc = docOption
		mr.GenesisMethod = DID(docOption.ID.GetMethod())
	}
}

// SetupGenesis set up genesis to boot the whole methed system
func (r *MethodRegistry) SetupGenesis() error {
	if r.GenesisMethod == "" {
		return fmt.Errorf("genesis Method is null")
	}

	if r.GenesisMethod != r.GenesisDoc.Content.(*ChainDoc).ID {
		return fmt.Errorf("genesis Method not matched with doc")
	}

	// register method:
	err := r.Apply(r.Admins[0], r.GenesisMethod)
	if err != nil {
		return fmt.Errorf("genesis apply err: %w", err)
	}
	err = r.AuditApply(r.GenesisMethod, true)
	if err != nil {
		return fmt.Errorf("genesis audit err: %w", err)
	}
	_, _, err = r.Register(r.GenesisDoc)
	if err != nil {
		return fmt.Errorf("genesis register err: %w", err)
	}

	return nil
}

// GetSelfID .
func (r *MethodRegistry) GetSelfID() DID {
	return r.GenesisMethod
}

// GetAdmins .
func (r *MethodRegistry) GetAdmins() []DID {
	return r.Admins
}

// AddAdmin .
func (r *MethodRegistry) AddAdmin(caller DID) error {
	if r.HasAdmin(caller) {
		return fmt.Errorf("caller %s is already an admin", caller)
	}
	r.Admins = append(r.Admins, caller)
	return nil
}

// RemoveAdmin .
func (r *MethodRegistry) RemoveAdmin(caller DID) error {
	for i, admin := range r.Admins {
		if admin == caller {
			r.Admins = append(r.Admins[:i], r.Admins[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("caller %s is not an admin", caller)
}

// HasAdmin .
func (r *MethodRegistry) HasAdmin(caller DID) bool {
	for _, v := range r.Admins {
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
		return fmt.Errorf("Method name is not standard")
	}

	status := r.getMethodStatus(method)
	if status != Initial {
		return fmt.Errorf("can not apply %s under status: %s", method, status)
	}
	// creates item in table
	err := r.Table.CreateItem(
		&MethodItem{
			BasicItem{
				ID:     DID(method),
				Status: ApplyAudit},
			caller})
	if err != nil {
		return fmt.Errorf("apply %s on table: %w", method, err)
	}
	return nil
}

// AuditApply .
// ATNS: only admin should call this.
func (r *MethodRegistry) AuditApply(method DID, result bool) error {
	exist := r.HasMethod(method)
	if exist == false {
		return fmt.Errorf("auditapply %s not existed", method)
	}
	status := r.getMethodStatus(method)
	if !(status == ApplyAudit || status == ApplyFailed) {
		return fmt.Errorf("can not auditapply %s under status: %s", method, status)
	}
	var err error = nil
	if result {
		err = r.auditStatus(method, ApplySuccess)
	} else {
		err = r.auditStatus(method, ApplyFailed)
	}
	return err
}

// Synchronize synchronizes table data between different registrys
func (r *MethodRegistry) Synchronize(item *MethodItem) error {
	return r.Table.CreateItem(item)
}

// Register ties method name to a method doc
// ATN: only did who owns method-name should call this
func (r *MethodRegistry) Register(docOption DocOption) (string, []byte, error) { // doc *ChainDoc
	return r.updateByStatus(docOption, ApplySuccess, Normal)
}

// Update .
// ATN: only did who owns method-name should call this.
func (r *MethodRegistry) Update(docOption DocOption) (string, []byte, error) {
	return r.updateByStatus(docOption, Normal, Normal)
}

// update with expected status
func (r *MethodRegistry) updateByStatus(docOption DocOption, expectedStatus StatusType, status StatusType) (string, []byte, error) {

	docAddr, docHash, method, err := r.updateDocdbOrNot(docOption, expectedStatus, status)
	if err != nil {
		return "", nil, err
	}

	item, err := r.Table.GetItem(method, MethodTableType)
	if err != nil {
		return docAddr, docHash, fmt.Errorf("table get item: %w ", err)
	}
	itemM := item.(*MethodItem)
	itemM.DocAddr = docAddr
	itemM.DocHash = docHash
	itemM.Status = status
	err = r.Table.UpdateItem(itemM)
	if err != nil {
		return docAddr, docHash, fmt.Errorf("table update item: %w ", err)
	}

	return docAddr, docHash, nil
}

func (r *MethodRegistry) updateDocdbOrNot(docOption DocOption, expectedStatus StatusType, status StatusType) (string, []byte, DID, error) {
	var docAddr string
	var docHash []byte
	var method DID
	if r.Mode == InternalDocDB {
		// check exist
		doc := docOption.Content.(*ChainDoc)
		method = doc.GetID()
		status := r.getMethodStatus(method)
		if status != expectedStatus {
			return "", nil, "", fmt.Errorf("Method %s is under status: %s, expected status: %s", method, status, expectedStatus)
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
		method = docOption.ID
		docAddr = docOption.Addr
		docHash = docOption.Hash
		status := r.getMethodStatus(method)
		if status != expectedStatus {
			return "", nil, "", fmt.Errorf("Method %s is under status: %s, expected status: %s", method, status, expectedStatus)
		}
	}
	return docAddr, docHash, method, nil
}

// Audit .
// ATN: only admin should call this.
func (r *MethodRegistry) Audit(method DID, status StatusType) error {
	exist := r.HasMethod(method)
	if exist == false {
		return fmt.Errorf("audit %s not existed", method)
	}
	return r.auditStatus(method, status)
}

// Freeze .
// ATN: only admdin should call this.
func (r *MethodRegistry) Freeze(method DID) error {
	exist := r.HasMethod(method)
	if exist == false {
		return fmt.Errorf("freeze %s not existed", method)
	}
	return r.auditStatus(method, Frozen)
}

// UnFreeze .
// ATN: only admdin should call this.
func (r *MethodRegistry) UnFreeze(method DID) error {
	exist := r.HasMethod(method)
	if exist == false {
		return fmt.Errorf("unfreeze %s not existed", method)
	}

	return r.auditStatus(method, Normal)
}

// Delete .
func (r *MethodRegistry) Delete(method DID) error {
	err := r.auditStatus(method, Initial)
	if err != nil {
		return fmt.Errorf("Method delete: %w", err)
	}

	r.Table.DeleteItem(method)

	if r.Mode == InternalDocDB {
		r.Docdb.Delete(method)
	}

	return nil
}

// Resolve looks up local-chain to resolve method.
// @*DIDDoc returns nil if mode is ExternalDocDB
func (r *MethodRegistry) Resolve(method DID) (*MethodItem, *ChainDoc, bool, error) {
	exist := r.HasMethod(method)
	if exist == false {
		return nil, nil, false, nil
	}
	item, err := r.Table.GetItem(method, MethodTableType)
	if err != nil {
		return nil, nil, false, fmt.Errorf("Method resolve table get: %w", err)
	}
	itemM := item.(*MethodItem)

	if r.Mode == InternalDocDB {
		doc, err := r.Docdb.Get(method, ChainDocType)
		if err != nil {
			return itemM, nil, true, fmt.Errorf("Method resolve docdb get: %w", err)
		}
		docM := doc.(*ChainDoc)
		return itemM, docM, true, nil
	}
	return itemM, nil, true, nil
}

// HasMethod .
func (r *MethodRegistry) HasMethod(method DID) bool {
	exist := r.Table.HasItem(method)
	return exist
}

func (r *MethodRegistry) getMethodStatus(method DID) StatusType {
	if !r.Table.HasItem(method) {
		return Initial
	}
	item, err := r.Table.GetItem(method, MethodTableType)
	if err != nil {
		r.logger.Error("method status get item:", err)
		return BadStatus
	}
	itemM := item.(*MethodItem)
	return itemM.Status
}

// auditStatus .
func (r *MethodRegistry) auditStatus(method DID, status StatusType) error {
	item, err := r.Table.GetItem(method, MethodTableType)
	if err != nil {
		return fmt.Errorf("aduitstatus table get: %w", err)
	}
	itemM := item.(*MethodItem)
	itemM.Status = status
	err = r.Table.UpdateItem(itemM)
	if err != nil {
		return fmt.Errorf("aduitstatus table update: %w", err)
	}
	return nil
}
