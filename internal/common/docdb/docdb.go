package docdb

import (
	"fmt"

	"github.com/meshplus/bitxhub-kit/log"
	"github.com/meshplus/bitxhub-kit/storage"
	"github.com/meshplus/bitxid/internal/common/types"
)

// DocDB .
type DocDB struct {
	basicAddr string
	store     storage.Storage
}

var logger = log.NewWithModule("doc.DB")
var _ types.DocDB = (*DocDB)(nil)

// NewDB .
func NewDB(S storage.Storage) (*DocDB, error) {
	return &DocDB{
		store:     S,
		basicAddr: ".",
	}, nil
}

// Has whether db has the item(by key)
func (d *DocDB) Has(key []byte) (bool, error) {
	exists, err := d.store.Has(key)
	if err != nil {
		logger.Error("d.store.Has err:", err)
		return false, err
	}
	return exists, err
}

// Create .
func (d *DocDB) Create(key, value []byte) (string, error) {
	exist, err := d.Has(key)
	if err != nil {
		return "", err
	}
	if exist == true {
		return "", fmt.Errorf("The key ALREADY existed in doc db")
	}
	err = d.store.Put(key, value)
	if err != nil {
		logger.Error("d.store.Put err", err)
		return "", err
	}
	return d.basicAddr + "/" + string(key), nil
}

// Update .
func (d *DocDB) Update(key, value []byte) (string, error) {
	exist, err := d.Has(key)
	if err != nil {
		return "", err
	}
	if exist == false {
		return "", fmt.Errorf("The key NOT existed in doc db")
	}
	err = d.store.Put(key, value)
	if err != nil {
		logger.Error("d.store.Put err", err)
		return "", err
	}
	return d.basicAddr + "/" + string(key), nil
}

// Get .
func (d *DocDB) Get(key []byte) (value []byte, err error) {
	exist, err := d.Has(key)
	if err != nil {
		return []byte{}, err
	}
	if exist == false {
		return []byte{}, fmt.Errorf("The key NOT existed in doc db")
	}
	value, err = d.store.Get(key)
	if err != nil {
		logger.Error("d.store.Get err", err)
		return []byte{}, err
	}
	return value, nil
}

// Delete .
func (d *DocDB) Delete(key []byte) error {
	err := d.store.Delete(key)
	if err != nil {
		logger.Error("d.store.Delete err", err)
		return err
	}
	return nil
}

// Close .
func (d *DocDB) Close() error {
	return d.store.Close()
}
