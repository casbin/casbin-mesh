/*
@Author: Weny Xu
@Date: 2021/09/02 16:05
*/

package auth

type RegistryOptions struct {
	grpc map[string]GrpcMiddleware
	http map[string]HttpMiddleware
}

func (r RegistryOptions) GetGrpcMiddleware(typ string) (GrpcMiddleware, bool) {
	if v, ok := r.grpc[typ]; ok {
		return v, ok
	}
	return nil, false
}

func (r RegistryOptions) GetHttpMiddleware(typ string) (HttpMiddleware, bool) {
	if v, ok := r.http[typ]; ok {
		return v, ok
	}
	return nil, false
}

func NewRegistryOptions() RegistryOptions {
	return RegistryOptions{}
}

func (r *RegistryOptions) RegisterHttpMiddleware(typ string, a HttpMiddleware) *RegistryOptions {
	r.http[typ] = a
	return r
}

func (r *RegistryOptions) RegisterGrpcMiddleware(typ string, a GrpcMiddleware) *RegistryOptions {
	r.grpc[typ] = a
	return r
}

func (r *RegistryOptions) Compile() AuthorMiddleware {
	return &authorMiddleware{*r}
}
