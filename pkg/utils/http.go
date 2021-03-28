/*
Copyright 2020 RS4
@Author: Weny Xu
@Date: 2021/03/28 20:19
*/

package utils

import (
	"fmt"
	"strings"
)

// NormalizeAddr ensures that the given URL has a HTTP protocol prefix.
// If none is supplied, it prefixes the URL with "http://".
func NormalizeAddr(addr string) string {
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		return fmt.Sprintf("http://%s", addr)
	}
	return addr
}

// EnsureHTTPS modifies the given URL, ensuring it is using the HTTPS protocol.
func EnsureHTTPS(addr string) string {
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		return fmt.Sprintf("https://%s", addr)
	}
	return strings.Replace(addr, "http://", "https://", 1)
}
