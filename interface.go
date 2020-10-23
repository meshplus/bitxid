package bitxid

// import "github.com/bitxhub/bitxid/pkg/common/types"

// DocDB stores info doc for an element
type DocDB interface {
	Create(key, value []byte) (string, error)
	Update(key, value []byte) (string, error)
	Get(key []byte) (value []byte, err error)
	Delete(key []byte) error
	Has(key []byte) (bool, error)
	Close() error
}

// RegistryTable represents state table for a registry
type RegistryTable interface {
	CreateItem(key []byte, item interface{}) error
	UpdateItem(key []byte, item interface{}) error
	GetItem(key []byte, item interface{}) (err error)
	HasItem(key []byte) (bool, error)
	DeleteItem(key []byte) error
	Close() error
}

// MethodManager .
type MethodManager interface {
	GetAdmins() []DID
	AddAdmin(caller DID) error
	HasAdmin(caller DID) bool
	Apply(caller DID, method DID) error
	AuditApply(method DID, result bool) error
	Audit(method DID, status int) error
	Register(doc MethodDoc) (string, []byte, error)
	Resolve(method DID) (MethodItem, MethodDoc, error)
	Update(doc MethodDoc) (string, []byte, error)
	Freeze(method DID) error
	UnFreeze(method DID) error
	Delete(method DID) error
	HasMethod(method DID) (bool, error)
}

// DIDManager .
type DIDManager interface {
	GetAdmins() []DID
	AddAdmin(caller DID) error
	HasAdmin(caller DID) bool
	GetMethod() DID
	Register(doc DIDDoc) (string, []byte, error)
	Resolve(did DID) (DIDItem, DIDDoc, error)
	Update(doc DIDDoc) (string, []byte, error)
	Freeze(did DID) error
	UnFreeze(did DID) error
	Delete(did DID) error
	HasDID(did DID) (bool, error)
}
