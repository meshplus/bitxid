package bitxid

import (
	"fmt"
	"strings"
)

// DID represents decentrilzed identifier and method names
// example identifier: did:bitxhub:appchain001
// example method name:
type DID string

// StatusType .
type StatusType int

// the rule of status code:
// end with 1 (001, 101, 301, etc.) means on audit
// end with 5 (005, 105, 205, 305, etc.) means audit failed
// end with 0 (010, 110, 200, 310, etc.) means good
// 101/105/110 301/305/310 not used currently
const (
	Error           StatusType = -001
	Initial         StatusType = 000
	ApplyAudit      StatusType = 001
	ApplyFailed     StatusType = 005
	ApplySuccess    StatusType = 010
	RegisterAudit   StatusType = 101
	RegisterFailed  StatusType = 105
	RegisterSuccess StatusType = 110
	Normal          StatusType = 200
	Frozen          StatusType = 205
	UpdateAudit     StatusType = 301
	UpdateFailed    StatusType = 305
	UpdateSuccess   StatusType = 310
)

// DocType .
type DocType int

// type of doc
const (
	MethodDocType DocType = iota
	DIDDocType
)

// TableType .
type TableType int

// type of doc
const (
	MethodTableType TableType = iota
	DIDTableType
)

// BasicDoc is the fundamental part of doc structure
type BasicDoc struct {
	ID             DID      `json:"id"`
	Type           string   `json:"type"`
	Created        uint64   `json:"created"`
	Updated        uint64   `json:"updated"`
	Controller     DID      `json:"controller"`
	PublicKey      []PubKey `json:"publicKey"`
	Authentication []Auth   `json:"authentication"`
}

// BasicItem is the fundamental part of item structure
type BasicItem struct {
	ID      DID
	DocAddr string     // addr where the doc file stored
	DocHash []byte     // hash of the doc file
	Status  StatusType // status of the item
}

// PubKey .
type PubKey struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	PublicKeyPem string `json:"publicKeyPem"`
}

// Auth .
type Auth struct {
	PublicKey []string `json:"publicKey"` // ID of PublicKey
}

// IsValidFormat .
func (did DID) IsValidFormat() bool {
	s := strings.Split(string(did), ":")
	if len(s) != 4 || s[0] != "did" || s[1] == "" || s[2] == "" || s[3] == "" {
		return false
	}
	return true
}

// GetRootMethod get root method from did-format string
func (did DID) GetRootMethod() string {
	if !did.IsValidFormat() {
		return ""
	}
	s := strings.Split(string(did), ":")
	return s[1]
}

// GetSubMethod get sub method from did-format string
func (did DID) GetSubMethod() string {
	if !did.IsValidFormat() {
		return ""
	}
	s := strings.Split(string(did), ":")
	return s[2]
}

// GetAddress get address from did-format string
func (did DID) GetAddress() string {
	if !did.IsValidFormat() {
		return ""
	}
	s := strings.Split(string(did), ":")
	return s[3]
}

// GetMethod .
func (did DID) GetMethod() string {
	return "did:" + did.GetRootMethod() + ":" + did.GetSubMethod() + ":."
}

func errJoin(module string, err error) error {
	return fmt.Errorf("%s: %v", module, err)
}
