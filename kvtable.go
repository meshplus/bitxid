package bitxid

import (
	"fmt"

	"github.com/meshplus/bitxhub-kit/log"
	"github.com/meshplus/bitxhub-kit/storage"
)

// KVTable .
type KVTable struct {
	store storage.Storage
}

var tablelogger = log.NewWithModule("registry.Table")

var _ RegistryTable = (*KVTable)(nil)

// NewKVTable .
func NewKVTable(S storage.Storage) (*KVTable, error) {
	return &KVTable{
		store: S,
	}, nil
}

// HasItem whether table has the item(by key)
func (r *KVTable) HasItem(key []byte) (bool, error) {
	exists, err := r.store.Has(key)
	if err != nil {
		tablelogger.Error("r.store.Has err:", err)
		return false, err
	}
	return exists, err
}

// SetItem sets without any checks
func (r *KVTable) setItem(key []byte, item interface{}) error {
	bitem, err := Struct2Bytes(item)
	if err != nil {
		tablelogger.Error("Struct2Bytes err", err)
		return err
	}
	err = r.store.Put(key, bitem)
	if err != nil {
		tablelogger.Error("store.Put err", err)
		return err
	}
	return nil
}

// CreateItem checks and sets
func (r *KVTable) CreateItem(key []byte, item interface{}) error {
	exist, err := r.HasItem(key)
	if err != nil {
		return err
	}
	if exist == true {
		return fmt.Errorf("The key ALREADY existed in registry KVTable")
	}
	err = r.setItem(key, item)
	if err != nil {
		tablelogger.Error("SetItem err", err)
		return err
	}
	return nil
}

// UpdateItem checks and sets
func (r *KVTable) UpdateItem(key []byte, item interface{}) error {
	exist, err := r.HasItem(key)
	if err != nil {
		return err
	}
	if exist == false {
		return fmt.Errorf("The key NOT existed in registry KVTable")
	}
	err = r.setItem(key, item)
	if err != nil {
		tablelogger.Error("SetItem err", err)
		return err
	}
	return nil
}

// GetItem checks ang gets
func (r *KVTable) GetItem(key []byte, item interface{}) error {
	exist, err := r.HasItem(key)
	if err != nil {
		return err
	}
	if exist == false {
		return fmt.Errorf("The key NOT existed in registry KVTable")
	}
	bitem, err := r.store.Get(key)
	if err != nil {
		tablelogger.Error("store.Get err", err)
		return err
	}
	err = Bytes2Struct(bitem, item)
	if err != nil {
		tablelogger.Error("Bytes2Struct err", err)
		return err
	}
	return nil
}

// DeleteItem without any checks
func (r *KVTable) DeleteItem(key []byte) error {
	err := r.store.Delete(key)
	if err != nil {
		tablelogger.Error("r.store.Delete err", err)
		return err
	}
	return nil
}

// Close .
func (r *KVTable) Close() error {
	return r.store.Close()
}
