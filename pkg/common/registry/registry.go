package registry

import (
	"fmt"

	"github.com/bitxhub/bitxid/pkg/common/types"
	"github.com/bitxhub/bitxid/pkg/common/utils"

	"github.com/meshplus/bitxhub-kit/log"
	"github.com/meshplus/bitxhub-kit/storage"
)

// Table .
type Table struct {
	store storage.Storage
}

var logger = log.NewWithModule("registry.Table")

var _ types.RegistryTable = (*Table)(nil)

// NewTable .
func NewTable(S storage.Storage) (*Table, error) {
	return &Table{
		store: S,
	}, nil
}

// HasItem whether table has the item(by key)
func (r *Table) HasItem(key []byte) (bool, error) {
	exists, err := r.store.Has(key)
	if err != nil {
		logger.Error("r.store.Has err:", err)
		return false, err
	}
	return exists, err
}

// SetItem sets without any checks
func (r *Table) setItem(key []byte, item interface{}) error {
	bitem, err := utils.Struct2Bytes(item)
	if err != nil {
		logger.Error("utils.Struct2Bytes err", err)
		return err
	}
	err = r.store.Put(key, bitem)
	if err != nil {
		logger.Error("store.Put err", err)
		return err
	}
	return nil
}

// CreateItem checks and sets
func (r *Table) CreateItem(key []byte, item interface{}) error {
	exist, err := r.HasItem(key)
	if err != nil {
		return err
	}
	if exist == true {
		return fmt.Errorf("The key ALREADY existed in registry table")
	}
	err = r.setItem(key, item)
	if err != nil {
		logger.Error("SetItem err", err)
		return err
	}
	return nil
}

// UpdateItem checks and sets
func (r *Table) UpdateItem(key []byte, item interface{}) error {
	exist, err := r.HasItem(key)
	if err != nil {
		return err
	}
	if exist == false {
		return fmt.Errorf("The key NOT existed in registry table")
	}
	err = r.setItem(key, item)
	if err != nil {
		logger.Error("SetItem err", err)
		return err
	}
	return nil
}

// GetItem checks ang gets
func (r *Table) GetItem(key []byte, item interface{}) error {
	exist, err := r.HasItem(key)
	if err != nil {
		return err
	}
	if exist == false {
		return fmt.Errorf("The key NOT existed in registry table")
	}
	bitem, err := r.store.Get(key)
	if err != nil {
		logger.Error("store.Get err", err)
		return err
	}
	err = utils.Bytes2Struct(bitem, item)
	if err != nil {
		logger.Error("utils.Bytes2Struct err", err)
		return err
	}
	return nil
}

// DeleteItem without any checks
func (r *Table) DeleteItem(key []byte) error {
	err := r.store.Delete(key)
	if err != nil {
		logger.Error("r.store.Delete err", err)
		return err
	}
	return nil
}

// Close .
func (r *Table) Close() error {
	return r.store.Close()
}
