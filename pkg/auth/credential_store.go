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
	"encoding/json"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"io"
)

// BasicAuthProvider is the interface an object must support to return basic auth information.
type BasicAuthProvider interface {
	BasicAuth() (string, string)
}

// Credential represents authentication and authorization configuration for a single user.
type Credential struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// CredentialsStore stores authentication and authorization information for all users.
type CredentialsStore struct {
	store map[string]string
}

// NewCredentialsStore returns a new instance of a CredentialStore.
func NewCredentialsStore() *CredentialsStore {
	return &CredentialsStore{
		store: make(map[string]string),
	}
}

var (
	ErrUserExists    = errors.New("user exists")
	ErrUserNotExists = errors.New("user not exists")
)

// Add adds a Account
func (c *CredentialsStore) Add(username string, password string) error {
	if _, ok := c.store[username]; ok {
		return ErrUserExists
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	c.store[username] = string(hashed)
	return nil
}

// Remove removes a Account
func (c *CredentialsStore) Remove(username string) error {
	if _, ok := c.store[username]; ok {
		delete(c.store, username)
		return nil
	}
	return ErrUserNotExists
}

// Update updates a Account
func (c *CredentialsStore) Update(username, newPassword string) error {
	if _, ok := c.store[username]; ok {
		c.store[username] = newPassword
		return nil
	}
	return ErrUserNotExists
}

// Load loads credential information from a reader.
func (c *CredentialsStore) Load(r io.Reader) error {
	dec := json.NewDecoder(r)
	err := dec.Decode(&c.store)
	if err != nil {
		return err
	}
	return nil
}

// Snapshot takes a snapshot
func (c *CredentialsStore) Snapshot(w io.Writer) error {
	data, err := json.Marshal(c.store)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// Check returns true if the password is correct for the given username.
func (c *CredentialsStore) Check(username, password string) bool {
	pw, ok := c.store[username]
	if !ok {
		return false
	}
	return password == pw ||
		bcrypt.CompareHashAndPassword([]byte(pw), []byte(password)) == nil
}

// CheckRequest returns true if b contains a valid username and password.
func (c *CredentialsStore) CheckRequest(b BasicAuthProvider) bool {
	username, password := b.BasicAuth()
	if !c.Check(username, password) {
		return false
	}
	return true
}
