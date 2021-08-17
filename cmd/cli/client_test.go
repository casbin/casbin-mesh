/*
@Author: Weny Xu
@Date: 2021/08/14 14:25
*/

package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

var (
	c *client
)

func TestMain(m *testing.M) {
	c = NewClient("127.0.0.1:4002")
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
