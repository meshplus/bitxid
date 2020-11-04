package bitxid

import (
	"fmt"

	"github.com/meshplus/bitxhub-kit/storage"
)

// KVDocDB .
type KVDocDB struct {
	basicAddr string
	store     storage.Storage
}

// var dblogger = log.NewWithModule("doc.DB")
var _ DocDB = (*KVDocDB)(nil)

// NewKVDocDB .
func NewKVDocDB(S storage.Storage) (*KVDocDB, error) {
	return &KVDocDB{
		store:     S,
		basicAddr: ".",
	}, nil
}

// Has whether db has the item(by key)
func (d *KVDocDB) Has(did DID) (bool, error) {
	return d.store.Has([]byte(did))
}

// Create .
func (d *KVDocDB) Create(doc Doc) (string, error) {
	did := doc.GetID()
	if did == DID("") {
		return "", fmt.Errorf("kvdb create doc id is null")
	}
	exist, err := d.Has(did)
	if err != nil {
		return "", err
	}
	if exist == true {
		return "", fmt.Errorf("Item %s already existed in kvdb", did)
	}
	valueBytes, err := doc.Marshal()
	if err != nil {
		return "", err
	}
	err = d.store.Put([]byte(did), valueBytes)
	if err != nil {
		return "", fmt.Errorf("kvdb store: %w", err)
	}
	return d.basicAddr + "/" + string(did), nil
}

// Update .
func (d *KVDocDB) Update(doc Doc) (string, error) {
	did := doc.GetID()
	if did == DID("") {
		return "", fmt.Errorf("kvdb update doc id is null")
	}
	exist, err := d.Has(did)
	if err != nil {
		return "", err
	}
	if exist == false {
		return "", fmt.Errorf("Item %s not existed in kvdb", did)
	}
	valueBytes, err := doc.Marshal()
	if err != nil {
		return "", err
	}
	err = d.store.Put([]byte(did), valueBytes)
	if err != nil {
		return "", fmt.Errorf("kvdb store: %w", err)
	}
	return d.basicAddr + "/" + string(did), nil
}

// Get .
func (d *KVDocDB) Get(did DID, typ DocType) (Doc, error) {
	exist, err := d.Has(did)
	if err != nil {
		return nil, err
	}
	if exist == false {
		return nil, fmt.Errorf("Key %s not existed in kvdb", did)
	}
	valueBytes, err := d.store.Get([]byte(did))
	if err != nil {
		return nil, fmt.Errorf("kvdb store: %w", err)
	}
	switch typ {
	case DIDDocType:
		dt := &DIDDoc{}
		err = dt.Unmarshal(valueBytes)
		if err != nil {
			return nil, fmt.Errorf("kvdb unmarshal did doc: %w", err)
		}
		return dt, nil
	case MethodDocType:
		mt := &MethodDoc{}
		err = mt.Unmarshal(valueBytes)
		if err != nil {
			return nil, fmt.Errorf("kvdb unmarshal method doc: %w", err)
		}
		return mt, nil
	default:
		return nil, fmt.Errorf("kvdb unknown doc type: %d", typ)
	}
}

// Delete .
func (d *KVDocDB) Delete(did DID) error {
	err := d.store.Delete([]byte(did))
	if err != nil {
		return fmt.Errorf("kvdb store: %w", err)
	}
	return nil
}

// Close .
func (d *KVDocDB) Close() error {
	err := d.store.Close()
	if err != nil {
		return fmt.Errorf("kvdb store: %w", err)
	}
	return nil
}
