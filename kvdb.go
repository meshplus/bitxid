package bitxid

import (
	"fmt"

	"github.com/meshplus/bitxhub-kit/storage"
)

// KVDocDB .
type KVDocDB struct {
	BasicAddr string          `json:"basic_addr"`
	Store     storage.Storage `json:"store"`
}

var _ DocDB = (*KVDocDB)(nil)

// NewKVDocDB .
func NewKVDocDB(s storage.Storage) (*KVDocDB, error) {
	return &KVDocDB{
		Store:     s,
		BasicAddr: ".",
	}, nil
}

func docKey(id DID) []byte {
	return []byte("doc-" + string(id))
}

// Has whether db has the item(by key)
func (d *KVDocDB) Has(did DID) bool {
	return d.Store.Has(docKey(did))
}

// Create .
func (d *KVDocDB) Create(doc Doc) (string, error) {
	did := doc.GetID()
	if did == DID("") {
		return "", fmt.Errorf("kvdb create doc id is null")
	}
	exist := d.Has(did)
	if exist {
		return "", fmt.Errorf("item %s already existed in kvdb", did)
	}
	valueBytes, err := doc.Marshal()
	if err != nil {
		return "", err
	}
	d.Store.Put(docKey(did), valueBytes)
	return d.BasicAddr + "/" + string(did), nil
}

// Update .
func (d *KVDocDB) Update(doc Doc) (string, error) {
	did := doc.GetID()
	if did == DID("") {
		return "", fmt.Errorf("kvdb update doc id is null")
	}
	exist := d.Has(did)
	if !exist {
		return "", fmt.Errorf("item %s not existed in kvdb", did)
	}
	valueBytes, err := doc.Marshal()
	if err != nil {
		return "", err
	}
	d.Store.Put(docKey(did), valueBytes)
	return d.BasicAddr + "/" + string(did), nil
}

// Get .
func (d *KVDocDB) Get(did DID, typ DIDType) (Doc, error) {
	exist := d.Has(did)
	if !exist {
		return nil, fmt.Errorf("key %s not existed in kvdb", did)
	}
	valueBytes := d.Store.Get(docKey(did))
	switch typ {
	case AccountDIDType:
		dt := &AccountDoc{}
		err := dt.Unmarshal(valueBytes)
		if err != nil {
			return nil, fmt.Errorf("kvdb unmarshal did doc: %w", err)
		}
		return dt, nil
	case ChainDIDType:
		mt := &ChainDoc{}
		err := mt.Unmarshal(valueBytes)
		if err != nil {
			return nil, fmt.Errorf("kvdb unmarshal method doc: %w", err)
		}
		return mt, nil
	default:
		return nil, fmt.Errorf("kvdb unknown doc type: %d", typ)
	}
}

// Delete .
func (d *KVDocDB) Delete(did DID) {
	d.Store.Delete(docKey(did))
}

// Close .
func (d *KVDocDB) Close() error {
	err := d.Store.Close()
	if err != nil {
		return fmt.Errorf("kvdb store: %w", err)
	}
	return nil
}
