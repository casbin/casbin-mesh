// Copyright 2023 The Casbin Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/casbin/casbin-mesh/proto/command"
	"github.com/casbin/casbin-mesh/server/auth"
	grpc2 "github.com/casbin/casbin-mesh/server/handler/grpc"
	_ "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type grpcServer struct {
	Core
	command.UnimplementedCasbinMeshServer
}

var (
	NamespaceExisted = errors.New("namespace already existed")

	// NamespaceNotExist namespace not exist
	NamespaceNotExist = errors.New("namespace not exist")
	// UnmarshalFailed unmarshal failed
	UnmarshalFailed = errors.New("unmarshal failed")
)

func FormatResponse(err error) *command.Response {
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}
	return &command.Response{
		Error: errMsg,
	}
}

func (s grpcServer) Enforce(ctx context.Context, request *command.EnforceRequest) (*command.EnforceResponse, error) {
	params := command.ToInterfaces(request.Payload.B)
	result, err := s.Core.Enforce(ctx, request.GetNamespace(), int32(request.Payload.GetLevel()), request.Payload.GetFreshness(), params...)
	if err != nil {
		return &command.EnforceResponse{
			Ok:    false,
			Error: err.Error(),
		}, nil
	}
	return &command.EnforceResponse{
		Ok: result,
	}, nil
}

func (s grpcServer) Request(ctx context.Context, cmd *command.Command) (*command.Response, error) {
	var err error
	switch cmd.Type {
	case command.Type_COMMAND_TYPE_CREATE_NAMESPACE:
		err = s.Core.CreateNamespace(ctx, cmd.GetNamespace())
		return FormatResponse(err), nil

	case command.Type_COMMAND_TYPE_SET_MODEL:
		var p command.SetModelFromString
		if err = proto.Unmarshal(cmd.Payload, &p); err != nil {
			return FormatResponse(UnmarshalFailed), nil
		}
		err = s.Core.SetModelFromString(ctx, cmd.GetNamespace(), p.GetText())
		return FormatResponse(err), nil

	case command.Type_COMMAND_TYPE_ADD_POLICIES:
		var p command.AddPoliciesPayload
		if err = proto.Unmarshal(cmd.Payload, &p); err != nil {
			return FormatResponse(UnmarshalFailed), nil
		}
		rules, err := s.Core.AddPolicies(ctx, cmd.GetNamespace(), p.GetSec(), p.GetPType(), command.ToStringArray(p.GetRules()))
		if err != nil {
			return FormatResponse(err), nil
		}
		return &command.Response{EffectedRules: command.NewStringArray(rules)}, nil

	case command.Type_COMMAND_TYPE_UPDATE_POLICIES:
		var p command.UpdatePoliciesPayload
		if err = proto.Unmarshal(cmd.Payload, &p); err != nil {
			return FormatResponse(UnmarshalFailed), nil
		}
		effected, err := s.Core.UpdatePolicies(ctx, cmd.GetNamespace(), p.GetSec(), p.GetPType(), command.ToStringArray(p.GetNewRules()), command.ToStringArray(p.GetOldRules()))
		if err != nil {
			return FormatResponse(err), nil
		}
		return &command.Response{Effected: effected}, nil

	case command.Type_COMMAND_TYPE_REMOVE_POLICIES:
		var p command.RemovePoliciesPayload
		if err = proto.Unmarshal(cmd.Payload, &p); err != nil {
			return FormatResponse(UnmarshalFailed), nil
		}
		rules, err := s.Core.RemovePolicies(ctx, cmd.GetNamespace(), p.GetSec(), p.GetPType(), command.ToStringArray(p.GetRules()))
		if err != nil {
			return FormatResponse(err), nil
		}
		return &command.Response{EffectedRules: command.NewStringArray(rules)}, nil

	case command.Type_COMMAND_TYPE_REMOVE_FILTERED_POLICY:
		var p command.RemoveFilteredPolicyPayload
		if err = proto.Unmarshal(cmd.Payload, &p); err != nil {
			return FormatResponse(UnmarshalFailed), nil
		}
		rules, err := s.Core.RemoveFilteredPolicy(ctx, cmd.GetNamespace(), p.GetSec(), p.GetPType(), p.GetFieldIndex(), p.GetFieldValues())
		if err != nil {
			return FormatResponse(err), nil
		}
		return &command.Response{EffectedRules: command.NewStringArray(rules)}, nil

	case command.Type_COMMAND_TYPE_CLEAR_POLICY:
		err = s.Core.Remove(ctx, cmd.GetNamespace())
		return FormatResponse(err), nil
	}
	return nil, nil
}

func (s grpcServer) ListNamespaces(ctx context.Context, req *command.ListNamespacesRequest) (*command.ListNamespacesResponse, error) {
	ns, err := s.Core.ListNamespaces(ctx)
	if err != nil {
		return nil, err
	}
	return &command.ListNamespacesResponse{Namespace: ns}, nil
}

func (s grpcServer) PrintModel(ctx context.Context, req *command.PrintModelRequest) (*command.PrintModelResponse, error) {
	model, err := s.Core.PrintModel(ctx, req.Namespace)
	if err != nil {
		return nil, err
	}
	return &command.PrintModelResponse{Model: model}, nil
}

func (s grpcServer) ListPolicies(ctx context.Context, req *command.ListPoliciesRequest) (*command.ListPoliciesResponse, error) {
	policies, err := s.Core.ListPolicies(ctx, req.Namespace, "", 0, 1000, false)
	if err != nil {
		return nil, err
	}
	return &command.ListPoliciesResponse{Policies: command.NewStringArray(policies)}, nil

}

func (s grpcServer) ShowStats(ctx context.Context, req *command.StatsRequest) (*command.StatsResponse, error) {

	stats, err := s.Stats(ctx)
	if err != nil {
		return nil, err
	}
	buf, err := json.Marshal(stats)
	if err != nil {
		return nil, err
	}
	return &command.StatsResponse{Payload: buf}, nil
}

func newServer(core Core) command.CasbinMeshServer {
	return &grpcServer{core, command.UnimplementedCasbinMeshServer{}}
}

func NewGrpcService(core Core) *grpc.Server {
	var interceptors []grpc.UnaryServerInterceptor
	switch core.AuthType() {
	case auth.Basic:
		interceptors = append(interceptors, grpc2.BasicAuthor(core.Check))
	}
	srv := grpc.NewServer(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(interceptors...)))
	command.RegisterCasbinMeshServer(srv, newServer(core))
	return srv
}
