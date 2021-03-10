/*
Copyright The casbind Authors.
@Date: 2021/03/10 15:24
*/

package service

type service struct {
	//opts *options.Options
}

type Service interface {
}

func New() Service {
	return service{}
}
