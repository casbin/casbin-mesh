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

package client

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/casbin/casbin-mesh/proto/command"
	"github.com/golang/protobuf/proto"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"log"
	"time"
)

type Client struct {
	grpcClient command.CasbinMeshClient
}

var (
	MarshalFailed = errors.New("marshal failed")
)

func (c Client) ShowStats(ctx context.Context) ([]byte, error) {
	resp, err := c.grpcClient.ShowStats(ctx, &command.StatsRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

func (c Client) AddPolicies(ctx context.Context, namespace, sec, ptype string, rules [][]string) ([][]string, error) {
	payload := command.AddPoliciesPayload{
		Sec:   sec,
		PType: ptype,
		Rules: command.NewStringArray(rules),
	}
	p, err := proto.Marshal(&payload)
	if err != nil {
		return nil, MarshalFailed
	}
	cmd := command.Command{
		Type:      command.Type_COMMAND_TYPE_ADD_POLICIES,
		Namespace: namespace,
		Payload:   p,
	}
	resp, err := c.grpcClient.Request(ctx, &cmd)
	if err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, errors.New(resp.Error)
	}
	return command.ToStringArray(resp.EffectedRules), nil
}

func (c Client) RemovePolicies(ctx context.Context, namespace, sec, ptype string, rules [][]string) ([][]string, error) {
	payload := command.RemovePoliciesPayload{
		Sec:   sec,
		PType: ptype,
		Rules: command.NewStringArray(rules),
	}
	p, err := proto.Marshal(&payload)
	if err != nil {
		return nil, MarshalFailed
	}
	cmd := command.Command{
		Type:      command.Type_COMMAND_TYPE_REMOVE_POLICIES,
		Namespace: namespace,
		Payload:   p,
	}
	resp, err := c.grpcClient.Request(ctx, &cmd)
	if err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, errors.New(resp.Error)
	}
	return command.ToStringArray(resp.EffectedRules), nil
}

func (c Client) UpdatePolicies(ctx context.Context, namespace, sec, ptype string, old, new [][]string) (bool, error) {
	log.Printf("sec:%s,ptype:%s,or%v,nr:%v", sec, ptype, old, new)
	payload := command.UpdatePoliciesPayload{
		Sec:      sec,
		PType:    ptype,
		OldRules: command.NewStringArray(old),
		NewRules: command.NewStringArray(new),
	}
	p, err := proto.Marshal(&payload)
	if err != nil {
		return false, MarshalFailed
	}
	cmd := command.Command{
		Type:      command.Type_COMMAND_TYPE_UPDATE_POLICIES,
		Namespace: namespace,
		Payload:   p,
	}
	resp, err := c.grpcClient.Request(ctx, &cmd)
	if err != nil {
		return false, err
	}
	if resp.Error != "" {
		return false, errors.New(resp.Error)
	}
	return resp.Effected, nil
}

func (c Client) ListNamespaces(ctx context.Context) ([]string, error) {
	resp, err := c.grpcClient.ListNamespaces(ctx, &command.ListNamespacesRequest{})
	if err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, errors.New(resp.Error)
	}
	return resp.Namespace, nil
}

func (c Client) ListPolicies(ctx context.Context, namespace string) ([][]string, error) {
	resp, err := c.grpcClient.ListPolicies(ctx, &command.ListPoliciesRequest{Namespace: namespace})
	if err != nil {
		return nil, err
	}
	return command.ToStringArray(resp.Policies), nil
}

func (c Client) Enforce(ctx context.Context, namespace string, level command.EnforcePayload_Level, freshness int64, params ...interface{}) (bool, error) {
	var B [][]byte
	for _, p := range params {
		b, err := json.Marshal(p)
		if err != nil {
			return false, err
		}
		B = append(B, b)
	}

	payload := &command.EnforcePayload{
		B:         B,
		Level:     level,
		Freshness: freshness,
	}
	cmd := &command.EnforceRequest{
		Namespace: namespace,
		Payload:   payload,
	}
	result, err := c.grpcClient.Enforce(ctx, cmd)
	if err != nil {
		return false, err
	}
	if result.Error != "" {
		return false, errors.New(result.Error)
	}
	return result.Ok, nil
}

func (c Client) PrintModel(ctx context.Context, namespace string) (string, error) {
	resp, err := c.grpcClient.PrintModel(ctx, &command.PrintModelRequest{Namespace: namespace})
	if err != nil {
		return "", err
	}
	if resp.Error != "" {
		return "", errors.New(resp.Error)
	}
	return resp.Model, nil
}

type Options struct {
	Target   string
	AuthType AuthType
	Username string
	Password string
}

func NewClient(op Options) *Client {
	var opts []grpc.DialOption
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())

	switch op.AuthType {
	case Basic:
		opts = append(opts, grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(BasicAuthor(op.Username, op.Password))))
	}

	conn, err := grpc.DialContext(ctx, op.Target, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	log.Println("login success!")
	//defer conn.Close()
	c := command.NewCasbinMeshClient(conn)
	return &Client{grpcClient: c}
}
