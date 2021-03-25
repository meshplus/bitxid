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

func TestMarshal(t *testing.T) {
	testStruct := s{
		A: "aaa",
		B: 1,
	}
	expectedByte := []byte{27, 255, 151, 3, 1, 1, 1, 115, 1, 255, 152, 0, 1, 2, 1, 1, 65, 1, 12, 0, 1, 1, 66, 1, 4, 0, 0, 0, 10, 255, 152, 1, 3, 97, 97, 97, 1, 2, 0}
	testBytes1, err := Marshal(testStruct)
	assert.Nil(t, err)
	assert.Equal(t, expectedByte, testBytes1)
	testBytes = testBytes1
}

func TestUnmarshal(t *testing.T) {
	expectedStruct := s{
		A: "aaa",
		B: 1,
	}
	var testStruct s
	err := Unmarshal(testBytes, &testStruct)
	assert.Nil(t, err)
	assert.Equal(t, expectedStruct, testStruct)
}
