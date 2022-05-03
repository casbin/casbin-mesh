// Copyright 2022 The casbin-neo Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hashicorp

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

func TestNewNode(t *testing.T) {
	n, err := NewNode(NodeOptions{
		NodeID:    "node1",
		Addr:      "127.0.0.1:0",
		bootstrap: true,
	})
	// wait for leader
	time.Sleep(3 * time.Second)

	assert.Nil(t, err)
	assert.Nil(t, n.MockSet("key1", "value1"))
	value, err := n.MockGet("key1")
	assert.Nil(t, err)
	assert.Equal(t, "value1", value)
}

func TestMultiNodes(t *testing.T) {
	n, err := NewNode(NodeOptions{
		NodeID:    "node0",
		Addr:      "127.0.0.1:22999",
		bootstrap: true,
	})
	// wait for leader
	time.Sleep(3 * time.Second)

	n1, err := NewNode(NodeOptions{
		NodeID:    "node1",
		Addr:      "127.0.0.1:23999",
		bootstrap: false,
	})
	assert.Nil(t, err)
	n2, err := NewNode(NodeOptions{
		NodeID:    "node2",
		Addr:      "127.0.0.1:24999",
		bootstrap: false,
	})
	assert.Nil(t, err)

	assert.Nil(t, n.Join("node1", "127.0.0.1:23999"))
	assert.Nil(t, n.Join("node2", "127.0.0.1:24999"))

	log.Println("sleeping")
	// waiting for sync logs
	time.Sleep(1 * time.Second)
	log.Println("wakeup")

	now := time.Now()
	err = n.MockSet("key1", "value1")
	assert.Nil(t, err)
	elapsed := time.Since(now)
	fmt.Printf("it takes %v to append two entries to followers\n", elapsed)
	//log.Println("start leader read")
	//value, err := n.MockLeaderGet("key1")
	//log.Println("end leader read")
	//assert.Nil(t, err)
	//assert.Equal(t, "value1", value)

	value, err := n1.MockGet("key1")
	assert.Nil(t, err)
	assert.Equal(t, "value1", value)

	value, err = n2.MockGet("key1")
	assert.Nil(t, err)
	assert.Equal(t, "value1", value)
}
