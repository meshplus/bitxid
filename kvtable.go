package bitxid

import (
	"fmt"

	"github.com/meshplus/bitxhub-kit/storage"
)

// KVTable .
type KVTable struct {
	Store storage.Storage `json:"store"`
}

var _ RegistryTable = (*KVTable)(nil)

// NewKVTable .
func NewKVTable(s storage.Storage) (*KVTable, error) {
	return &KVTable{
		Store: s,
	}, nil
}

func tbKey(id DID) []byte {
	return []byte("tb-" + string(id))
}

// HasItem whether table has the item(by key)
func (r *KVTable) HasItem(did DID) bool {
	exists := r.Store.Has(tbKey(did))
	return exists
}

// SetItem sets without any checks
func (r *KVTable) setItem(did DID, item TableItem) error {
	bitem, err := item.Marshal()
	if err != nil {
		return fmt.Errorf("kvtable marshal: %w", err)
	}
	r.Store.Put(tbKey(did), bitem)
	return nil
}

// CreateItem checks and sets
func (r *KVTable) CreateItem(item TableItem) error {
	did := item.GetID()
	if did == DID("") {
		return fmt.Errorf("kvtable create item id is null")
	}
	exist := r.HasItem(did)
	if exist {
		return fmt.Errorf("key %s already existed in kvtable", did)
	}
	return r.setItem(did, item)
}

// UpdateItem checks and sets
func (r *KVTable) UpdateItem(item TableItem) error {
	did := item.GetID()
	if did == DID("") {
		return fmt.Errorf("kvtable create item id is null")
	}
	exist := r.HasItem(did)
	if !exist {
		return fmt.Errorf("key %s not existed in kvtable", did)
	}
	return r.setItem(did, item)
}

// GetItem checks ang gets
func (r *KVTable) GetItem(did DID, typ DIDType) (TableItem, error) {
	exist := r.HasItem(did)
	if !exist {
		return nil, fmt.Errorf("key %s not existed in kvtable", did)
	}
	itemBytes := r.Store.Get(tbKey(did))
	switch typ {
	case AccountDIDType:
		di := &AccountItem{}
		err := di.Unmarshal(itemBytes)
		if err != nil {
			return nil, fmt.Errorf("kvtable unmarshal did item: %w", err)
		}
		return di, nil
	case ChainDIDType:
		mi := &ChainItem{}
		err := mi.Unmarshal(itemBytes)
		if err != nil {
			return nil, fmt.Errorf("kvtable unmarshal method item: %w", err)
		}
		return mi, nil
	default:
		return nil, fmt.Errorf("kvtable unknown table type: %d", typ)
	}
}

// DeleteItem without any checks
func (r *KVTable) DeleteItem(did DID) {
	r.Store.Delete(tbKey(did))
}

// Close .
func (r *KVTable) Close() error {
	err := r.Store.Close()
	if err != nil {
		return fmt.Errorf("kvtable store: %w", err)
	}
	return nil
}
