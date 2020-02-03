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
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type Accumulator struct {
	num  int
	lock sync.Mutex
}

func MustPass(t *testing.T, e error, what string) {
	if e != nil {
		t.Errorf("%s: did not pass: %v", what, e)
	}
}

func MustFail(t *testing.T, e error, expect error, what string) {
	if e == nil {
		t.Errorf("%s: should have failed but did not", what)
		return
	}

	if expect == nil {
		return
	}

	if errors.Is(e, expect) {
		return
	}

	t.Errorf("%s: should have failed as %v but was %v", what, expect, e)
}

func (a *Accumulator) Add(x *int, result *int) error {
	if *x >= 1000 {
		return errors.New("addend too large")
	}

	a.lock.Lock()
	a.num += *x
	*result = a.num
	a.lock.Unlock()
	return nil
}

func makePair(t *testing.T, url string, a *Accumulator) (RpcServer, RpcClient) {
	server := NewRpcServer()
	client := NewRpcClient()

	MustPass(t, server.Register(a), "Register Accumulator")
	MustPass(t, server.Listen(url), "server listen")

	server.ServeAsync(3)

	MustPass(t, client.Dial(url), "client dial")
	time.Sleep(time.Millisecond*20) // give time for settling
	return server, client
}

func TestRpcBasic(t *testing.T) {

	a := &Accumulator{}

	server, client := makePair(t, "inproc:///rpc_basic", a)
	defer server.Close()
	defer client.Close()

	var arg, res int

	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	arg = 5
	MustPass(t, client.Call(ctx, "Accumulator.Add", &arg, &res), "call")
	if res != 5 {
		t.Errorf("Wrong result: %v", res)
	}

	ctx, _ = context.WithTimeout(context.Background(), time.Second)
	arg = 57
	MustPass(t, client.Call(ctx, "Accumulator.Add", &arg, &res), "call2")
	if res != 5+57 {
		t.Errorf("Wrong result: %v", res)
	}
}

func TestRpcMethodNotFound(t *testing.T) {

	a := &Accumulator{}

	server, client := makePair(t, "inproc:///rpc_not_found", a)
	defer server.Close()
	defer client.Close()

	var arg, res int

	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	e := client.Call(ctx, "Accumulator.DoesNotExist", &arg, &res)
	MustFail(t, e, nil, "method not existent")
	ne, ok := e.(*Error)
	if !ok {
		t.Errorf("expected our error but got %v", e)
	}
	if ne.Code != ErrMethodNotFound {
		t.Errorf("wrong code: %v != %v", ne.Code, ErrMethodNotFound)
	}
	if ne.Message != "method not found" {
		t.Errorf("wrong message: %v", ne.Message)
	}
}

func TestRpcMethodFails(t *testing.T) {

	a := &Accumulator{}

	server, client := makePair(t, "inproc:///rpc_fails", a)
	defer server.Close()
	defer client.Close()

	var arg, res int

	ctx, _ := context.WithTimeout(context.Background(), time.Second)

	arg = 1e9
	e := client.Call(ctx, "Accumulator.Add", &arg, &res)
	MustFail(t, e, nil, "method not existent")
	ne, ok := e.(*Error)
	if !ok {
		t.Errorf("expected our error but got %v", e)
	}
	if ne.Code != ErrUnspecified {
		t.Errorf("wrong code: %v != %v", ne.Code, ErrUnspecified)
	}
	if ne.Message != "addend too large" {
		t.Errorf("wrong message: %v", ne.Message)
	}
}

func Subtract(args *[]int, res *int) error {
	if len(*args) != 2 {
		return errors.New("bad params")
	}
	*res = (*args)[0] - (*args)[1]
	return nil
}

func TestRpcBareFunc(t *testing.T) {

	a := &Accumulator{}

	server, client := makePair(t, "inproc:///rpc_bare_func", a)
	defer server.Close()
	defer client.Close()

	MustPass(t, server.RegisterFunc("subtract", Subtract), "register bare name")
	var arg, res int

	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	arg = 5
	MustPass(t, client.Call(ctx, "Accumulator.Add", &arg, &res), "call")
	if res != 5 {
		t.Errorf("Wrong result: %v", res)
	}

	ctx, _ = context.WithTimeout(context.Background(), time.Second)
	arg = 57
	MustPass(t, client.Call(ctx, "Accumulator.Add", &arg, &res), "call2")
	if res != 5+57 {
		t.Errorf("Wrong result: %v", res)
	}

	ctx, _ = context.WithTimeout(context.Background(), time.Second)
	MustPass(t, client.Call(ctx, "subtract", []int{5, 3}[:], &res), "call2")
	if res != 2 {
		t.Errorf("Wrong result: %v", res)
	}
}

func TestRpcTime(t *testing.T) {

	a := &Accumulator{}

	server, client := makePair(t, "inproc:///rpc_bare_func", a)
	defer server.Close()
	defer client.Close()

	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	now := time.Now().UnixNano()
	var res int64

	MustPass(t, client.Call(ctx, "_rpc.time", nil, &res), "call")
	if res < now {
		t.Errorf("Wrong result: want %v < %v", now, res)
	}
}

func TestRpcMethods(t *testing.T) {

	a := &Accumulator{}

	server, client := makePair(t, "inproc:///rpc_bare_func", a)
	defer server.Close()
	defer client.Close()

	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	var res []string

	MustPass(t, client.Call(ctx, "_rpc.methods", nil, &res), "call")
	if len(res) < 3 {
		t.Errorf("method list too short")
	}
	for _, m := range res {
		t.Logf("method: %s", m)
	}
}
