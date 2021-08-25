/*
@Author: Weny Xu
@Date: 2021/08/25 21:51
*/

package auth

import (
	"encoding/json"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"io"
)

// BasicAuther is the interface an object must support to return basic auth information.
type BasicAuther interface {
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
func (c *CredentialsStore) CheckRequest(b BasicAuther) bool {
	username, password := b.BasicAuth()
	if !c.Check(username, password) {
		return false
	}
	return true
}
