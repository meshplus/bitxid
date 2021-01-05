package bitxid

// Doc .
type Doc interface {
	Marshal() ([]byte, error)
	Unmarshal(docBytes []byte) error
	GetID() DID
}

// TableItem .
type TableItem interface {
	Marshal() ([]byte, error)
	Unmarshal(itemBytes []byte) error
	GetID() DID
}

// DocDB stores info doc for an element
type DocDB interface {
	Create(doc Doc) (string, error)
	Update(doc Doc) (string, error)
	Get(did DID, typ DocType) (Doc, error)
	Delete(did DID)
	Has(did DID) bool
	Close() error
}

// RegistryTable represents state table for a registry
type RegistryTable interface {
	CreateItem(item TableItem) error
	UpdateItem(item TableItem) error
	GetItem(did DID, typ TableType) (TableItem, error)
	HasItem(did DID) bool
	DeleteItem(did DID)
	Close() error
}

// MethodManager .
type MethodManager interface {
	Apply(caller DID, method DID) error
	AuditApply(method DID, result bool) error
	Audit(method DID, status StatusType) error
	Register(docOption DocOption) (string, []byte, error)
	Resolve(method DID) (*MethodItem, *MethodDoc, bool, error)
	Update(docOption DocOption) (string, []byte, error)
	Delete(method DID) error
	HasMethod(method DID) bool
}

// DIDManager .
type DIDManager interface {
	Register(docOption DocOption) (string, []byte, error)
	Resolve(did DID) (*DIDItem, *DIDDoc, bool, error)
	Update(docOption DocOption) (string, []byte, error)
	Delete(did DID) error
	HasDID(did DID) bool
}
