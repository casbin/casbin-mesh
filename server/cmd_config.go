// Copyright 2022 The Casbin Mesh Authors.
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

type CmdConfig struct {
	EnableAuth                 bool
	RootUsername               string
	RootPassword               string
	RaftAddr                   string
	RaftAdv                    string
	JoinSrcIP                  string
	X509CACert                 string
	X509Cert                   string
	X509Key                    string
	NodeID                     string
	JoinAddr                   string
	JoinAttempts               int
	JoinInterval               string
	NoVerify                   bool
	PprofEnabled               bool
	RaftLogLevel               string
	RaftNonVoter               bool
	RaftSnapThreshold          uint64
	RaftSnapInterval           string
	RaftLeaderLeaseTimeout     string
	RaftHeartbeatTimeout       string
	RaftElectionTimeout        string
	RaftApplyTimeout           string
	RaftOpenTimeout            string
	RaftWaitForLeader          bool
	RaftShutdownOnRemove       bool
	CompressionSize            int
	CompressionBatch           int
	ShowVersion                bool
	CpuProfile                 string
	MemProfile                 string
	Encrypt                    bool
	DataPath                   string
	ConfigPath                 string
	ServerHTTPAddress          string
	ServerHTTPAdvertiseAddress string
	ServerGRPCAddress          string
	ServerGPRCAdvertiseAddress string
	RaftCAFile                 string
	RaftKeyFile                string
	RaftCertFile               string
	RaftTLSEnabled             bool
	ServerCAFile               string
	ServerKeyFile              string
	ServerCertFile             string
	ServerTLSEnabled           bool
	PprofAddress               string
}

func (c *CmdConfig) getRaftCAFile() string {
	if c.RaftCAFile != "" {
		return c.RaftCAFile
	}
	return c.X509CACert
}

func (c *CmdConfig) getRaftKeyFile() string {
	if c.RaftKeyFile != "" {
		return c.RaftKeyFile
	}
	return c.X509Key
}

func (c *CmdConfig) getRaftCertFile() string {
	if c.RaftCertFile != "" {
		return c.RaftCertFile
	}
	return c.X509Cert
}

func (c *CmdConfig) isRaftTlsEnabled() bool {
	return c.RaftTLSEnabled || c.Encrypt
}

func (c *CmdConfig) getServerCAFile() string {
	if c.ServerCAFile != "" {
		return c.ServerCAFile
	}
	return c.X509CACert
}

func (c *CmdConfig) getServerKeyFile() string {
	if c.ServerKeyFile != "" {
		return c.ServerKeyFile
	}
	return c.X509Key
}

func (c *CmdConfig) getServerCertFile() string {
	if c.ServerCertFile != "" {
		return c.ServerCertFile
	}
	return c.X509Cert
}

func (c *CmdConfig) isServerTlsEnabled() bool {
	return c.ServerTLSEnabled || c.Encrypt
}
