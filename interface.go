package bitxid

// Doc represents did doc
type Doc interface {
	Marshal() ([]byte, error)
	Unmarshal(docBytes []byte) error
	GetID() DID
	IsValidFormat() bool
}

// TableItem represents the table item of a registry table
type TableItem interface {
	Marshal() ([]byte, error)
	Unmarshal(itemBytes []byte) error
	GetID() DID
}

// DocDB represents did Doc db (used under InternalDocDB mode)
type DocDB interface {
	Create(doc Doc) (string, error)
	Update(doc Doc) (string, error)
	Get(did DID, typ DocType) (Doc, error)
	Delete(did DID)
	Has(did DID) bool
	Close() error
}

// RegistryTable represents state table of a registry
type RegistryTable interface {
	CreateItem(item TableItem) error
	UpdateItem(item TableItem) error
	GetItem(did DID, typ TableType) (TableItem, error)
	HasItem(did DID) bool
	DeleteItem(did DID)
	Close() error
}

type BasicManager interface {
	SetupGenesis() error
	GetSelfID() DID
	GetAdmins() []DID
	AddAdmin(caller DID) error
	RemoveAdmin(caller DID) error
	HasAdmin(caller DID) bool
}

// MethodManager represents chain did management registry
type MethodManager interface {
	BasicManager
	HasMethod(method DID) bool

	Apply(caller DID, method DID) error
	AuditApply(method DID, result bool) error
	Audit(method DID, status StatusType) error
	Register(docOption DocOption) (string, []byte, error)
	Resolve(method DID) (*MethodItem, *ChainDoc, bool, error)
	Update(docOption DocOption) (string, []byte, error)
	Freeze(method DID) error
	UnFreeze(method DID) error
	Delete(method DID) error
}

// DIDManager represents account did management registry
type DIDManager interface {
	BasicManager
	GetMethod() DID
	HasDID(did DID) bool

	Register(docOption DocOption) (string, []byte, error)
	Resolve(did DID) (*DIDItem, *DIDDoc, bool, error)
	Update(docOption DocOption) (string, []byte, error)
	Freeze(did DID) error
	UnFreeze(did DID) error
	Delete(did DID) error
}
