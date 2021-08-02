package core

import (
	"context"
	"errors"

	"github.com/casbin/casbin-mesh/proto/command"
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
		err = s.Core.AddPolicies(ctx, cmd.GetNamespace(), p.GetSec(), p.GetPType(), command.ToStringArray(p.GetRules()))
		return FormatResponse(err), nil

	case command.Type_COMMAND_TYPE_UPDATE_POLICIES:
		var p command.UpdatePoliciesPayload
		if err = proto.Unmarshal(cmd.Payload, &p); err != nil {
			return FormatResponse(UnmarshalFailed), nil
		}
		err = s.Core.UpdatePolicies(ctx, cmd.GetNamespace(), p.GetSec(), p.GetPType(), command.ToStringArray(p.GetNewRules()), command.ToStringArray(p.GetOldRules()))
		return FormatResponse(err), nil

	case command.Type_COMMAND_TYPE_REMOVE_POLICIES:
		var p command.RemovePoliciesPayload
		if err = proto.Unmarshal(cmd.Payload, &p); err != nil {
			return FormatResponse(UnmarshalFailed), nil
		}
		err = s.Core.RemovePolicies(ctx, cmd.GetNamespace(), p.GetSec(), p.GetPType(), command.ToStringArray(p.GetRules()))
		return FormatResponse(err), nil

	case command.Type_COMMAND_TYPE_REMOVE_FILTERED_POLICY:
		var p command.RemoveFilteredPolicyPayload
		if err = proto.Unmarshal(cmd.Payload, &p); err != nil {
			return FormatResponse(UnmarshalFailed), nil
		}
		err = s.Core.RemoveFilteredPolicy(ctx, cmd.GetNamespace(), p.GetSec(), p.GetPType(), p.GetFieldIndex(), p.GetFieldValues())
		return FormatResponse(err), nil

	case command.Type_COMMAND_TYPE_CLEAR_POLICY:
		err = s.Core.Remove(ctx, cmd.GetNamespace())
		return FormatResponse(err), nil
	}
	return nil, nil
}

func newServer(core Core) command.CasbinMeshServer {
	return &grpcServer{core, command.UnimplementedCasbinMeshServer{}}
}
func NewGrpcService(core Core) *grpc.Server {
	srv := grpc.NewServer()
	command.RegisterCasbinMeshServer(srv, newServer(core))
	return srv
}
