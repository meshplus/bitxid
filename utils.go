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

// UnmarshalAccountDoc converts byte doc to struct doc
func UnmarshalAccountDoc(docBytes []byte) (AccountDoc, error) {
	docStruct := AccountDoc{}
	err := Unmarshal(docBytes, &docStruct)
	if err != nil {
		return AccountDoc{}, err
	}
	return docStruct, nil
}

// MarshalAccountDoc converts struct doc to byte doc
func MarshalAccountDoc(docStruct AccountDoc) ([]byte, error) {
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
