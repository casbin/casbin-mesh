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

package main

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"os"
	"strings"
)

// CasbinRule represents a Casbin rule line.
type CasbinRule struct {
	Key   string `json:"key"`
	PType string `json:"p_type"`
	V0    string `json:"v0"`
	V1    string `json:"v1"`
	V2    string `json:"v2"`
	V3    string `json:"v3"`
	V4    string `json:"v4"`
	V5    string `json:"v5"`
}

func convertRule(ptype string, line []string) CasbinRule {
	rule := CasbinRule{PType: ptype}

	keySlice := []string{ptype}

	l := len(line)
	if l > 0 {
		rule.V0 = line[0]
		keySlice = append(keySlice, line[0])
	}
	if l > 1 {
		rule.V1 = line[1]
		keySlice = append(keySlice, line[1])
	}
	if l > 2 {
		rule.V2 = line[2]
		keySlice = append(keySlice, line[2])
	}
	if l > 3 {
		rule.V3 = line[3]
		keySlice = append(keySlice, line[3])
	}
	if l > 4 {
		rule.V4 = line[4]
		keySlice = append(keySlice, line[4])
	}
	if l > 5 {
		rule.V5 = line[5]
		keySlice = append(keySlice, line[5])
	}

	rule.Key = strings.Join(keySlice, "::")

	return rule
}

func (rule CasbinRule) ToStringArray() (out []string) {
	if rule.V0 != "" {
		out = append(out, rule.V0)
	} else {
		return
	}
	if rule.V1 != "" {
		out = append(out, rule.V1)
	} else {
		return
	}
	if rule.V2 != "" {
		out = append(out, rule.V2)
	} else {
		return
	}
	if rule.V3 != "" {
		out = append(out, rule.V3)
	} else {
		return
	}
	if rule.V4 != "" {
		out = append(out, rule.V4)
	} else {
		return
	}
	if rule.V5 != "" {
		out = append(out, rule.V5)
	} else {
		return
	}

	return
}

type Rules []CasbinRule
type RuleGroups [][]CasbinRule

func (r Rules) Print() {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"INDEX", "Type", "V0", "V1", "V2", "V3", "V4", "V5"})
	for index, rule := range r {
		t.AppendRow(table.Row{index, rule.PType, rule.V0, rule.V1, rule.V2, rule.V3, rule.V4, rule.V5})
	}
	t.Render()
}

func (rule Rules) to2DString() [][]string {
	var out [][]string
	for _, r := range rule {
		out = append(out, r.ToStringArray())
	}
	return out
}

func (rule Rules) last() *CasbinRule {
	if len(rule) > 0 {
		i := len(rule) - 1
		return &rule[i]
	}
	return nil
}

func (rule RuleGroups) last() *CasbinRule {
	if len(rule) > 0 {
		i := len(rule) - 1
		g := rule[i]
		if len(g) > 0 {
			return &g[len(g)-1]
		}
	}
	return nil
}

func (rule RuleGroups) to2DString() (old [][]string, new [][]string) {
	for _, rg := range rule {
		if len(rg) > 1 {
			old = append(old, rg[0].ToStringArray())
			new = append(new, rg[1].ToStringArray())
		}
	}
	return
}
