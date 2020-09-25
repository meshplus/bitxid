package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type S struct {
	A string
	B int
}

// TestBytes2Struct
var TestByte []byte

func TestStruct2Bytes(t *testing.T) {
	testStruct := S{
		A: "aaa",
		B: 1,
	}
	expectedByte := []byte{27, 255, 129, 3, 1, 1, 1, 83, 1, 255, 130, 0, 1, 2, 1, 1, 65, 1, 12, 0, 1, 1, 66, 1, 4, 0, 0, 0, 10, 255, 130, 1, 3, 97, 97, 97, 1, 2, 0}
	testByte, err := Struct2Bytes(testStruct)
	assert.Nil(t, err)
	assert.Equal(t, expectedByte, testByte)
	TestByte = testByte
}

func TestBytes2Struct(t *testing.T) {
	expectedStruct := S{
		A: "aaa",
		B: 1,
	}
	var testStruct S
	err := Bytes2Struct(TestByte, &testStruct)
	assert.Nil(t, err)
	assert.Equal(t, expectedStruct, testStruct)
}
