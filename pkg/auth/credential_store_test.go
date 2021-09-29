// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package auth

import (
	"strings"
	"testing"
)

type testBasicAuther struct {
	ok       bool
	username string
	password string
}

func (t *testBasicAuther) BasicAuth() (string, string) {
	return t.username, t.password
}

func Test_AuthLoadSingle(t *testing.T) {
	const jsonStream = `
			{"username1": "password1"}
	`

	store := NewCredentialsStore()
	if err := store.Load(strings.NewReader(jsonStream)); err != nil {
		t.Fatalf("failed to load single credential: %s", err.Error())
	}

	if check := store.Check("username1", "password1"); !check {
		t.Fatalf("single credential not loaded correctly")
	}
	if check := store.Check("username1", "wrong"); check {
		t.Fatalf("single credential not loaded correctly")
	}

	if check := store.Check("wrong", "password1"); check {
		t.Fatalf("single credential not loaded correctly")
	}
	if check := store.Check("wrong", "wrong"); check {
		t.Fatalf("single credential not loaded correctly")
	}
}

func Test_AuthLoadHashedSingleRequest(t *testing.T) {
	const jsonStream = `
		{
				"username1": "$2a$10$fKRHxrEuyDTP6tXIiDycr.nyC8Q7UMIfc31YMyXHDLgRDyhLK3VFS",
				"username2": "password2"
		}
	`

	store := NewCredentialsStore()
	if err := store.Load(strings.NewReader(jsonStream)); err != nil {
		t.Fatalf("failed to load multiple credentials: %s", err.Error())
	}

	b1 := &testBasicAuther{
		username: "username1",
		password: "password1",
	}
	b2 := &testBasicAuther{
		username: "username2",
		password: "password2",
	}

	b3 := &testBasicAuther{
		username: "username1",
		password: "wrong",
	}
	b4 := &testBasicAuther{
		username: "username2",
		password: "wrong",
	}

	if check := store.CheckRequest(b1); !check {
		t.Fatalf("username1 (b1) credential not checked correctly via request")
	}
	if check := store.CheckRequest(b2); !check {
		t.Fatalf("username2 (b2) credential not checked correctly via request")
	}
	if check := store.CheckRequest(b3); check {
		t.Fatalf("username1 (b3) credential not checked correctly via request")
	}
	if check := store.CheckRequest(b4); check {
		t.Fatalf("username2 (b4) credential not checked correctly via request")
	}
}
