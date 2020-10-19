package docdb

import (
	"testing"

	"github.com/meshplus/bitxhub-kit/storage/leveldb"
	"github.com/stretchr/testify/assert"
)

const dbPath string = "../../../config/doc.db"

func TestCURD(t *testing.T) {
	key := []byte("4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b")
	value := []byte("The Times 03/Jan/2009 Chancellor on brink of second bailout for banks.")
	valueUpdated := []byte("=_=_=_=_=_=_=_=_=_=_=")
	s, err := leveldb.New(dbPath)
	assert.Nil(t, err)
	d, err := NewDB(s)
	assert.Nil(t, err)
	// test create:
	ret1, err := d.Create(key, value)
	assert.Nil(t, err)
	assert.Equal(t, "./"+string(key), ret1)
	// test has:
	ret2, err := d.Has(key)
	assert.Nil(t, err)
	assert.Equal(t, true, ret2)
	// test get:
	ret3, err := d.Get(key)
	assert.Nil(t, err)
	assert.Equal(t, ret3, value)
	// test update:
	ret4, err := d.Update(key, valueUpdated)
	assert.Nil(t, err)
	assert.Equal(t, "./"+string(key), ret4)
	ret5, err := d.Get(key)
	assert.Nil(t, err)
	assert.Equal(t, ret5, valueUpdated)
	// test delete:
	err = d.Delete(key)
	assert.Nil(t, err)
	ret6, err := d.Has(key)
	assert.Nil(t, err)
	assert.Equal(t, false, ret6)
}
