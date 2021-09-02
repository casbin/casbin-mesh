/*
@Author: Weny Xu
@Date: 2021/09/02 16:47
*/

package auth

import "errors"

const (
	Basic    = "Basic"
	AuthType = "AuthType"
)

var (
	ERRUNAUTHORIZED      = errors.New("UNAUTHORIZED")
	ERRUNSUPPORTAUTHTYPE = errors.New("UNSUPPORT AUTH TYPE")
)
