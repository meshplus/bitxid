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
func (d *KVDocDB) Has(key []byte) (bool, error) {
	exists, err := d.store.Has(key)
	if err != nil {
		dblogger.Error("d.store.Has err:", err)
		return false, err
	}
	return exists, err
}

// Create .
func (d *KVDocDB) Create(key, value []byte) (string, error) {
	exist, err := d.Has(key)
	if err != nil {
		return "", err
	}
	if exist == true {
		return "", fmt.Errorf("The key ALREADY existed in doc db")
	}
	err = d.store.Put(key, value)
	if err != nil {
		dblogger.Error("d.store.Put err", err)
		return "", err
	}
	return d.basicAddr + "/" + string(key), nil
}

// Update .
func (d *KVDocDB) Update(key, value []byte) (string, error) {
	exist, err := d.Has(key)
	if err != nil {
		return "", err
	}
	if exist == false {
		return "", fmt.Errorf("The key NOT existed in doc db")
	}
	err = d.store.Put(key, value)
	if err != nil {
		dblogger.Error("d.store.Put err", err)
		return "", err
	}
	return d.basicAddr + "/" + string(key), nil
}

// Get .
func (d *KVDocDB) Get(key []byte) (value []byte, err error) {
	exist, err := d.Has(key)
	if err != nil {
		return []byte{}, err
	}
	if exist == false {
		return []byte{}, fmt.Errorf("The key NOT existed in doc db")
	}
	value, err = d.store.Get(key)
	if err != nil {
		dblogger.Error("d.store.Get err", err)
		return []byte{}, err
	}
	return value, nil
}

// Delete .
func (d *KVDocDB) Delete(key []byte) error {
	err := d.store.Delete(key)
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
