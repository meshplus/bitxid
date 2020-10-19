package types

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
