// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/casbin/casbin-mesh/pkg/auth"
	"github.com/casbin/casbin-mesh/pkg/handler/http"
	"github.com/go-playground/validator"
	"golang.org/x/net/context"
	"io"
	"io/ioutil"
	http2 "net/http"
)

type httpService struct {
	http.Server
	Core
	*validator.Validate
}

type Middleware func(handlerFunc http.HandlerFunc) http.HandlerFunc

func chain(outer Middleware, others ...Middleware) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		for i := len(others) - 1; i >= 0; i-- {
			next = others[i](next)
		}
		return outer(next)
	}
}

func NewHttpService(core Core) *httpService {
	httpS := http.New()
	validate := validator.New()
	srv := httpService{httpS, core, validate}
	// set response header
	httpS.Use(setResponseHeader)

	// enable global middleware
	switch core.AuthType() {
	case auth.Basic:
		httpS.Use(http.BasicAuthor(core.Check))
	}

	httpS.Handle("/join", srv.handleJoin)
	httpS.Handle("/remove", srv.handleRemove)

	// write
	httpS.Handle("/create/namespace", chain(srv.autoForwardToLeader)(srv.handleCreateNameSpace))
	httpS.Handle("/list/namespaces", chain(srv.autoForwardToLeader)(srv.handleListNamespace))
	httpS.Handle("/print/model", chain(srv.autoForwardToLeader)(srv.handlePrintModel))
	httpS.Handle("/list/policies", chain(srv.autoForwardToLeader)(srv.handleListPolicies))
	httpS.Handle("/set/model", chain(srv.autoForwardToLeader)(srv.handleSetModelFromString))
	httpS.Handle("/add/policies", chain(srv.autoForwardToLeader)(srv.handleAddPolicies))
	httpS.Handle("/remove/policies", chain(srv.autoForwardToLeader)(srv.handleRemovePolicies))
	httpS.Handle("/remove/filtered_policies", chain(srv.autoForwardToLeader)(srv.handleRemoveFilteredPolicy))
	httpS.Handle("/update/policies", chain(srv.autoForwardToLeader)(srv.handleUpdatePolicies))
	httpS.Handle("/clear/policy", chain(srv.autoForwardToLeader)(srv.handleClearPolicy))

	// read
	httpS.Handle("/enforce", srv.handleEnforce)
	httpS.Handle("/stats", srv.handleStats)
	return &srv
}

type JoinRequest struct {
	ID       string            `json:"id" validate:"required"`
	Addr     string            `json:"addr" validate:"required"`
	Voter    bool              `json:"voter" validate:"required"`
	Metadata map[string]string `json:"metadata"`
}

func setResponseHeader(ctx *http.Context) error {
	ctx.ResponseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	return nil
}

func (s *httpService) autoForwardToLeader(fn http.HandlerFunc) http.HandlerFunc {
	return func(c *http.Context) error {
		if s.IsLeader(context.TODO()) {
			return fn(c)
		} else {
			schema := "http"
			if c.Request.TLS != nil {
				schema = "https"
			}
			c.ResponseWriter.Header().Del("Vary")
			c.ResponseWriter.Header().Del("Access-Control-Allow-Origin")

			body, err := ioutil.ReadAll(c.Request.Body)
			if err != nil {
				http2.Error(c.ResponseWriter, err.Error(), http2.StatusInternalServerError)
			}
			url := fmt.Sprintf("%s://%s%s", schema, s.LeaderAddr(), c.Request.RequestURI)
			proxyReq, err := http2.NewRequest(c.Request.Method, url, bytes.NewReader(body))

			// clone the header
			proxyReq.Header = make(http2.Header)
			for h, val := range c.Request.Header {
				proxyReq.Header[h] = val
			}

			// forward the incoming request to leader
			resp, err := http2.DefaultClient.Do(proxyReq)
			if err != nil {
				http2.Error(c.ResponseWriter, err.Error(), http2.StatusBadGateway)
				return nil
			}
			// copy the response
			_, err = io.Copy(c.ResponseWriter, resp.Body)
			if err != nil {
				fmt.Printf("Copy failed:%s", err)
				return err
			}
			defer resp.Body.Close()
		}
		return nil
	}
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
	return ctx.StatusCode(http2.StatusOK).JSON(EnforceReply{Ok: output})
}

type AddPoliciesRequest struct {
	NS    string     `json:"ns" validate:"required"`
	Sec   string     `json:"sec" validate:"required"`
	PType string     `json:"ptype" validate:"required"`
	Rules [][]string `json:"rules" validate:"required"`
}

type Response struct {
	Effected      bool       `json:"effected,omitempty"`
	EffectedRules [][]string `json:"effected_rules,omitempty"`
}

func (s *httpService) handleAddPolicies(ctx *http.Context) (err error) {
	var request AddPoliciesRequest
	if err = s.decode(ctx.Request.Body, &request); err != nil {
		return
	}
	var rules [][]string
	if rules, err = s.AddPolicies(context.TODO(), request.NS, request.Sec, request.PType, request.Rules); err != nil {
		return err
	}
	return ctx.StatusCode(http2.StatusOK).JSON(Response{EffectedRules: rules})
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
	var rules [][]string
	if rules, err = s.RemovePolicies(context.TODO(), request.NS, request.Sec, request.PType, request.Rules); err != nil {
		return
	}

	return ctx.StatusCode(http2.StatusOK).JSON(Response{EffectedRules: rules})
}

type RemoveFilteredPolicyRequest struct {
	NS          string   `json:"ns" validate:"required"`
	Sec         string   `json:"sec" validate:"required"`
	PType       string   `json:"ptype" validate:"required"`
	FieldIndex  *int32   `json:"fieldIndex" validate:"required"`
	FieldValues []string `json:"fieldValues" validate:"required"`
}

func (s *httpService) handleRemoveFilteredPolicy(ctx *http.Context) (err error) {
	var request RemoveFilteredPolicyRequest
	if err = s.decode(ctx.Request.Body, &request); err != nil {
		return
	}
	var rules [][]string
	if rules, err = s.RemoveFilteredPolicy(context.TODO(), request.NS, request.Sec, request.PType, *request.FieldIndex, request.FieldValues); err != nil {
		return
	}
	return ctx.StatusCode(http2.StatusOK).JSON(Response{EffectedRules: rules})
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
	var effected bool
	if effected, err = s.UpdatePolicies(context.TODO(), request.NS, request.Sec, request.PType, request.NewRules, request.OldRules); err != nil {
		return
	}
	return ctx.StatusCode(http2.StatusOK).JSON(Response{Effected: effected})
}

type ClearPolicyRequest struct {
	NS string `json:"ns" validate:"required"`
}

func (s *httpService) handleClearPolicy(ctx *http.Context) (err error) {
	var request ClearPolicyRequest
	if err = s.decode(ctx.Request.Body, &request); err != nil {
		return
	}
	if err = s.ClearPolicy(context.TODO(), request.NS); err != nil {
		return
	}
	ctx.StatusCode(http2.StatusOK)
	return
}

type ListPoliciesRequest struct {
	NS      string `json:"ns" validate:"required"`
	Cursor  string `json:"cursor"`
	Skip    int64  `json:"skip"`
	Limit   int64  `json:"limit"`
	Reverse bool   `json:"reverse"`
}

func (s *httpService) handleListPolicies(ctx *http.Context) error {
	var request ListPoliciesRequest
	if err := s.decode(ctx.Request.Body, &request); err != nil {
		return err
	}
	out, err := s.ListPolicies(context.TODO(), request.NS, request.Cursor, request.Skip, request.Limit, request.Reverse)
	if err != nil {
		return err
	}
	return ctx.StatusCode(http2.StatusOK).JSON(out)
}

type PrintModelRequest struct {
	NS string `json:"ns" validate:"required"`
}

func (s *httpService) handlePrintModel(ctx *http.Context) error {
	var request PrintModelRequest
	if err := s.decode(ctx.Request.Body, &request); err != nil {
		return err
	}
	out, err := s.PrintModel(context.TODO(), request.NS)
	if err != nil {
		return err
	}
	return ctx.StatusCode(http2.StatusOK).JSON(out)
}

func (s *httpService) handleListNamespace(ctx *http.Context) error {
	out, err := s.ListNamespaces(context.TODO())
	if err != nil {
		return err
	}
	return ctx.StatusCode(http2.StatusOK).JSON(out)
}

func (s *httpService) handleStats(ctx *http.Context) error {
	out, err := s.Stats(context.TODO())
	if err != nil {
		return err
	}
	return ctx.StatusCode(http2.StatusOK).JSON(out)
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
