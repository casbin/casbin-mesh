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
	"fmt"
	"os"

	"github.com/erikgeiser/promptkit/selection"
	"github.com/erikgeiser/promptkit/textinput"
)

const customTemplate = `
	{{- "┏" }}━{{ Repeat "━" (Len .Prompt) }}━┯━{{ Repeat "━" 13 }}{{ "━━━━┓\n" }}
	{{- "┃" }} {{ Bold .Prompt }} │ {{ .Input -}}
	{{- Repeat " " (Max 0 (Sub 16 (Len .Input))) }}
	{{- if not .Valid -}}
		{{- Foreground "1" (Bold "✘") -}}
	{{- else -}}
		{{- Foreground "2" (Bold "✔") -}}
	{{- end -}}┃
	{{- "\n┗" }}━{{ Repeat "━" (Len .Prompt) }}━┷━{{ Repeat "━" 13 }}{{ "━━━━┛" -}}
	`

const customResultTemplate = `
	{{- Bold (print "🖥️  Connecting to " (Foreground "32" .FinalValue) "\n") -}}
	`

func GetHost() string {
	input := textinput.New("Enter an Host address")
	input.Placeholder = "127.0.0.1:4002"
	input.Template = customTemplate
	input.ResultTemplate = customResultTemplate
	input.CharLimit = 15

	ip, err := input.RunPrompt()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	return ip
}

func GetAuthType() string {
	sp := selection.New("Auth Type:",
		selection.Choices([]string{"Noop", "Basic"}))
	sp.PageSize = 3

	choice, err := sp.RunPrompt()
	if err != nil {
		fmt.Printf("Error: %v\n", err)

		os.Exit(1)
	}

	return choice.String
}

func GetUsername() string {
	input := textinput.New("Username:")

	username, err := input.RunPrompt()
	if err != nil {
		fmt.Printf("Error: %v\n", err)

		os.Exit(1)
	}
	return username
}

func GetPWD() string {
	input := textinput.New("Password:")
	input.Hidden = true

	pwd, err := input.RunPrompt()
	if err != nil {
		fmt.Printf("Error: %v\n", err)

		os.Exit(1)
	}
	return pwd
}
