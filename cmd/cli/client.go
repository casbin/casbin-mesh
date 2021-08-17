/*
@Author: Weny Xu
@Date: 2021/08/09 21:17
*/

package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/casbin/casbin-mesh/proto/command"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"log"
	"time"
)

type client struct {
	grpcClient command.CasbinMeshClient
}

var (
	MarshalFailed = errors.New("marshal failed")
)

func (c client) ShowStats(ctx context.Context) ([]byte, error) {
	resp, err := c.grpcClient.ShowStats(ctx, &command.StatsRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

func (c client) AddPolicies(ctx context.Context, namespace, sec, ptype string, rules [][]string) ([][]string, error) {
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

func (c client) RemovePolicies(ctx context.Context, namespace, sec, ptype string, rules [][]string) ([][]string, error) {
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

func (c client) UpdatePolicies(ctx context.Context, namespace, sec, ptype string, old, new [][]string) (bool, error) {
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

func (c client) ListNamespaces(ctx context.Context) ([]string, error) {
	resp, err := c.grpcClient.ListNamespaces(ctx, &command.ListNamespacesRequest{})
	if err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, errors.New(resp.Error)
	}
	return resp.Namespace, nil
}

func (c client) ListPolicies(ctx context.Context, namespace string) ([][]string, error) {
	resp, err := c.grpcClient.ListPolicies(ctx, &command.ListPoliciesRequest{Namespace: namespace})
	if err != nil {
		return nil, err
	}
	return command.ToStringArray(resp.Policies), nil
}

func (c client) Enforce(ctx context.Context, namespace string, level command.EnforcePayload_Level, freshness int64, params ...interface{}) (bool, error) {
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

func (c client) PrintModel(ctx context.Context, namespace string) (string, error) {
	resp, err := c.grpcClient.PrintModel(ctx, &command.PrintModelRequest{Namespace: namespace})
	if err != nil {
		return "", err
	}
	if resp.Error != "" {
		return "", errors.New(resp.Error)
	}
	return resp.Model, nil
}

func NewClient(target string) *client {
	var opts []grpc.DialOption
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	conn, err := grpc.DialContext(ctx, target, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	log.Println("login success!")
	//defer conn.Close()
	c := command.NewCasbinMeshClient(conn)
	return &client{grpcClient: c}
}
