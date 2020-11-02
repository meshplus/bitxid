package bitxid

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/meshplus/bitxhub-kit/storage/leveldb"
	"github.com/stretchr/testify/assert"
)

func TestDBCURD(t *testing.T) {
	dir, err := ioutil.TempDir("testdata", "doc.db")
	assert.Nil(t, err)

	defer os.RemoveAll(dir)

	key := DID("did:bitxhub:appchain001:.")
	value := DIDDoc{
		BasicDoc: BasicDoc{ID: "did:bitxhub:appchain001:."},
	}
	valueUpdated := DIDDoc{
		BasicDoc: BasicDoc{ID: "did:bitxhub:appchain001:."},
		Service:  "test",
	}
	// dbPath := filepath.Join(dir, "docdb")
	s, err := leveldb.New(dir)
	assert.Nil(t, err)
	d, err := NewKVDocDB(s)
	assert.Nil(t, err)
	// test create:
	ret1, err := d.Create(key, &value)
	assert.Nil(t, err)
	assert.Equal(t, "./"+string(key), ret1)
	// test has:
	ret2, err := d.Has(key)
	assert.Nil(t, err)
	assert.Equal(t, true, ret2)
	// test get:
	ret3, err := d.Get(key, DIDDocType)
	assert.Nil(t, err)
	assert.Equal(t, *ret3.(*DIDDoc), value)
	// test update:
	ret4, err := d.Update(key, &valueUpdated)
	assert.Nil(t, err)
	assert.Equal(t, "./"+string(key), ret4)
	ret5, err := d.Get(key, DIDDocType)
	assert.Nil(t, err)
	assert.Equal(t, *ret5.(*DIDDoc), valueUpdated)
	// test delete:
	err = d.Delete(key)
	assert.Nil(t, err)
	ret6, err := d.Has(key)
	assert.Nil(t, err)
	assert.Equal(t, false, ret6)
}
