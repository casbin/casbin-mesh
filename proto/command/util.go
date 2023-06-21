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

package command

import "encoding/json"

func NewStringArray(input [][]string) []*StringArray {
	var out []*StringArray
	for _, s := range input {
		out = append(out, &StringArray{
			S: s,
		})
	}
	return out
}

func ToStringArray(input []*StringArray) [][]string {
	var out [][]string
	for _, i := range input {
		out = append(out, i.GetS())
	}
	return out
}

func ToInterfaces(input [][]byte) []interface{} {
	var params []interface{}
	var err error
	for _, b := range input {
		var tmp interface{}
		err = json.Unmarshal(b, &tmp)
		if err != nil {
			return nil
		}
		params = append(params, tmp)
	}
	return params
}
