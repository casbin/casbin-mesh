/*
@Author: Weny Xu
@Date: 2021/09/02 16:47
*/

package auth

import "errors"

type AuthType string

var (
	Basic = AuthType("Basic")
	Noop  = AuthType("Noop")
)

var (
	ErrUnauthorized        = errors.New("unauthorized")
	ErrUnsupportedAuthType = errors.New("unsupported auth type")
)
