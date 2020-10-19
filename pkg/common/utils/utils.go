package utils

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
