package bitxid

import (
	"fmt"

	"github.com/meshplus/bitxhub-kit/storage"
)

// KVTable .
type KVTable struct {
	store storage.Storage
}

// var tablelogger = log.NewWithModule("registry.Table")

var _ RegistryTable = (*KVTable)(nil)

// NewKVTable .
func NewKVTable(S storage.Storage) (*KVTable, error) {
	return &KVTable{
		store: S,
	}, nil
}

// HasItem whether table has the item(by key)
func (r *KVTable) HasItem(key DID) (bool, error) {
	exists, err := r.store.Has([]byte(key))
	if err != nil {
		return false, fmt.Errorf("kvtable store: %w", err)
	}
	return exists, err
}

// SetItem sets without any checks
func (r *KVTable) setItem(key DID, item interface{}) error {
	bitem, err := Struct2Bytes(item)
	if err != nil {
		return fmt.Errorf("kvtable marshal: %w", err)
	}
	err = r.store.Put([]byte(key), bitem)
	if err != nil {
		return fmt.Errorf("kvtable store: %w", err)
	}
	return nil
}

// CreateItem checks and sets
func (r *KVTable) CreateItem(key DID, item interface{}) error {
	exist, err := r.HasItem(key)
	if err != nil {
		return err
	}
	if exist == true {
		return fmt.Errorf("Key %s already existed in kvtable", key)
	}
	err = r.setItem(key, item)
	if err != nil {
		return err
	}
	return nil
}

// UpdateItem checks and sets
func (r *KVTable) UpdateItem(key DID, item interface{}) error {
	exist, err := r.HasItem(key)
	if err != nil {
		return err
	}
	if exist == false {
		return fmt.Errorf("Key %s not existed in kvtable", key)
	}
	err = r.setItem(key, item)
	if err != nil {
		return err
	}
	return nil
}

// GetItem checks ang gets
func (r *KVTable) GetItem(key DID, item interface{}) error {
	exist, err := r.HasItem(key)
	if err != nil {
		return err
	}
	if exist == false {
		return fmt.Errorf("Key %s not existed in kvtable", key)
	}
	bitem, err := r.store.Get([]byte(key))
	if err != nil {
		return fmt.Errorf("kvtable store: %w", err)
	}
	err = Bytes2Struct(bitem, item)
	if err != nil {
		return fmt.Errorf("kvtable unmarshal: %w", err)
	}
	return nil
}

// DeleteItem without any checks
func (r *KVTable) DeleteItem(key DID) error {
	err := r.store.Delete([]byte(key))
	if err != nil {
		return fmt.Errorf("kvtable store: %w", err)
	}
	return nil
}

// Close .
func (r *KVTable) Close() error {
	err := r.store.Close()
	if err != nil {
		return fmt.Errorf("kvtable store: %w", err)
	}
	return nil
}
