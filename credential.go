package bitxid

import (
	"fmt"

	"github.com/meshplus/bitxhub-kit/storage"
)

// ClaimTyp represents claim type
type ClaimTyp struct {
	ID      string     `json:"id"` // the universal id of claim type
	Content []FieldTyp `json:"content"`
}

// FieldTyp represents field type
type FieldTyp struct {
	Field string `json:"field"` // field name
	Typ   string `json:"typ"`   // field type e.g. int, float, string
}

// Marshal marshals claimTyp
func (c *ClaimTyp) Marshal() ([]byte, error) {
	return Marshal(c)
}

// Unmarshal unmarshals claimTyp
func (c *ClaimTyp) Unmarshal(docBytes []byte) error {
	err := Unmarshal(docBytes, &c)
	return err
}

// Credential represents verifiable credential
type Credential struct {
	ID         string `json:"id"`
	Typ        string `json:"typ"`
	Issuer     string `json:"issuer"`
	Issued     string `json:"issued"`
	Expiration string `json:"expiration"`
	Claim      string `json:"claim"` // jsonSchema string
	Signature  Sig    `json:"signature"`
}

// Sig represents signature data
type Sig struct {
	Typ     string `json:"typ"`
	Content string `json:"content"`
}

// Marshal marshals credential
func (c *Credential) Marshal() ([]byte, error) {
	return Marshal(c)
}

// Unmarshal unmarshals credential
func (c *Credential) Unmarshal(docBytes []byte) error {
	err := Unmarshal(docBytes, &c)
	return err
}

var _ VCManager = (*VCRegistry)(nil)

// VCRegistry represents verifiable credential management registry
type VCRegistry struct {
	Store  storage.Storage `json:"store"`
	CTlist []string        `json:"ct_list"`
}

// NewVCRegistry news a NewVCRegistry
func NewVCRegistry(s storage.Storage) (*VCRegistry, error) {
	return &VCRegistry{
		Store: s,
	}, nil
}

// CreateClaimTyp creates new claim type
func (vcr *VCRegistry) CreateClaimTyp(ct ClaimTyp) (string, error) {
	ctb, err := ct.Marshal()
	if err != nil {
		return "", fmt.Errorf("claim type marshal: %w", err)
	}
	vcr.Store.Put(claimKey(ct.ID), ctb)
	vcr.CTlist = append(vcr.CTlist, ct.ID)
	return ct.ID, nil
}

// GetClaimTyp gets a claim type
func (vcr *VCRegistry) GetClaimTyp(ctid string) (*ClaimTyp, error) {
	ctb := vcr.Store.Get(claimKey(ctid))
	c := &ClaimTyp{}
	err := c.Unmarshal(ctb)
	if err != nil {
		return nil, fmt.Errorf("claim type marshal: %w", err)
	}
	return c, nil
}

// DeleteClaimtyp deletes a claim type
func (vcr *VCRegistry) DeleteClaimtyp(ctid string) {
	vcr.Store.Delete(claimKey(ctid))
	for i, ct := range vcr.CTlist {
		if ct == ctid {
			vcr.CTlist = append(vcr.CTlist[:i], vcr.CTlist[i+1:]...)
		}
	}
}

// GetAllClaimTyps gets all claim types
func (vcr *VCRegistry) GetAllClaimTyps() ([]*ClaimTyp, error) {
	clist := []*ClaimTyp{}
	for _, ctid := range vcr.CTlist {
		ct, err := vcr.GetClaimTyp(ctid)
		if err != nil {
			return nil, fmt.Errorf("get claim type: %w", err)
		}
		clist = append(clist, ct)
	}
	return clist, nil
}

// StoreVC stores a vc
func (vcr *VCRegistry) StoreVC(c Credential) (string, error) {
	cb, err := c.Marshal()
	if err != nil {
		return "", fmt.Errorf("vc marshal: %w", err)
	}
	vcr.Store.Put(vcKey(c.ID), cb)
	return c.ID, nil
}

// GetVC gets a vc
func (vcr *VCRegistry) GetVC(cid string) (*Credential, error) {
	cb := vcr.Store.Get(vcKey(cid))
	c := &Credential{}
	err := c.Unmarshal(cb)
	if err != nil {
		return nil, fmt.Errorf("vc marshal: %w", err)
	}
	return c, nil
}

// DeleteVC deletes a vc
func (vcr *VCRegistry) DeleteVC(cid string) {
	vcr.Store.Delete(vcKey(cid))
}

func claimKey(id string) []byte {
	return []byte("claim-" + string(id))
}

func vcKey(id string) []byte {
	return []byte("vc-" + string(id))
}
