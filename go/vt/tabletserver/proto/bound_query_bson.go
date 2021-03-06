// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package proto

// DO NOT EDIT.
// FILE GENERATED BY BSONGEN.

import (
	"bytes"

	"github.com/youtube/vitess/go/bson"
	"github.com/youtube/vitess/go/bytes2"
)

// MarshalBson bson-encodes BoundQuery.
func (boundQuery *BoundQuery) MarshalBson(buf *bytes2.ChunkedWriter, key string) {
	bson.EncodeOptionalPrefix(buf, bson.Object, key)
	lenWriter := bson.NewLenWriter(buf)

	bson.EncodeString(buf, "Sql", boundQuery.Sql)
	// map[string]interface{}
	{
		bson.EncodePrefix(buf, bson.Object, "BindVariables")
		lenWriter := bson.NewLenWriter(buf)
		for _k, _v1 := range boundQuery.BindVariables {
			bson.EncodeInterface(buf, _k, _v1)
		}
		lenWriter.Close()
	}

	lenWriter.Close()
}

// UnmarshalBson bson-decodes into BoundQuery.
func (boundQuery *BoundQuery) UnmarshalBson(buf *bytes.Buffer, kind byte) {
	switch kind {
	case bson.EOO, bson.Object:
		// valid
	case bson.Null:
		return
	default:
		panic(bson.NewBsonError("unexpected kind %v for BoundQuery", kind))
	}
	bson.Next(buf, 4)

	for kind := bson.NextByte(buf); kind != bson.EOO; kind = bson.NextByte(buf) {
		switch bson.ReadCString(buf) {
		case "Sql":
			boundQuery.Sql = bson.DecodeString(buf, kind)
		case "BindVariables":
			// map[string]interface{}
			if kind != bson.Null {
				if kind != bson.Object {
					panic(bson.NewBsonError("unexpected kind %v for boundQuery.BindVariables", kind))
				}
				bson.Next(buf, 4)
				boundQuery.BindVariables = make(map[string]interface{})
				for kind := bson.NextByte(buf); kind != bson.EOO; kind = bson.NextByte(buf) {
					_k := bson.ReadCString(buf)
					var _v1 interface{}
					_v1 = bson.DecodeInterface(buf, kind)
					boundQuery.BindVariables[_k] = _v1
				}
			}
		default:
			bson.Skip(buf, kind)
		}
	}
}
