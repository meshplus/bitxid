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
	Unmarshal(docBytes []byte) error
	GetID() DID
}

// DocDB stores info doc for an element
type DocDB interface {
	Create(doc Doc) (string, error)
	Update(doc Doc) (string, error)
	Get(did DID, typ DocType) (Doc, error)
	Delete(did DID) error
	Has(did DID) (bool, error)
	Close() error
}

// RegistryTable represents state table for a registry
type RegistryTable interface {
	CreateItem(item TableItem) error
	UpdateItem(item TableItem) error
	GetItem(did DID, typ TableType) (TableItem, error)
	HasItem(did DID) (bool, error)
	DeleteItem(did DID) error
	Close() error
}

// MethodManager .
type MethodManager interface {
	Apply(caller DID, method DID) error
	AuditApply(method DID, result bool) error
	Audit(method DID, status StatusType) error
	Register(doc MethodDoc) (string, []byte, error)
	Resolve(method DID) (MethodItem, MethodDoc, error)
	Update(doc MethodDoc) (string, []byte, error)
	Delete(method DID) error
	HasMethod(method DID) (bool, error)
}

// DIDManager .
type DIDManager interface {
	Register(doc DIDDoc) (string, []byte, error)
	Resolve(did DID) (DIDItem, DIDDoc, error)
	Update(doc DIDDoc) (string, []byte, error)
	Delete(did DID) error
	HasDID(did DID) (bool, error)
}
