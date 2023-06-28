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

package server

type Raft struct {
	ID               string               `yaml:"id"`
	Address          string               `yaml:"address"`
	AdvertiseAddress string               `yaml:"advertiseAddress"`
	TLS              TLSConfig            `yaml:"tls"`
	Authentication   AuthenticationConfig `yaml:"authentication"`
	Client           ClientConfig         `yaml:"client"`
}

type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CAFile   string `yaml:"caFile"`
	CertFile string `yaml:"certFile"`
	KeyFile  string `yaml:"keyFile"`
	AuthType string `yaml:"authType"`
}

type AuthenticationConfig struct {
	Enabled   bool                   `yaml:"enabled"`
	Providers map[string]interface{} `yaml:"providers"`
}

type ClientConfig struct {
	TLS            TLSConfig      `yaml:"tls"`
	Authentication AuthDefinition `yaml:"authentication"`
}

type AuthDefinition struct {
	AuthType  string `yaml:"authType"`
	Parameter string `yaml:"parameter"`
}

type API struct {
	Address          string               `yaml:"address"`
	AdvertiseAddress string               `yaml:"advertiseAddress"`
	Protocols        Protocols            `yaml:"protocols"`
	TLS              TLSConfig            `yaml:"tls"`
	Authentication   AuthenticationConfig `yaml:"authentication"`
	Client           ClientConfig         `yaml:"client"`
}

type Protocols struct {
	HTTP bool `yaml:"http"`
	GRPC bool `yaml:"grpc"`
}

type Config struct {
	Raft  Raft  `yaml:"raft"`
	API   API   `yaml:"api"`
	Pprof Pprof `yaml:"pprof"`
}

type Pprof struct {
	Enabled bool   `yaml:"enabled"`
	Address string `yaml:"address"`
}
