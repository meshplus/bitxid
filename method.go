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
	Parent string `json:"parent"`
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

var _ MethodManager = (*MethodRegistry)(nil)

// MethodRegistry .
type MethodRegistry struct {
	config *MethodConfig
	table  RegistryTable
	docdb  DocDB
	logger logrus.FieldLogger
	admins []DID // admins of the registry
	// network
}

var _ TableItem = (*MethodItem)(nil)

// MethodItem reperesents a method item, element of registry table.
// Registry table is used together with docdb.
type MethodItem struct {
	BasicItem
	Owner DID // owner of the method, is a did
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

// NewMethodRegistry news a MethodRegistry
func NewMethodRegistry(ts storage.Storage, ds storage.Storage, l logrus.FieldLogger, options ...func(*MethodConfig)) (*MethodRegistry, error) {
	rt, err := NewKVTable(ts)
	if err != nil {
		return nil, fmt.Errorf("Method new table: %w", err)
	}
	db, err := NewKVDocDB(ds)
	if err != nil {
		return nil, fmt.Errorf("Method new docdb: %w", err)
	}

	conf := &MethodConfig{
		Admin:         "did:bitxhub:relayroot:0x00000001",
		Addr:          ".",
		IsRoot:        true,
		GenesisMetohd: "did:bitxhub:relayroot:.",
		GenesisDoc: &MethodDoc{
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
	}
	for _, option := range options {
		option(conf)
	}

	return &MethodRegistry{
		table:  rt,
		docdb:  db,
		logger: l,
		config: conf,
		admins: []DID{conf.Admin},
	}, nil
}

// WithAdmin .
func WithAdmin(a DID) func(*MethodConfig) {
	return func(s *MethodConfig) {
		s.Admin = a
	}
}

// WithGenesisMetohd .
func WithGenesisMetohd(m DID) func(*MethodConfig) {
	return func(s *MethodConfig) {
		s.GenesisMetohd = m
	}
}

// WithGenesisDoc .
func WithGenesisDoc(doc *MethodDoc) func(*MethodConfig) {
	return func(s *MethodConfig) {
		s.GenesisDoc = doc
	}
}

// SetupGenesis set up genesis to boot the whole methed system
func (r *MethodRegistry) SetupGenesis() error {
	if !r.config.IsRoot {
		return fmt.Errorf("Method genesis registry not root")
	}
	if r.config.GenesisMetohd != r.config.GenesisDoc.ID {
		return fmt.Errorf("Method genesis Method not matched with doc")
	}
	// register method
	err := r.Apply(r.config.Admin, r.config.GenesisMetohd)
	if err != nil {
		return fmt.Errorf("Method genesis err: %w", err)
	}

	err = r.AuditApply(r.config.GenesisMetohd, true)
	if err != nil {
		return fmt.Errorf("Method genesis err: %w", err)
	}
	r.logger.Info()
	_, _, err = r.Register(r.config.GenesisDoc)
	if err != nil {
		return fmt.Errorf("Method genesis err: %w", err)
	}
	// add admins did
	r.AddAdmin(DID(r.config.Admin))

	return nil
}

// GetSelfID .
func (r *MethodRegistry) GetSelfID() DID {
	return r.config.GenesisMetohd
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
		return fmt.Errorf("can not apply Method under status: %d", status)
	}
	// creates item in table
	err := r.table.CreateItem(
		&MethodItem{
			BasicItem{
				ID:     DID(method),
				Status: ApplyAudit},
			caller})
	if err != nil {
		return fmt.Errorf("apply Method on table: %w", err)
	}
	return nil
}

// AuditApply .
// ATNS: only admin should call this.
func (r *MethodRegistry) AuditApply(method DID, result bool) error {
	exist := r.HasMethod(method)
	if exist == false {
		return fmt.Errorf("auditapply Method %s not existed", method)
	}
	status := r.getMethodStatus(method)
	if !(status == ApplyAudit || status == ApplyFailed) {
		return fmt.Errorf("can not auditapply under status: %d", status)
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
func (r *MethodRegistry) Register(doc *MethodDoc) (string, []byte, error) {
	method := DID(doc.ID)
	status := r.getMethodStatus(method)
	if status != ApplySuccess {
		return "", nil, fmt.Errorf("can not register under status, %d", status)
	}

	docBytes, err := doc.Marshal()
	if err != nil {
		return "", nil, fmt.Errorf("Method register doc marshal, %w", err)
	}

	docAddr, err := r.docdb.Create(doc)
	if err != nil {
		return "", nil, fmt.Errorf("Method register on docdb, %w", err)
	}
	docHash := sha256.Sum256(docBytes)
	// update MethodRegistry table
	item, err := r.table.GetItem(method, MethodTableType)
	if err != nil {
		return "", nil, fmt.Errorf("Method register table get, %w", err)
	}
	itemM := item.(*MethodItem)
	itemM.Status = Normal
	itemM.DocAddr = docAddr
	itemM.DocHash = docHash[:]
	err = r.table.UpdateItem(itemM)
	if err != nil {
		return docAddr, itemM.DocHash, fmt.Errorf("Method register table update: %w", err)
	}
	return docAddr, itemM.DocHash, nil
}

// Update .
// ATN: only did who owns method-name should call this.
func (r *MethodRegistry) Update(doc *MethodDoc) (string, []byte, error) {
	// check exist
	method := DID(doc.ID)
	status := r.getMethodStatus(method)
	if status != Normal {
		return "", nil, fmt.Errorf("can not update under status: %d", status)
	}

	docBytes, err := doc.Marshal()
	if err != nil {
		return "", nil, fmt.Errorf("Method update doc marshal: %w", err)
	}

	docAddr, err := r.docdb.Update(doc)
	if err != nil {
		return "", nil, fmt.Errorf("Method update on docdb: %w", err)
	}
	docHash := sha256.Sum256(docBytes)

	item, err := r.table.GetItem(method, MethodTableType)
	if err != nil {
		return docAddr, docHash[:], fmt.Errorf("Method update table get: %w", err)
	}
	itemM := item.(*MethodItem)
	itemM.DocAddr = docAddr
	itemM.DocHash = docHash[:]
	err = r.table.UpdateItem(itemM)
	if err != nil {
		return docAddr, docHash[:], fmt.Errorf("Method update table update: %w", err)
	}

	return docAddr, docHash[:], nil
}

// Audit .
// ATN: only admin should call this.
func (r *MethodRegistry) Audit(method DID, status StatusType) error {
	exist := r.HasMethod(method)
	if exist == false {
		return fmt.Errorf("audit Method %s not existed", method)
	}
	return r.auditStatus(method, status)
}

// Freeze .
// ATN: only admdin should call this.
func (r *MethodRegistry) Freeze(method DID) error {
	exist := r.HasMethod(method)
	if exist == false {
		return fmt.Errorf("freeze Method %s not existed", method)
	}
	return r.auditStatus(method, Frozen)
}

// UnFreeze .
// ATN: only admdin should call this.
func (r *MethodRegistry) UnFreeze(method DID) error {
	exist := r.HasMethod(method)
	if exist == false {
		return fmt.Errorf("unfreeze Method %s not existed", method)
	}

	return r.auditStatus(method, Normal)
}

// Delete .
func (r *MethodRegistry) Delete(method DID) error {
	err := r.auditStatus(method, Initial)
	if err != nil {
		return fmt.Errorf("Method delete status aduit: %w", err)
	}
	r.docdb.Delete(method)
	r.table.DeleteItem(method)
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
	doc, err := r.docdb.Get(method, MethodDocType)
	if err != nil {
		return itemM, nil, true, fmt.Errorf("Method resolve docdb get: %w", err)
	}
	docM := doc.(*MethodDoc)
	return itemM, docM, true, nil
}

// MethodHasAccount checks whether account exists on the method blockchain
func (r *MethodRegistry) MethodHasAccount(method string, account string) {

}

// HasMethod .
func (r *MethodRegistry) HasMethod(method DID) bool {
	exist := r.table.HasItem(method)
	return exist
}

func (r *MethodRegistry) getMethodStatus(method DID) StatusType {
	item, err := r.table.GetItem(method, MethodTableType)
	if err != nil {
		r.logger.Warn("method status get item:", err)
		return Initial
	}
	itemM := item.(*MethodItem)
	return itemM.Status
}

// auditStatus .
func (r *MethodRegistry) auditStatus(method DID, status StatusType) error {
	item, err := r.table.GetItem(method, MethodTableType)
	if err != nil {
		return fmt.Errorf("Method aduit status table get: %w", err)
	}
	itemM := item.(*MethodItem)
	itemM.Status = status
	err = r.table.UpdateItem(itemM)
	if err != nil {
		return fmt.Errorf("Method aduit status table update: %w", err)
	}
	return nil
}
