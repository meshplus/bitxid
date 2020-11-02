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
func (d *KVDocDB) Has(key DID) (bool, error) {
	return d.store.Has([]byte(key))
}

// Create .
func (d *KVDocDB) Create(key DID, value Doc) (string, error) {
	exist, err := d.Has(key)
	if err != nil {
		return "", err
	}
	if exist == true {
		return "", fmt.Errorf("Key %s already existed in kvdb", key)
	}
	valueByte, err := value.Marshal()
	if err != nil {
		return "", err
	}
	err = d.store.Put([]byte(key), valueByte)
	if err != nil {
		return "", fmt.Errorf("kvdb store: %w", err)
	}
	return d.basicAddr + "/" + string(key), nil
}

// Update .
func (d *KVDocDB) Update(key DID, value Doc) (string, error) {
	exist, err := d.Has(key)
	if err != nil {
		return "", err
	}
	if exist == false {
		return "", fmt.Errorf("Key %s not existed in kvdb", key)
	}
	valueBytes, err := value.Marshal()
	if err != nil {
		return "", err
	}
	err = d.store.Put([]byte(key), valueBytes)
	if err != nil {
		return "", fmt.Errorf("kvdb store: %w", err)
	}
	return d.basicAddr + "/" + string(key), nil
}

// Get .
func (d *KVDocDB) Get(key DID, typ int) (Doc, error) {
	exist, err := d.Has(key)
	if err != nil {
		return nil, err
	}
	if exist == false {
		return nil, fmt.Errorf("Key %s not existed in kvdb", key)
	}
	valueBytes, err := d.store.Get([]byte(key))
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
func (d *KVDocDB) Delete(key DID) error {
	err := d.store.Delete([]byte(key))
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
