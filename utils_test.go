package bitxid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type s struct {
	A string
	B int
}

var testBytes []byte

func TestStruct2Bytes(t *testing.T) {
	testStruct := s{
		A: "aaa",
		B: 1,
	}
	// expectedByte := []byte{27, 255, 153, 3, 1, 1, 1, 83, 1, 255, 154, 0, 1, 2, 1, 1, 65, 1, 12, 0, 1, 1, 66, 1, 4, 0, 0, 0, 10, 255, 154, 1, 3, 97, 97, 97, 1, 2, 0}
	expectedByte := []byte{27, 255, 153, 3, 1, 1, 1, 115, 1, 255, 154, 0, 1, 2, 1, 1, 65, 1, 12, 0, 1, 1, 66, 1, 4, 0, 0, 0, 10, 255, 154, 1, 3, 97, 97, 97, 1, 2, 0}
	testBytes1, err := Struct2Bytes(testStruct)
	assert.Nil(t, err)
	assert.Equal(t, expectedByte, testBytes1)
	testBytes = testBytes1
}

func TestBytes2Struct(t *testing.T) {
	expectedStruct := s{
		A: "aaa",
		B: 1,
	}
	var testStruct s
	err := Bytes2Struct(testBytes, &testStruct)
	assert.Nil(t, err)
	assert.Equal(t, expectedStruct, testStruct)
}
