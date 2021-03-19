/*
Copyright The casbind Authors.
@Date: 2021/03/10 14:42
*/

package options

import (
	"github.com/WenyXu/casbind/pkg/transport/grpc"
	"github.com/WenyXu/casbind/pkg/transport/http"
)

type Options struct {
	httpConfig http.Config
	grpcConfig grpc.Config
}

type Option func(*Options)

func NewOptionsFormFlags() (*Options, error) {
	return nil, nil
}

func NewOptionsFormEnvs() (*Options, error) {
	return nil, nil
}

func Default(opt ...Option) *Options {
	return new(append([]Option{}, opt...))
}

func new(opt ...Option) *Options {
	opts := &Options{}
	for _, o := range opt {
		o(opts)
	}
	return opts
}
