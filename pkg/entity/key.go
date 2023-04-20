package entity

import (
	"errors"

	"google.golang.org/protobuf/proto"
)

type Key struct {
	Id         string
	Type       string
	Raw        interface{}
	EncodeFunc KeyEncodeFunc
}

type KeyEncodeFunc func(interface{}) (proto.Message, error)

// Encode is a helper method used to invoke EncodeFunc safely.
func (key Key) Encode() (proto.Message, error) {
	if key.EncodeFunc == nil {
		// Return generic zero value and an error.
		return nil, errors.New("encodeFunc is nil")
	}

	return key.EncodeFunc(key.Raw)
}
