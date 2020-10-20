package types

import "strings"

// DID .
type DID string

// BasicDoc is the fundamental part of doc structure
type BasicDoc struct {
	ID             string   `json:"id"`
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
	if len(s) != 4 && s[0] != "did" && (s[1] == "" || s[2] == "" || s[3] == "") {
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
