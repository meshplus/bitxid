package bitxid

// Doc .
type Doc interface {
	Marshal() ([]byte, error)
	Unmarshal(docBytes []byte) error
}

// DocDB stores info doc for an element
type DocDB interface {
	Create(key DID, value Doc) (string, error)
	Update(key DID, value Doc) (string, error)
	Get(key DID, typ int) (Doc, error)
	Delete(key DID) error
	Has(key DID) (bool, error)
	Close() error
}

// RegistryTable represents state table for a registry
type RegistryTable interface {
	CreateItem(key DID, item interface{}) error
	UpdateItem(key DID, item interface{}) error
	GetItem(key DID, item interface{}) error
	HasItem(key DID) (bool, error)
	DeleteItem(key DID) error
	Close() error
}

// MethodManager .
type MethodManager interface {
	Apply(caller DID, method DID) error
	AuditApply(method DID, result bool) error
	Audit(method DID, status int) error
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
