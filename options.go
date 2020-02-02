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

import (
	"crypto/tls"
	"time"
)

// These are options.  We share these between the rpcServer and the rpcClient.

// OptionDialAsync indicates that we wish to dial asynchronously if it is true.
type OptionDialAsync bool

// OptionReconnectTime is a reconnection time -- used for dialers that
// redial when a connection is lost.
type OptionReconnectTime time.Duration

// OptionMaxReconnectTime is the maximum reconnection time.
type OptionMaxReconnectTime time.Duration

// OptionTLSConfig encapsulates a TLS configuration
type OptionTLSConfig *tls.Config

// OptionRetryTime is the time between retries.  Zero means never retry.
type OptionRetryTime time.Duration

// OptionOther lets us encapsulate any other option.
type OptionOther struct {
	Name  string
	Value interface{}
}
