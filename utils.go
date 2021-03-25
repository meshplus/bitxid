package bitxid

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

// Marshal .
func Marshal(s interface{}) ([]byte, error) {
	buf := bytes.Buffer{}
	err := gob.NewEncoder(&buf).Encode(s)
	if err != nil {
		return []byte{}, fmt.Errorf("gob encode err: %w", err)
	}
	return buf.Bytes(), nil
}

// Unmarshal .
func Unmarshal(b []byte, s interface{}) error {
	buf := bytes.NewBuffer(b)
	err := gob.NewDecoder(buf).Decode(s)
	if err != nil {
		return fmt.Errorf("gob decode err: %w", err)
	}
	return nil
}

// UnmarshalDIDDoc converts byte doc to struct doc
func UnmarshalDIDDoc(docBytes []byte) (DIDDoc, error) {
	docStruct := DIDDoc{}
	err := Unmarshal(docBytes, &docStruct)
	if err != nil {
		return DIDDoc{}, err
	}
	return docStruct, nil
}

// MarshalDIDDoc converts struct doc to byte doc
func MarshalDIDDoc(docStruct DIDDoc) ([]byte, error) {
	docBytes, err := Marshal(docStruct)
	if err != nil {
		return nil, err
	}
	return docBytes, nil
}

// UnmarshalChainDoc converts byte doc to struct doc
func UnmarshalChainDoc(docBytes []byte) (ChainDoc, error) {
	docStruct := ChainDoc{}
	err := Unmarshal(docBytes, &docStruct)
	if err != nil {
		return ChainDoc{}, err
	}
	return docStruct, nil
}

// MarshalChainDoc converts struct doc to byte doc
func MarshalChainDoc(docStruct ChainDoc) ([]byte, error) {
	docBytes, err := Marshal(docStruct)
	if err != nil {
		return nil, err
	}
	return docBytes, nil
}
