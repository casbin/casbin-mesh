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

package main

import (
	"github.com/c-bata/go-prompt"
)

//func completer(d prompt.Document) []prompt.Suggest {
//	s := []prompt.Suggest{
//		{Text: "show", Description: "Store the username and age"},
//		{Text: "create", Description: "Store the article text posted by user"},
//		{Text: "delete", Description: "Store the text commented to articles"},
//		{Text: "exit"},
//	}
//	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursorWithSpace(), true)
//}
//
//func action(ctx *cli.Context) error {
//	addr := ctx.String("Addr")
//	c := NewCtx(NewClient(options{}))
//	p := prompt.New(c.Executor, c.Completer)
//	p.Run()
//	return nil
//}

func main() {
	c := NewClient(options{
		target:   "localhost:4002",
		authType: Basic,
		username: "root",
		password: "root",
	})
	ctx := NewCtx(c)
	p := prompt.New(
		ctx.Executor,
		ctx.Completer,
		prompt.OptionCompletionOnDown(),
		prompt.OptionPrefix("127.0.0.1:4002 (Primary) >> "),
	)
	p.Run()
}
