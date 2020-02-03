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

import "github.com/vmihailenco/msgpack"

// Nil allows a pointer to encode as Nil. This will be handled
// arguments or parameters by encoding as the nil value, and
// likewise decoding as nil.  This is needed because Go doesn't
// have a void type.  To use this, just declare either parameters
// or returns as *Nil.
type Nil struct {}

func (*Nil) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeNil()
}

func (*Nil) DecodeMsgpack(dec *msgpack.Decoder) error {
	return dec.DecodeNil()
}