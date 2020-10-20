package types

import (
	"testing"

	"gotest.tools/assert"
)

const did string = "did:bitxhub:appchain001:0x12345678"

func TestGetRootMethod(t *testing.T) {
	method := DID(did).GetRootMethod()
	assert.Equal(t, "bitxhub", method)
}

func TestGetAddress(t *testing.T) {
	addr := DID(did).GetAddress()
	assert.Equal(t, "0x12345678", addr)
}

func TestGetSubMethod(t *testing.T) {
	method := DID(did).GetSubMethod()
	assert.Equal(t, "appchain001", method)
}
