/*
Copyright 2020 RS4
@Author: Weny Xu
@Date: 2021/03/28 20:34
*/

package service

import (
	"context"
	"encoding/json"
	"io"
	http2 "net/http"

	"github.com/WenyXu/casbind/pkg/transport/http"
	"github.com/go-playground/validator"
)

type httpService struct {
	http.Server
	Service
	*validator.Validate
}

func NewHttpService(core Service) *httpService {
	s := http.New()
	validate := validator.New()
	ser := httpService{s, core, validate}
	s.Handle("/join", ser.handleJoin)
	s.Handle("/remove", ser.handleRemove)
	s.Handle("/create/namespace", ser.handleCreateNameSpace)
	s.Handle("/set/model", ser.handleSetModelFromString)
	s.Handle("/enforce", ser.handleEnforce)
	s.Handle("/add/policies", ser.handleAddPolicies)
	s.Handle("/remove/policies", ser.handleRemovePolicies)
	s.Handle("/remove/filtered_policies", ser.handleRemoveFilteredPolicy)
	s.Handle("/update/policies", ser.handleUpdatePolicies)
	s.Handle("/update/policy", ser.handleUpdatePolicy)
	s.Handle("/clear/policy", ser.handleClearPolicy)
	s.Handle("/stats", ser.handleStats)
	return &ser
}

type JoinRequest struct {
	ID       string            `json:"id" validate:"required"`
	Addr     string            `json:"addr" validate:"required"`
	Voter    bool              `json:"voter" validate:"required"`
	Metadata map[string]string `json:"metadata"`
}

func (s *httpService) handleJoin(ctx *http.Context) (err error) {
	var request JoinRequest
	if err = s.decode(ctx.Request.Body, &request); err != nil {
		return
	}
	if err = s.Join(context.TODO(), request.ID, request.Addr, request.Voter, request.Metadata); err != nil {
		return
	}
	ctx.StatusCode(http2.StatusOK)
	return nil
}

type RemoveRequest struct {
	ID string `json:"id" validate:"required"`
}

func (s *httpService) handleRemove(ctx *http.Context) (err error) {
	var request RemoveRequest
	if err = s.decode(ctx.Request.Body, &request); err != nil {
		return
	}
	if err = s.Remove(context.TODO(), request.ID); err != nil {
		return
	}
	ctx.StatusCode(http2.StatusOK)
	return nil
}

type CreateNameSpaceRequest struct {
	NS string `json:"ns" validate:"required"`
}

func (s *httpService) handleCreateNameSpace(ctx *http.Context) (err error) {
	var request CreateNameSpaceRequest
	if err = s.decode(ctx.Request.Body, &request); err != nil {
		return
	}
	if err = s.CreateNamespace(context.TODO(), request.NS); err != nil {
		return
	}
	ctx.StatusCode(http2.StatusOK)
	return nil
}

type SetModelFromStringRequest struct {
	NS   string `json:"ns" validate:"required"`
	Text string `json:"text" validate:"required"`
}

func (s *httpService) handleSetModelFromString(ctx *http.Context) (err error) {
	var request SetModelFromStringRequest
	if err = s.decode(ctx.Request.Body, &request); err != nil {
		return
	}
	if err = s.SetModelFromString(context.TODO(), request.NS, request.Text); err != nil {
		return
	}
	ctx.StatusCode(http2.StatusOK)
	return nil
}

type EnforceRequest struct {
	NS        string        `json:"ns" validate:"required"`
	Level     int32         `json:"level"`
	Freshness int64         `json:"freshness"`
	Params    []interface{} `json:"params"`
}

type EnforceReply struct {
	Ok bool `json:"ok"`
}

func (s *httpService) handleEnforce(ctx *http.Context) (err error) {
	var request EnforceRequest
	var output bool
	if err = s.decode(ctx.Request.Body, &request); err != nil {
		return
	}
	if output, err = s.Enforce(context.TODO(), request.NS, request.Level, request.Freshness, request.Params...); err != nil {
		return
	}
	return ctx.StatusCode(http2.StatusOK).Write(EnforceReply{Ok: output})
}

type AddPoliciesRequest struct {
	NS    string     `json:"ns" validate:"required"`
	Sec   string     `json:"sec" validate:"required"`
	PType string     `json:"ptype" validate:"required"`
	Rules [][]string `json:"rules" validate:"required"`
}

func (s *httpService) handleAddPolicies(ctx *http.Context) (err error) {
	var request AddPoliciesRequest
	if err = s.decode(ctx.Request.Body, &request); err != nil {
		return
	}
	if err = s.AddPolicies(context.TODO(), request.NS, request.Sec, request.PType, request.Rules); err != nil {
		return
	}
	ctx.StatusCode(http2.StatusOK)
	return
}

type RemovePoliciesRequest struct {
	NS    string     `json:"ns" validate:"required"`
	Sec   string     `json:"sec" validate:"required"`
	PType string     `json:"ptype" validate:"required"`
	Rules [][]string `json:"rules" validate:"required"`
}

func (s *httpService) handleRemovePolicies(ctx *http.Context) (err error) {
	var request RemovePoliciesRequest
	if err = s.decode(ctx.Request.Body, &request); err != nil {
		return
	}
	if err = s.RemovePolicies(context.TODO(), request.NS, request.Sec, request.PType, request.Rules); err != nil {
		return
	}
	ctx.StatusCode(http2.StatusOK)
	return
}

type RemoveFilteredPolicyRequest struct {
	NS          string   `json:"ns" validate:"required"`
	Sec         string   `json:"sec" validate:"required"`
	PType       string   `json:"ptype" validate:"required"`
	FieldIndex  int32    `json:"fieldIndex" validate:"required"`
	FieldValues []string `json:"fieldValues" validate:"required"`
}

func (s *httpService) handleRemoveFilteredPolicy(ctx *http.Context) (err error) {
	var request RemoveFilteredPolicyRequest
	if err = s.decode(ctx.Request.Body, &request); err != nil {
		return
	}
	if err = s.RemoveFilteredPolicy(context.TODO(), request.NS, request.Sec, request.PType, request.FieldIndex, request.FieldValues); err != nil {
		return
	}
	ctx.StatusCode(http2.StatusOK)
	return
}

type UpdatePolicyRequest struct {
	NS      string   `json:"ns" validate:"required"`
	Sec     string   `json:"sec" validate:"required"`
	PType   string   `json:"ptype" validate:"required"`
	NewRule []string `json:"newRule" validate:"required"`
	OldRule []string `json:"oldRule" validate:"required"`
}

func (s *httpService) handleUpdatePolicy(ctx *http.Context) (err error) {
	var request UpdatePolicyRequest
	if err = s.decode(ctx.Request.Body, &request); err != nil {
		return
	}
	if err = s.UpdatePolicy(context.TODO(), request.NS, request.Sec, request.PType, request.NewRule, request.OldRule); err != nil {
		return
	}
	ctx.StatusCode(http2.StatusOK)
	return
}

type UpdatePoliciesRequest struct {
	NS       string     `json:"ns" validate:"required"`
	Sec      string     `json:"sec" validate:"required"`
	PType    string     `json:"ptype" validate:"required"`
	NewRules [][]string `json:"newRules" validate:"required"`
	OldRules [][]string `json:"oldRules" validate:"required"`
}

func (s *httpService) handleUpdatePolicies(ctx *http.Context) (err error) {
	var request UpdatePoliciesRequest
	if err = s.decode(ctx.Request.Body, &request); err != nil {
		return
	}
	if err = s.UpdatePolicies(context.TODO(), request.NS, request.Sec, request.PType, request.NewRules, request.OldRules); err != nil {
		return
	}
	ctx.StatusCode(http2.StatusOK)
	return
}

type ClearPolicyRequest struct {
	NS string `json:"ns" validate:"required"`
}

func (s *httpService) handleClearPolicy(ctx *http.Context) (err error) {
	var request UpdatePoliciesRequest
	if err = s.decode(ctx.Request.Body, &request); err != nil {
		return
	}
	if err = s.ClearPolicy(context.TODO(), request.NS); err != nil {
		return
	}
	ctx.StatusCode(http2.StatusOK)
	return
}

func (s *httpService) handleStats(ctx *http.Context) error {
	out, err := s.Stats(context.TODO())
	if err != nil {
		return err
	}
	return ctx.StatusCode(http2.StatusOK).Write(out)
}

func (s *httpService) decode(reader io.ReadCloser, output interface{}) (err error) {
	if err = json.NewDecoder(reader).Decode(&output); err != nil {
		return
	}
	if err = s.Validate.Struct(output); err != nil {
		return
	}
	return nil
}
