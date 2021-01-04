package bitxid

import (
	"crypto/sha256"
	"fmt"

	"github.com/meshplus/bitxhub-kit/storage"

	"github.com/sirupsen/logrus"
)

var _ Doc = (*MethodDoc)(nil)

// MethodDoc .
type MethodDoc struct {
	BasicDoc
	Extra []byte `json:"extra"` // for further usage
}

// Marshal .
func (md *MethodDoc) Marshal() ([]byte, error) {
	return Struct2Bytes(md)
}

// Unmarshal .
func (md *MethodDoc) Unmarshal(docBytes []byte) error {
	return Bytes2Struct(docBytes, &md)
}

// GetID .
func (md *MethodDoc) GetID() DID {
	return md.ID
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
	mode          RegistryMode
	isRoot        bool
	admins        []DID
	table         RegistryTable
	docdb         DocDB
	genesisMetohd DID
	genesisDoc    DocOption
	logger        logrus.FieldLogger
}

// NewMethodRegistry news a MethodRegistry
func NewMethodRegistry(ts storage.Storage, l logrus.FieldLogger, options ...func(*MethodRegistry)) (*MethodRegistry, error) {
	rt, _ := NewKVTable(ts)
	db, _ := NewKVDocDB(nil)

	mr := &MethodRegistry{ // default config
		mode:          ExternalDocDB,
		table:         rt,
		docdb:         db,
		logger:        l,
		admins:        []DID{"did:bitxhub:relayroot:0x00000001"},
		isRoot:        true,
		genesisMetohd: "did:bitxhub:relayroot:.",
		genesisDoc: DocOption{
			Content: &MethodDoc{
				BasicDoc: BasicDoc{
					ID:   "did:bitxhub:relayroot:.",
					Type: "method",
					PublicKey: []PubKey{
						{ID: "KEY#1",
							Type:         "Secp256k1",
							PublicKeyPem: "02b97c30de767f084ce3080168ee293053ba33b235d7116a3263d29f1450936b71"},
					},
					Controller: DID("did:bitxhub:relayroot:0x00000001"),
					Authentication: []Auth{
						{PublicKey: []string{"KEY#1"}},
					},
				},
			},
		},
	}

	for _, option := range options {
		option(mr)
	}

	return mr, nil
}

// WithDocStorage .
func WithDocStorage(ds storage.Storage) func(*MethodRegistry) {
	return func(m *MethodRegistry) {
		db, _ := NewKVDocDB(ds)
		m.docdb = db
		m.mode = InternalDocDB
	}
}

// WithAdmin .
func WithAdmin(a DID) func(*MethodRegistry) {
	return func(m *MethodRegistry) {
		m.admins = []DID{a}
	}
}

// WithGenesisMetohd .
func WithGenesisMetohd(m DID) func(*MethodRegistry) {
	return func(mr *MethodRegistry) {
		mr.genesisMetohd = m
	}
}

// WithGenesisDoc .
func WithGenesisDoc(doc *MethodDoc) func(*MethodRegistry) {
	return func(mr *MethodRegistry) {
		mr.genesisDoc = DocOption{Content: doc}
	}
}

// SetupGenesis set up genesis to boot the whole methed system
func (r *MethodRegistry) SetupGenesis() error { // docOption DocOption
	if !r.isRoot {
		return fmt.Errorf("Method genesis registry not root")
	}
	if r.genesisMetohd != r.genesisDoc.Content.(*MethodDoc).ID {
		return fmt.Errorf("Method genesis Method not matched with doc")
	}

	// register method
	err := r.Apply(r.admins[0], r.genesisMetohd)
	if err != nil {
		return fmt.Errorf("Method genesis apply err: %w", err)
	}

	err = r.AuditApply(r.genesisMetohd, true)
	if err != nil {
		return fmt.Errorf("Method genesis audit err: %w", err)
	}

	_, _, err = r.Register(r.genesisDoc)
	if err != nil {
		return fmt.Errorf("Method genesis register err: %w", err)
	}

	return nil
}

// GetSelfID .
func (r *MethodRegistry) GetSelfID() DID {
	return r.genesisMetohd
}

// GetAdmins .
func (r *MethodRegistry) GetAdmins() []DID {
	return r.admins
}

// AddAdmin .
func (r *MethodRegistry) AddAdmin(caller DID) error {
	if r.HasAdmin(caller) {
		return fmt.Errorf("caller %s is already an admin", caller)
	}
	r.admins = append(r.admins, caller)
	return nil
}

// HasAdmin .
func (r *MethodRegistry) HasAdmin(caller DID) bool {
	for _, v := range r.admins {
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
	err := r.table.CreateItem(
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
	return r.table.CreateItem(item)
}

// Register ties method name to a method doc
// ATN: only did who owns method-name should call this
func (r *MethodRegistry) Register(docOption DocOption) (string, []byte, error) { // doc *MethodDoc
	return r.update(docOption, ApplySuccess, Normal)
}

// Update .
// ATN: only did who owns method-name should call this.
func (r *MethodRegistry) Update(docOption DocOption) (string, []byte, error) {
	return r.update(docOption, Normal, Normal)
}

// update with expected status
func (r *MethodRegistry) update(docOption DocOption, expectedStatus StatusType, status StatusType) (string, []byte, error) {
	var docAddr string
	var docHash []byte
	var method DID
	if r.mode == InternalDocDB {
		// check exist
		doc := docOption.Content.(*MethodDoc)
		method = doc.GetID()
		status := r.getMethodStatus(method)
		if status != expectedStatus {
			return "", nil, fmt.Errorf("Method %s is under status: %s, expectd status: %s", method, status, expectedStatus)
		}

		docBytes, err := doc.Marshal()
		if err != nil {
			return "", nil, fmt.Errorf("doc marshal: %w ", err)
		}

		if r.docdb.Has(doc.GetID()) {
			docAddr, err = r.docdb.Update(doc)
		} else {
			docAddr, err = r.docdb.Create(doc)
		}

		if err != nil {
			return "", nil, fmt.Errorf("update docdb: %w ", err)
		}

		docHash32 := sha256.Sum256(docBytes)
		docHash = docHash32[:]
	} else {
		method = docOption.ID
		docAddr = docOption.Addr
		docHash = docOption.Hash
		status := r.getMethodStatus(method)
		if status != expectedStatus {
			return "", nil, fmt.Errorf("Method %s is under status: %s, expectd status: %s", method, status, expectedStatus)
		}
	}

	item, err := r.table.GetItem(method, MethodTableType)
	if err != nil {
		return docAddr, docHash[:], fmt.Errorf("table get item: %w ", err)
	}
	itemM := item.(*MethodItem)
	itemM.DocAddr = docAddr
	itemM.DocHash = docHash
	itemM.Status = status
	err = r.table.UpdateItem(itemM)
	if err != nil {
		return docAddr, docHash, fmt.Errorf("table update item: %w ", err)
	}

	return docAddr, docHash, nil
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

	r.table.DeleteItem(method)

	if r.mode == InternalDocDB {
		r.docdb.Delete(method)
	}

	return nil
}

// Resolve .
func (r *MethodRegistry) Resolve(method DID) (*MethodItem, *MethodDoc, bool, error) {
	exist := r.HasMethod(method)
	if exist == false {
		return nil, nil, false, nil
	}
	item, err := r.table.GetItem(method, MethodTableType)
	if err != nil {
		return nil, nil, false, fmt.Errorf("Method resolve table get: %w", err)
	}
	itemM := item.(*MethodItem)

	docM := &MethodDoc{}
	if r.mode == InternalDocDB {
		doc, err := r.docdb.Get(method, MethodDocType)
		if err != nil {
			return itemM, nil, true, fmt.Errorf("Method resolve docdb get: %w", err)
		}
		docM = doc.(*MethodDoc)
	}
	return itemM, docM, true, nil
}

// HasMethod .
func (r *MethodRegistry) HasMethod(method DID) bool {
	exist := r.table.HasItem(method)
	return exist
}

func (r *MethodRegistry) getMethodStatus(method DID) StatusType {
	if !r.table.HasItem(method) {
		return Initial
	}
	item, err := r.table.GetItem(method, MethodTableType)
	if err != nil {
		r.logger.Error("method status get item:", err)
		return Error
	}
	itemM := item.(*MethodItem)
	return itemM.Status
}

// auditStatus .
func (r *MethodRegistry) auditStatus(method DID, status StatusType) error {
	item, err := r.table.GetItem(method, MethodTableType)
	if err != nil {
		return fmt.Errorf("aduitstatus table get: %w", err)
	}
	itemM := item.(*MethodItem)
	itemM.Status = status
	err = r.table.UpdateItem(itemM)
	if err != nil {
		return fmt.Errorf("aduitstatus table update: %w", err)
	}
	return nil
}
