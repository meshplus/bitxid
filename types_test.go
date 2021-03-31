package bitxid

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testDid = DID("did:bitxhub:test-method:test-address")

func TestGetRootMethod(t *testing.T) {
	method := testDid.GetRootMethod()
	assert.Equal(t, "bitxhub", method)
}

func TestGetSubMethod(t *testing.T) {
	method := testDid.GetSubMethod()
	assert.Equal(t, "test-method", method)
}

func TestGetAddress(t *testing.T) {
	addr := testDid.GetAddress()
	assert.Equal(t, "test-address", addr)
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
