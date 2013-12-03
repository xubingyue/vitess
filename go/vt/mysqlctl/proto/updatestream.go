// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package proto

import (
	"github.com/youtube/vitess/go/rpcwrap"
	"github.com/youtube/vitess/go/vt/key"
)

// UpdateStreamRequest is used to make a request for ServeUpdateStream.
type UpdateStreamRequest struct {
	GroupId string
}

// KeyrangeRequest is used to make a request for StreamKeyrange.
type KeyrangeRequest struct {
	GroupId  string
	Keyrange key.KeyRange
}

// UpdateStream defines the rpc API for the update stream service.
type UpdateStream interface {
	ServeUpdateStream(req *UpdateStreamRequest, sendReply func(reply interface{}) error) (err error)
	StreamKeyrange(req *KeyrangeRequest, sendReply func(reply interface{}) error) (err error)
}

// RegisterAuthenticated regiesters a varaiable that satisfies the UpdateStream interface
// as an rpc service that requires authentication.
func RegisterAuthenticated(service UpdateStream) {
	rpcwrap.RegisterAuthenticated(service)
}
