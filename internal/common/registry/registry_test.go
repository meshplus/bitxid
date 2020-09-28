package registry

import (
	"testing"

	"github.com/meshplus/bitxhub-kit/storage/leveldb"
	"github.com/stretchr/testify/assert"
)

const rtPath string = "../../../config/registry.store"

type TestStruct struct {
	A int
	B string
	C []byte
	D []string
	E *SubStruct
}
type SubStruct struct {
	M int
	N string
	O []byte
	P []string
}

func TestCURD(t *testing.T) {
	key := []byte("did:bitxhub001:appchain1:.")
	item := TestStruct{
		A: 1,
		B: "abc",
		C: []byte("cde"),
		D: []string{"f", "g", "high"},
	}
	s, err := leveldb.New(rtPath)
	assert.Nil(t, err)
	rt, err := NewTable(s)
	assert.Nil(t, err)

	// test HasItem:
	ret1, err := rt.HasItem(key)
	assert.Nil(t, err)
	assert.Equal(t, false, ret1)
	// test CreateItem:
	err = rt.CreateItem(key, item)
	assert.Nil(t, err)
	// test CreateItem:
	item2 := TestStruct{}
	err = rt.GetItem(key, &item2)
	assert.Nil(t, err)
	assert.Equal(t, item, item2)
	// test
	item3 := TestStruct{
		A: 1,
		B: "abc",
		C: []byte("cde"),
		D: []string{"f", "g", "high"},
		E: &SubStruct{
			M: 1,
			N: "b",
			O: []byte("aaa"),
			P: []string{"l", "ll", "lll"},
		},
	}
	err = rt.UpdateItem(key, item3)
	assert.Nil(t, err)
	item4 := TestStruct{}
	err = rt.GetItem(key, &item4)
	assert.Nil(t, err)
	assert.Equal(t, item3, item4)
	// test DeleteItem:
	err = rt.DeleteItem(key)
	assert.Nil(t, err)
	ret2, err := rt.HasItem(key)
	assert.Nil(t, err)
	assert.Equal(t, false, ret2)
}
