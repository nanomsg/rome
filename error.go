// Copyright 2020 Staysail Systems, Inc. <info@staysail.tech>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use file except in compliance with the License.
// You may obtain a copy of the license at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rome

import "errors"

// Error is a protocol error exposed in response to a request.
type Error struct {
	Code    int         `msgpack:"code"`
	Message string      `msgpack:"message"`
	Data    interface{} `msgpack:"data,omitempty"`
	base    error
}

// Error codes.  The first set are defined by JSON-RPC, and we use the
// same values.  We also insert our own specific values in the range
// allocated for that (starting at -32000).
const (
	ErrParse          = -32700
	ErrInvalidRequest = -32600
	ErrMethodNotFound = -32601
	ErrInvalidParams  = -32602
	ErrInternal       = -32603
	ErrUnspecified    = -32000
	ErrBadVersion     = -32001
)

// NewError is used to generate a new error object.  When rpcServer methods
// desire to return an error object, this is preferred to give the most
// specific error.
func NewError(code int, message string, data interface{}) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Data:    data,
		base:    nil,
	}
}

func (z *Error) Error() string {
	return z.Message
}

func (z *Error) Unwrap() error {
	return z.base
}

func ErrorWrap(e error) *Error {
	var z *Error
	if errors.As(e, &z) {
		return z
	}
	return &Error{
		Code:    ErrUnspecified,
		Message: e.Error(),
		Data:    e,
		base:    e,
	}
}
