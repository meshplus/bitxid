package types

// DocStorage stores info-doc for an element
type DocDB interface {
	Create(key, value []byte) error
	Update(key, value []byte) error
	Get(key []byte) (value []byte, err error)
	Delete(key []byte) error
	Has(key []byte) (bool, error)
}

// RegistryTable represents state table for a registry
type RegistryTable interface {
	CreateItem(key []byte, item interface{}) error
	UpdateItem(key []byte, item interface{}) error
	GetItem(key []byte, item interface{}) (err error)
	HasItem(key []byte) (bool, error)
	DeleteItem(key []byte) error
}

type RegistryNetwork interface {
}
