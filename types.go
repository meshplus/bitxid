package bitxid

import (
	"fmt"
	"strings"
)

// DID represents decentrilzed identifier and method names
// example identifier: did:bitxhub:appchain001
// example method name:
type DID string

// the rule of status code:
// end with 1 (001, 101, 301, etc.) means on audit
// end with 5 (005, 105, 205, 305, etc.) means audit failed
// end with 0 (010, 110, 200, 310, etc.) means good
// 101/105/110 301/305/310 not used currently
const (
	Error           int = -001
	Initial         int = 000
	ApplyAudit      int = 001
	ApplyFailed     int = 005
	ApplySuccess    int = 010
	RegisterAudit   int = 101
	RegisterFailed  int = 105
	RegisterSuccess int = 110
	Normal          int = 200
	Frozen          int = 205
	UpdateAudit     int = 301
	UpdateFailed    int = 305
	UpdateSuccess   int = 310
)

// type of doc
const (
	MethodDocType int = iota
	DIDDocType
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
