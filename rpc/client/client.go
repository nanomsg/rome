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

package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/vmihailenco/msgpack"
	"go.nanomsg.org/mangos/v3"
	"go.nanomsg.org/mangos/v3/protocol/req"
	// import all the transports
	_ "go.nanomsg.org/mangos/v3/transport/all"

	"go.nanomsg.org/rome"
	"go.nanomsg.org/rome/rpc"
)

// This relates to the RPC rpcClient.

type Client interface {
	// Call calls an RPC.  This runs synchronously, and may be canceled
	// via the context.  It is possible for multiple outstanding calls
	// to be posted this way.  The args may be nil, or an array or a
	// a structure.  The results are either nil (for a method that has
	// no return), or a pointer to a structure that can be unmarshaled.
	//
	// The expectation is that type safe wrappers will be provided
	// by service packages.
	Call(ctx context.Context, name string, args interface{}, results interface{}) error

	// Dial is used to dial a remote server.  This may be called multiple
	// times to dial to different servers.  If multiple connections are
	// present, then the rpcClient will automatically select the best
	// one based on readiness to service the request.
	Dial(url string, opts ...interface{}) error

	// Listen is much like dial, but acts as a server.  This allows
	// the normal TCP server/rpcClient roles to be reversed while still
	// maintaining the REQ/REP higher level roles.  It is possible
	// to freely mix and match multiple calls of Listen, with or without
	// calls to Dial.
	Listen(url string, opts ...interface{}) error

	// SetOption sets global options on the rpcClient, such as retry times.
	SetOption(opts ...interface{}) error

	// Close closes down the socket.  In-flight requests will be aborted
	// and return accordingly.
	Close()
}

func NewClient() Client {
	c := &rpcClient{}
	c.socket, _ = req.NewSocket()
	return c
}

type rpcClient struct {
	socket mangos.Socket
}

func (c *rpcClient) Close() {
	_ = c.socket.Close()
}

// Re-export some names from the parent rome package.

type OptionOther = rome.OptionOther
type OptionDialAsync = rome.OptionDialAsync
type OptionReconnectTime = rome.OptionReconnectTime
type OptionMaxReconnectTime = rome.OptionMaxReconnectTime
type OptionTLSConfig = rome.OptionTLSConfig
type OptionRetryTime = rome.OptionRetryTime
type Nil = rome.Nil
type Error = rpc.Error

// NB: We don't use the mangos timeout options.  Instead we rely on the
// context to provide a global timeout which encompasses both the send
// and the receive time.

func (c *rpcClient) SetOption(opts ...interface{}) error {
	for _, o := range opts {
		switch v := o.(type) {
		case OptionOther:
			return c.socket.SetOption(v.Name, v.Value)
		case OptionDialAsync:
			return c.socket.SetOption(mangos.OptionDialAsynch, v)
		case OptionReconnectTime:
			return c.socket.SetOption(mangos.OptionReconnectTime, v)
		case OptionMaxReconnectTime:
			return c.socket.SetOption(mangos.OptionMaxReconnectTime, v)
		case OptionTLSConfig:
			return c.socket.SetOption(mangos.OptionTLSConfig, v)
		case OptionRetryTime:
			return c.socket.SetOption(mangos.OptionRetryTime, v)
		default:
			return errors.New("unknown option")
		}
	}
	return nil
}

func (c *rpcClient) Dial(url string, opts ...interface{}) error {
	d, e := c.socket.NewDialer(url, nil)
	if e != nil {
		return e
	}
	for _, o := range opts {

		switch v := o.(type) {
		case OptionOther:
			e = d.SetOption(v.Name, v.Value)
			if e != nil {
				return e
			}
		case OptionDialAsync:
			e = d.SetOption(mangos.OptionDialAsynch, v)
			if e != nil {
				return e
			}
		case OptionReconnectTime:
			e = d.SetOption(mangos.OptionReconnectTime, v)
			if e != nil {
				return e
			}
		case OptionMaxReconnectTime:
			e = d.SetOption(mangos.OptionMaxReconnectTime, v)
			if e != nil {
				return e
			}
		case OptionTLSConfig:
			e = d.SetOption(mangos.OptionTLSConfig, v)
			if e != nil {
				return e
			}
		default:
			return errors.New("unknown option")
		}
	}

	return d.Dial()
}

func (c *rpcClient) Listen(url string, opts ...interface{}) error {
	l, e := c.socket.NewListener(url, nil)
	if e != nil {
		return e
	}
	for _, o := range opts {
		switch v := o.(type) {
		case OptionTLSConfig:
			e = l.SetOption(mangos.OptionTLSConfig, v)
			if e != nil {
				return e
			}
		case OptionOther:
			e = l.SetOption(v.Name, v.Value)
			if e != nil {
				return e
			}
		default:
			return errors.New("unknown option")
		}
	}
	return l.Listen()
}

func decodeError(dec *msgpack.Decoder, de *Error) error {
	l, e := dec.DecodeArrayLen()
	if e != nil {
		return fmt.Errorf("failed decoding error len: %w", e)
	}
	if l != 3 {
		return fmt.Errorf("error length (%d) wrong", l)
	}
	if de.Code, e = dec.DecodeInt(); e != nil {
		return fmt.Errorf("failed decoding error code: %v", e)
	}
	if de.Message, e = dec.DecodeString(); e != nil {
		return fmt.Errorf("failed decoding error: %w", e)
	}
	if de.Data, e = dec.DecodeInterface(); e != nil {
		return fmt.Errorf("failed decoding error data: %w", e)
	}
	return de
}

func (c *rpcClient) Call(ctx context.Context, name string, args interface{}, result interface{}) error {
	mc, err := c.socket.OpenContext()
	if err != nil {
		return err
	}
	var doneError error
	defer func() {
		_ = mc.Close()
	}()

	doneQ := make(chan struct{})

	go func() {
		defer close(doneQ)

		buf := &bytes.Buffer{}
		enc := msgpack.NewEncoder(buf)

		// Encoding header cannot fail.
		_ = enc.EncodeArrayLen(3)
		_ = enc.EncodeUint8(1) // version
		_ = enc.EncodeString(name)

		if e := enc.Encode(args); e != nil {
			doneError = e
			return
		}

		if e := mc.Send(buf.Bytes()); e != nil {
			doneError = e
			return
		}

		b, e := mc.Recv()
		if e != nil {
			doneError = e
			return
		}
		dec := msgpack.NewDecoder(bytes.NewReader(b))

		if l, e := dec.DecodeArrayLen(); e != nil {
			doneError = fmt.Errorf("failed decoding header: %w", e)
			return
		} else if l != 3 {
			doneError = errors.New("header array length wrong")
			return
		}

		if v, e := dec.DecodeUint8(); e != nil {
			doneError = fmt.Errorf("failed decoding version: %w", e)
			return
		} else if v != 1 {
			doneError = errors.New("bad version")
			return
		}

		if pass, e := dec.DecodeBool(); e != nil {
			doneError = fmt.Errorf("failed to decoding success: %w", e)
			return
		} else if !pass {
			doneError = decodeError(dec, &Error{})
			return
		}

		// If the caller has passed nil, that means either the function
		// returns no results, or the caller does not care.  Either
		// way we discard them.
		if result == nil {
			return
		}

		if e := dec.Decode(result); e != nil {
			doneError = e
			return
		}
	}()

	select {
	case <-ctx.Done():
		_ = mc.Close() // This should cause the other side to wake.
		err = ctx.Err()
	case <-doneQ:
		err = doneError
	}
	<-doneQ

	return err
}
