package bitxid

import (
	"fmt"
	"strings"
)

// RegistryMode .
type RegistryMode int

// type of RegistryMode:
// @ExternalDocDB: Doc store won't be mastered by Registry
// @InternalDocDB: Doc store will be mastered by Registry
const (
	ExternalDocDB RegistryMode = iota
	InternalDocDB
)

// DID represents decentrilzed identifier and method names
// example identifier: did:bitxhub:appchain001
// example method name:
type DID string

// StatusType .
type StatusType string

// the rule of status code:
// @BadStatus: something went wrong during get status
// @Normal: AuditSuccess or Unfrozen
const (
	BadStatus      StatusType = "BadStatus"
	Initial        StatusType = "Initial"
	ApplyAudit     StatusType = "ApplyAudit"
	ApplyFailed    StatusType = "ApplyFailed"
	ApplySuccess   StatusType = "ApplySuccess"
	RegisterAudit  StatusType = "RegisterAudit"
	RegisterFailed StatusType = "RegisterFailed"
	UpdateAudit    StatusType = "UpdateAudit"
	UpdateFailed   StatusType = "UpdateFailed"
	Frozen         StatusType = "Frozen"
	Normal         StatusType = "Normal"
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

// KeyType .
type KeyType int

// value of keytype
const (
	AES KeyType = iota
	ThirdDES
	RSA
	Secp256k1
	ECDSAP256
	ECDSAP384
	ECDSAP521
	Ed25519
)

// DocOption .
// Content should be nil if Registry.mode == ExternalDocDB
type DocOption struct {
	ID      DID
	Addr    string
	Hash    []byte
	Content Doc
}

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
