package bitxid

import (
	"fmt"

	"github.com/meshplus/bitxhub-kit/log"
	"github.com/meshplus/bitxhub-kit/storage"
)

// KVDocDB .
type KVDocDB struct {
	basicAddr string
	store     storage.Storage
}

var dblogger = log.NewWithModule("doc.DB")
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
	exists, err := d.store.Has([]byte(key))
	if err != nil {
		dblogger.Error("d.store.Has err:", err)
		return false, err
	}
	return exists, err
}

// Create .
func (d *KVDocDB) Create(key DID, value Doc) (string, error) {
	exist, err := d.Has(key)
	if err != nil {
		return "", err
	}
	if exist == true {
		return "", fmt.Errorf("The key ALREADY existed in doc db")
	}
	valueByte, err := value.Marshal()
	if err != nil {
		return "", err
	}
	err = d.store.Put([]byte(key), valueByte)
	if err != nil {
		dblogger.Error("d.store.Put err", err)
		return "", err
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
		return "", fmt.Errorf("The key NOT existed in doc db")
	}
	valueBytes, err := value.Marshal()
	if err != nil {
		return "", err
	}
	err = d.store.Put([]byte(key), valueBytes)
	if err != nil {
		dblogger.Error("d.store.Put err", err)
		return "", err
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
		return nil, fmt.Errorf("The key NOT existed in doc db")
	}
	valueBytes, err := d.store.Get([]byte(key))
	if err != nil {
		dblogger.Error("d.store.Get err", err)
		return nil, err
	}
	switch typ {
	case DIDDocType:
		dt := &DIDDoc{}
		dt.Unmarshal(valueBytes)
		if err != nil {
			dblogger.Error("value.Unmarshal err", err)
			return nil, err
		}
		return dt, nil
	case MethodDocType:
		mt := &MethodDoc{}
		mt.Unmarshal(valueBytes)
		if err != nil {
			dblogger.Error("value.Unmarshal err", err)
			return nil, err
		}
		return mt, nil
	default:
		return nil, fmt.Errorf("unknown Doc type %d", typ)
	}
	// fmt.Println("[get] value Doc:", value)
}

// Delete .
func (d *KVDocDB) Delete(key DID) error {
	err := d.store.Delete([]byte(key))
	if err != nil {
		dblogger.Error("d.store.Delete err", err)
		return err
	}
	return nil
}

// Close .
func (d *KVDocDB) Close() error {
	return d.store.Close()
}
