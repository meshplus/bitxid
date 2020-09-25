package docdb

import (
	"errors"

	"github.com/meshplus/bitxhub-kit/log"
	"github.com/meshplus/bitxhub-kit/storage"
	"github.com/meshplus/bitxid/internal/common/types"
)

// DocDB .
type DocDB struct {
	store storage.Storage
}

var logger = log.NewWithModule("doc.DB")
var _ types.DocDB = (*DocDB)(nil)

// NewDB .
func NewDB(S storage.Storage) (*DocDB, error) {
	return &DocDB{
		store: S,
	}, nil
}

// Has whether db has the item(by key)
func (d *DocDB) Has(key []byte) (bool, error) {
	exists, err := d.store.Has(key)
	if err != nil {
		logger.Error("[d.store.Has] err:", err)
		return false, err
	}
	return exists, err
}

// Create .
func (d *DocDB) Create(key, value []byte) error {
	exist, err := d.Has(key)
	if err != nil {
		return err
	}
	if exist == true {
		return errors.New("The key ALREADY existed in doc db")
	}
	err = d.store.Put(key, value)
	if err != nil {
		logger.Error("[d.store.Put] err", err)
		return err
	}
	return nil
}

// Update .
func (d *DocDB) Update(key, value []byte) error {
	exist, err := d.Has(key)
	if err != nil {
		return err
	}
	if exist == false {
		return errors.New("The key NOT existed in doc db")
	}
	err = d.store.Put(key, value)
	if err != nil {
		logger.Error("[d.store.Put] err", err)
		return err
	}
	return nil
}

// Get .
func (d *DocDB) Get(key []byte) (value []byte, err error) {
	exist, err := d.Has(key)
	if err != nil {
		return []byte{}, err
	}
	if exist == false {
		return []byte{}, errors.New("The key NOT existed in doc db")
	}
	value, err = d.store.Get(key)
	if err != nil {
		logger.Error("[d.store.Get] err", err)
		return []byte{}, err
	}
	return value, nil
}

// Delete .
func (d *DocDB) Delete(key []byte) error {
	err := d.store.Delete(key)
	if err != nil {
		logger.Error("[d.store.Delete] err", err)
		return err
	}
	return nil
}
