package bitxid

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/meshplus/bitxhub-kit/storage/leveldb"
	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	A int
	B string
	C []byte
	D []string
	E *subStruct
}
type subStruct struct {
	M int
	N string
	O []byte
	P []string
}

func TestTABLECURD(t *testing.T) {

	dir, err := ioutil.TempDir("testdata", "registry.table")
	assert.Nil(t, err)

	defer os.RemoveAll(dir)

	key := DID("a:b:c:1")
	item := ChainItem{
		BasicItem{ID: key,
			DocAddr: "./abc",
			DocHash: []byte("cde"),
			Status:  Initial},
		"a:b:c:1",
	}
	s, err := leveldb.New(dir)

	assert.Nil(t, err)
	rt, err := NewKVTable(s)
	assert.Nil(t, err)

	// test HasItem:
	ret1 := rt.HasItem(key)
	// assert.Nil(t, err)
	assert.Equal(t, false, ret1)
	// test CreateItem:
	err = rt.CreateItem(&item)
	assert.Nil(t, err)
	// test CreateItem:
	item2, err := rt.GetItem(key, ChainDIDType)
	assert.Nil(t, err)
	assert.Equal(t, item, *item2.(*ChainItem))
	// test
	item3 := ChainItem{
		BasicItem{
			ID:      DID("a:b:c:1"),
			DocAddr: "./abc",
			DocHash: []byte("fgh"),
			Status:  Normal},
		"a:b:c:1",
	}
	err = rt.UpdateItem(&item3)
	assert.Nil(t, err)
	item4, err := rt.GetItem(key, ChainDIDType)
	assert.Nil(t, err)
	assert.Equal(t, item3, *item4.(*ChainItem))
	// test DeleteItem:
	rt.DeleteItem(key)
	// assert.Nil(t, err)
	ret2 := rt.HasItem(key)
	// assert.Nil(t, err)
	assert.Equal(t, false, ret2)
}
