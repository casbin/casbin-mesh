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

package client

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

var (
	c *Client
)

func TestMain(m *testing.M) {
	//c = NewClient("127.0.0.1:4002")
	os.Exit(m.Run())
}

func TestClient_Enforce(t *testing.T) {
	result, err := c.Enforce(context.TODO(), "default", 0, int64(time.Millisecond*200), "leo", "data", "read")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result)
}

func BenchmarkClient_Enforce_Level_0(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = c.Enforce(context.TODO(), "default", 0, int64(time.Millisecond*200), "leo", "data", "read")
		}
	})
}

func BenchmarkClient_Enforce_Level_1(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = c.Enforce(context.TODO(), "default", 1, 0, "leo", "data", "read")
		}
	})
}

func BenchmarkClient_Enforce_Level_2(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			result, err := c.Enforce(context.TODO(), "default", 2, int64(time.Millisecond*200), "leo", "data", "read")
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(result)
		}
	})
}
