package bitxid

import (
	"fmt"
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

func TestErrJoin(t *testing.T) {
	err := errJoin("doc db store", fmt.Errorf("a error"))
	assert.NotNil(t, err)
}

func TestIsValidFormat(t *testing.T) {
	did := DID("did:bitxid:appchain001:0x123")
	res := did.IsValidFormat()
	assert.Equal(t, true, res)

	did = DID("did:::")
	res = did.IsValidFormat()
	assert.Equal(t, false, res)

	method := DID("did:bitxid:appchain001:.")
	res = method.IsValidFormat()
	assert.Equal(t, true, res)

	method = DID("did:bitxid:")
	res = method.IsValidFormat()
	assert.Equal(t, false, res)
}
