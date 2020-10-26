package bitxid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRootMethod(t *testing.T) {
	method := DID(did).GetRootMethod()
	assert.Equal(t, "bitxhub", method)
}

func TestGetSubMethod(t *testing.T) {
	method := DID(did).GetSubMethod()
	assert.Equal(t, "appchain001", method)
}

func TestGetAddress(t *testing.T) {
	addr := DID(did).GetAddress()
	assert.Equal(t, "0x12345678", addr)
}
