package bitxid

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
)

// Struct2Bytes .
func Struct2Bytes(s interface{}) ([]byte, error) {
	buf := bytes.Buffer{}
	err := gob.NewEncoder(&buf).Encode(s)
	if err != nil {
		fmt.Println("gob Encode err:", err)
		return []byte{}, err
	}
	return buf.Bytes(), nil
}

// Bytes2Struct .
func Bytes2Struct(b []byte, s interface{}) error {
	buf := bytes.NewBuffer(b)
	err := gob.NewDecoder(buf).Decode(s)
	if err != nil {
		fmt.Println("gob Decode:", err)
		return err
	}
	return nil
}

// UnmarshalDIDDoc converts byte doc to struct doc
func UnmarshalDIDDoc(docBytes []byte) (DIDDoc, error) {
	docStruct := DIDDoc{}
	err := Bytes2Struct(docBytes, &docStruct)
	if err != nil {
		return DIDDoc{}, err
	}
	return docStruct, nil
}

// MarshalDIDDoc converts struct doc to byte doc
func MarshalDIDDoc(docStruct DIDDoc) ([]byte, error) {
	docBytes, err := Struct2Bytes(docStruct)
	if err != nil {
		return nil, err
	}
	return docBytes, nil
}

// UnmarshalMethodDoc converts byte doc to struct doc
func UnmarshalMethodDoc(docBytes []byte) (MethodDoc, error) {
	docStruct := MethodDoc{}
	err := Bytes2Struct(docBytes, &docStruct)
	if err != nil {
		return MethodDoc{}, err
	}
	return docStruct, nil
}

// MarshalMethodDoc converts struct doc to byte doc
func MarshalMethodDoc(docStruct MethodDoc) ([]byte, error) {
	docBytes, err := Struct2Bytes(docStruct)
	if err != nil {
		return nil, err
	}
	return docBytes, nil
}

//

// Bytesbuf2Struct not good for struct contains string or slice
func Bytesbuf2Struct(buf *bytes.Buffer, s interface{}) error {
	err := binary.Read(buf, binary.BigEndian, s)
	if err != nil {
		fmt.Println("binary.Read:", err)
		return err
	}
	return nil
}

// Struct2Bytesbuf not good for struct contains string or slice
func Struct2Bytesbuf(s interface{}, buf *bytes.Buffer) error {
	err := binary.Write(buf, binary.BigEndian, s)
	if err != nil {
		fmt.Println("binary.Write:", err)
		return err
	}
	return nil
}
