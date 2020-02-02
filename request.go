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

// The definitions in this file relate to request/response (i.e. RPC.)

// Request represents a request object over the wire.
type Request struct {
	// Version is the protocol version, and must be 1.
	Version byte `msgpack:"version"`

	// Method name.
	Method string `msgpack:"method"`

	// Parameters for the request (optional).
	Parameters interface{} `msgpack:"params,omitempty"`
}

// Response represents a response to a request.
type Response struct {
	// Version is the protocol version, and must be 1.
	Version byte `msgpack:"version"`

	// Success is true on success, false on failure.
	Success bool `msgpack:"success"`

	// Result is optional result data.  Will not be present
	// if Success is false.
	Result interface{} `msgpack:"result,omitempty"`

	// Error object must be present if Success is false, and
	// must NOT be present if Success is true.
	Error *Error `msgpack:"err,omitempty"`
}
