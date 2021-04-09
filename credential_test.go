package bitxid

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/meshplus/bitxhub-kit/storage/leveldb"
	"github.com/stretchr/testify/assert"
)

var testCT = ClaimTyp{
	ID: "Asset001",
	Content: []*FieldTyp{
		{
			Field: "from",
			Typ:   "string",
		},
		{
			Field: "to",
			Typ:   "string",
		},
		{
			Field: "amount",
			Typ:   "uint64",
		},
		{
			Field: "asset",
			Typ:   "string",
		},
	},
}

var testVC = Credential{
	ID:         "VC001",
	Typ:        "Asset001",
	Issuer:     DID("did:bitxhub:appchain001:0x12345678"),
	Issued:     uint64(time.Now().Second()),
	Expiration: uint64(time.Now().Second()) + 864000,
	Claim:      `{"from":"alice","to":"bob","amout":"20.21","asset":"coin"}`,
	Signature: Sig{
		Typ: "RsaSignature2018",
		Content: `eyJhbGciOiJQUzI1NiIsImI2NCI6ZmFsc2UsImNyaXQiOlsiYjY0Il19
		..DJBMvvFAIC00nSGB6Tn0XKbbF9XrsaJZREWvR2aONYTQQxnyXirtXnlewJMB
		Bn2h9hfcGZrvnC1b6PgWmukzFJ1IiH1dWgnDIS81BH-IxXnPkbuYDeySorc4
		QU9MJxdVkY5EL4HYbcIfwKj6X4LBQ2_ZHZIu1jdqLcRZqHcsDF5KKylKc1TH
		n5VRWy5WhYg_gBnyWny8E6Qkrze53MR7OuAmmNJ1m1nN8SxDrG6a08L78J0-
		Fbas5OjAQz3c17GY8mVuDPOBIOVjMEghBlgl3nOi1ysxbRGhHLEK4s0KKbeR
		ogZdgt1DkQxDFxxn41QWDw_mmMCjs9qxg0zcZzqEJw`,
	},
}

func TestVC(t *testing.T) {
	dir, err := ioutil.TempDir("", "vc.store")
	assert.Nil(t, err)
	s, err := leveldb.New(dir)
	assert.Nil(t, err)
	vcr, err := NewVCRegistry(s)
	assert.Nil(t, err)

	testCreateClaimTyp(t, vcr)
	testGetClaimTyp(t, vcr)
	testGetAllClaimTyps(t, vcr)
	testDeleteClaimtyp(t, vcr)

	testStoreVC(t, vcr)
	testGetVC(t, vcr)
	testDeleteVC(t, vcr)
}

func testCreateClaimTyp(t *testing.T, vcr *VCRegistry) {
	id, err := vcr.CreateClaimTyp(testCT)
	assert.Nil(t, err)
	assert.Equal(t, id, testCT.ID)
}

func testGetClaimTyp(t *testing.T, vcr *VCRegistry) {
	ct, err := vcr.GetClaimTyp(testCT.ID)
	assert.Nil(t, err)
	assert.Equal(t, ct, &testCT)
}

func testGetAllClaimTyps(t *testing.T, vcr *VCRegistry) {
	ctlist, err := vcr.GetAllClaimTyps()
	assert.Nil(t, err)
	assert.NotEqual(t, len(ctlist), 0)
}

func testDeleteClaimtyp(t *testing.T, vcr *VCRegistry) {
	vcr.DeleteClaimtyp(testCT.ID)
	ct, err := vcr.GetClaimTyp(testCT.ID)
	assert.Nil(t, err)
	assert.Equal(t, ct, (*ClaimTyp)(nil)) // delete successfully
}

func testStoreVC(t *testing.T, vcr *VCRegistry) {
	id, err := vcr.StoreVC(testVC)
	assert.Nil(t, err)
	assert.Equal(t, id, testVC.ID)
}

func testGetVC(t *testing.T, vcr *VCRegistry) {
	vc, err := vcr.GetVC(testVC.ID)
	assert.Nil(t, err)
	assert.Equal(t, vc, &testVC)
}

func testDeleteVC(t *testing.T, vcr *VCRegistry) {
	vcr.DeleteVC(testVC.ID)
	vc, err := vcr.GetVC(testVC.ID)
	assert.Nil(t, err)
	assert.Equal(t, vc, (*Credential)(nil))
}
